package paypal

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// PayPal Orders v2 client (AUTHORIZE intent + manual capture), matching the Stripe payment flow:
// the buyer approves on PayPal, the verified webhook authorizes the order, and the admin captures.
// No card or bank data ever reaches this server — only PayPal order/authorization ids.
//
// Secrets come from config.yaml (populated by the environment). Nothing here logs credentials.
const (
	// SandboxBaseURL is the PayPal sandbox REST endpoint.
	SandboxBaseURL = "https://api-m.sandbox.paypal.com"
	// LiveBaseURL is the PayPal live REST endpoint.
	LiveBaseURL = "https://api-m.paypal.com"

	httpTimeout  = 20 * time.Second
	maxBodyBytes = 1 << 20

	// RateScale is the fixed-point scale of FXRatePPM: 1_000_000 == 1:1.
	RateScale = int64(1_000_000)
	// maxFXRatePPM bounds a configured rate at 100:1 so a config typo cannot overcharge.
	maxFXRatePPM = 100 * RateScale
)

// OrderView is the subset of a PayPal order this integration uses.
type OrderView struct {
	ID              string // PayPal order id
	Status          string // CREATED | APPROVED | COMPLETED | ...
	ApprovalURL     string // buyer approval link (approve / payer-action)
	AuthorizationID string // set once the order was authorized
	CaptureID       string // set once the authorization was captured
	AmountMinor     int64  // order amount in minor units
	Currency        string
	// ChargeAmount/ChargeCurrency are what we asked PayPal to charge; FXRatePPM is the
	// booking->charge rate snapshot in parts per million (RateScale == 1:1).
	ChargeAmount   int64
	ChargeCurrency string
	FXRatePPM      int64
}

// ChargePlan is what PayPal will charge for a booking currency: the booking currency itself when
// PayPal supports it, otherwise the configured settlement currency at the configured rate.
type ChargePlan struct {
	AmountMinor int64
	Currency    string
	FXRatePPM   int64
}

// PlanCharge resolves the charged amount and currency for a booking. The rate comes only from
// config (admin-maintained), never from the client, and the result is snapshotted on the payment
// so the webhook can verify the exact amount that was charged.
func PlanCharge(amountMinor int64, currency string) (*ChargePlan, error) {
	if amountMinor <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}
	cur := strings.ToUpper(strings.TrimSpace(currency))
	if cur == "" {
		cur = strings.ToUpper(strings.TrimSpace(global.GVA_CONFIG.Payment.PayPal.Currency))
	}
	if cur == "" {
		cur = "USD"
	}
	if SupportedCurrency(cur) {
		return &ChargePlan{AmountMinor: amountMinor, Currency: cur, FXRatePPM: RateScale}, nil
	}

	c := global.GVA_CONFIG.Payment.PayPal
	settle := strings.ToUpper(strings.TrimSpace(c.SettlementCurrency))
	if settle == "" {
		settle = "USD"
	}
	if !SupportedCurrency(settle) {
		return nil, fmt.Errorf("paypal settlement currency %s is not supported", settle)
	}
	if settle == cur {
		return nil, fmt.Errorf("paypal cannot charge %s", cur)
	}
	rate, ok := fxRate(c.FXRates, cur)
	if !ok || rate <= 0 {
		return nil, fmt.Errorf("paypal cannot charge %s: configure payment.paypal.fx-rates.%s", cur, cur)
	}
	ratePPM := int64(math.Round(rate * float64(RateScale)))
	if ratePPM <= 0 || ratePPM > maxFXRatePPM {
		return nil, fmt.Errorf("paypal fx rate for %s is out of range", cur)
	}
	// half-up integer rounding keeps the charge exact and reproducible from the snapshot
	charge := (amountMinor*ratePPM + RateScale/2) / RateScale
	if charge <= 0 {
		return nil, fmt.Errorf("paypal amount for %s rounds to zero", cur)
	}
	return &ChargePlan{AmountMinor: charge, Currency: settle, FXRatePPM: ratePPM}, nil
}

// fxRate looks a currency up case-insensitively in the configured rate map.
func fxRate(rates map[string]float64, currency string) (float64, bool) {
	if rate, ok := rates[currency]; ok {
		return rate, true
	}
	for k, v := range rates {
		if strings.EqualFold(strings.TrimSpace(k), currency) {
			return v, true
		}
	}
	return 0, false
}

var httpClient = &http.Client{Timeout: httpTimeout}

// supportedCurrencies are the fiat currencies the PayPal REST API accepts. AED is deliberately
// absent (PayPal does not support it), so an AED booking cannot be charged through PayPal and is
// rejected instead of being silently converted at a guessed rate.
var supportedCurrencies = map[string]bool{
	"AUD": true, "BRL": true, "CAD": true, "CNY": true, "CZK": true, "DKK": true, "EUR": true,
	"HKD": true, "HUF": true, "ILS": true, "JPY": true, "MYR": true, "MXN": true, "TWD": true,
	"NZD": true, "NOK": true, "PHP": true, "PLN": true, "GBP": true, "RUB": true, "SGD": true,
	"SEK": true, "CHF": true, "THB": true, "USD": true,
}

var zeroDecimalCurrencies = map[string]bool{"HUF": true, "JPY": true, "TWD": true}

// Enabled reports whether PayPal has usable API credentials.
func Enabled() bool {
	c := global.GVA_CONFIG.Payment.PayPal
	return c.Enabled &&
		strings.TrimSpace(c.ClientID) != "" &&
		strings.TrimSpace(c.ClientSecret) != ""
}

// SupportedCurrency reports whether PayPal can charge in the given currency.
func SupportedCurrency(currency string) bool {
	return supportedCurrencies[strings.ToUpper(strings.TrimSpace(currency))]
}

func baseURL() (string, error) {
	u := strings.TrimSpace(global.GVA_CONFIG.Payment.PayPal.BaseURL)
	if u == "" {
		u = SandboxBaseURL
	}
	u = strings.TrimRight(u, "/")
	// Refuse anything but TLS: credentials and payment handles must never travel in clear text.
	if !strings.HasPrefix(strings.ToLower(u), "https://") {
		return "", errors.New("paypal base url must use https")
	}
	return u, nil
}

// --- OAuth2 client-credentials token (cached until shortly before it expires) ---

var (
	tokenMu    sync.Mutex
	cachedTok  string
	tokenUntil time.Time
)

func accessToken(ctx context.Context) (string, error) {
	c := global.GVA_CONFIG.Payment.PayPal
	clientID := strings.TrimSpace(c.ClientID)
	secret := strings.TrimSpace(c.ClientSecret)
	if clientID == "" || secret == "" {
		return "", errors.New("paypal credentials are not configured")
	}

	tokenMu.Lock()
	defer tokenMu.Unlock()
	if cachedTok != "" && time.Now().Before(tokenUntil) {
		return cachedTok, nil
	}

	base, err := baseURL()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/oauth2/token",
		strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+secret)))

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("paypal token request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", apiError(resp.StatusCode, raw)
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.AccessToken == "" {
		return "", errors.New("paypal returned no access token")
	}
	ttl := time.Duration(out.ExpiresIn) * time.Second
	if ttl <= time.Minute {
		ttl = 5 * time.Minute
	}
	cachedTok = out.AccessToken
	tokenUntil = time.Now().Add(ttl - 30*time.Second)
	return cachedTok, nil
}

// --- Orders ---

// CreateOrder opens an AUTHORIZE order for a booking and returns its buyer approval link.
func CreateOrder(ctx context.Context, bookingID uint, bookingCode string, amountMinor int64, currency string) (*OrderView, error) {
	if !Enabled() {
		return nil, errors.New("paypal is not configured")
	}
	if amountMinor <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}
	plan, err := PlanCharge(amountMinor, currency)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(bookingCode)
	if code == "" {
		return nil, errors.New("booking code is required")
	}

	unit := map[string]interface{}{
		"reference_id": code,
		"custom_id":    strconv.FormatUint(uint64(bookingID), 10),
		"description":  "Booking " + code,
		"amount": map[string]string{
			"currency_code": plan.Currency,
			"value":         FormatAmount(plan.AmountMinor, plan.Currency),
		},
	}
	payload := map[string]interface{}{
		"intent":         "AUTHORIZE",
		"purchase_units": []interface{}{unit},
	}
	if appCtx := applicationContext(); len(appCtx) > 0 {
		payload["application_context"] = appCtx
	}

	var out orderPayload
	if err := call(ctx, http.MethodPost, "/v2/checkout/orders", payload, &out); err != nil {
		return nil, err
	}
	view := out.view()
	view.ChargeAmount = plan.AmountMinor
	view.ChargeCurrency = plan.Currency
	view.FXRatePPM = plan.FXRatePPM
	if view.ID == "" {
		return nil, errors.New("paypal created no order id")
	}
	if view.ApprovalURL == "" {
		return nil, errors.New("paypal returned no approval link")
	}
	return view, nil
}

// GetOrder reads an order's current status, authorization and capture ids.
func GetOrder(ctx context.Context, orderID string) (*OrderView, error) {
	if !Enabled() {
		return nil, errors.New("paypal is not configured")
	}
	id := strings.TrimSpace(orderID)
	if id == "" {
		return nil, errors.New("paypal order id is required")
	}
	var out orderPayload
	if err := call(ctx, http.MethodGet, "/v2/checkout/orders/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.view(), nil
}

// AuthorizeOrder authorizes a buyer-approved order (funds held, capturable for 29 days).
func AuthorizeOrder(ctx context.Context, orderID string) (*OrderView, error) {
	if !Enabled() {
		return nil, errors.New("paypal is not configured")
	}
	id := strings.TrimSpace(orderID)
	if id == "" {
		return nil, errors.New("paypal order id is required")
	}
	var out orderPayload
	if err := call(ctx, http.MethodPost, "/v2/checkout/orders/"+url.PathEscape(id)+"/authorize",
		map[string]interface{}{}, &out); err != nil {
		return nil, err
	}
	return out.view(), nil
}

// CaptureOrder captures the order's authorization (resolving the authorization id first).
func CaptureOrder(ctx context.Context, orderID string) (string, error) {
	order, err := GetOrder(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.CaptureID != "" {
		return order.CaptureID, nil // already captured: idempotent
	}
	if order.AuthorizationID == "" {
		return "", errors.New("paypal order has no authorization to capture; the buyer has not approved yet")
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := call(ctx, http.MethodPost,
		"/v2/payments/authorizations/"+url.PathEscape(order.AuthorizationID)+"/capture",
		map[string]interface{}{}, &out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", errors.New("paypal capture returned no id")
	}
	return out.ID, nil
}

// RefundOrder refunds a captured order (full when amount is nil/zero) and returns the refund id.
func RefundOrder(ctx context.Context, orderID string, amountMinor *int64, reason string) (string, error) {
	order, err := GetOrder(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.CaptureID == "" {
		return "", errors.New("paypal payment has no captured transaction to refund")
	}
	body := map[string]interface{}{}
	if amountMinor != nil && *amountMinor > 0 {
		body["amount"] = map[string]string{
			"currency_code": order.Currency,
			"value":         FormatAmount(*amountMinor, order.Currency),
		}
	}
	if note := strings.TrimSpace(reason); note != "" {
		if len(note) > 255 {
			note = note[:255]
		}
		body["note_to_payer"] = note
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := call(ctx, http.MethodPost,
		"/v2/payments/captures/"+url.PathEscape(order.CaptureID)+"/refund", body, &out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", errors.New("paypal refund returned no id")
	}
	return out.ID, nil
}

// --- Webhook verification ---

// VerifyWebhook validates a webhook call with PayPal's verify-webhook-signature API. The raw
// payload is part of the signed message, so it must be passed unmodified.
func VerifyWebhook(headers http.Header, payload []byte) error {
	c := global.GVA_CONFIG.Payment.PayPal
	webhookID := strings.TrimSpace(c.WebhookID)
	if webhookID == "" {
		return errors.New("paypal webhook id is not configured")
	}
	if !Enabled() {
		return errors.New("paypal is not configured")
	}
	transmissionID := strings.TrimSpace(headers.Get("Paypal-Transmission-Id"))
	transmissionTime := strings.TrimSpace(headers.Get("Paypal-Transmission-Time"))
	signature := strings.TrimSpace(headers.Get("Paypal-Transmission-Sig"))
	certURL := strings.TrimSpace(headers.Get("Paypal-Cert-Url"))
	authAlgo := strings.TrimSpace(headers.Get("Paypal-Auth-Algo"))
	if transmissionID == "" || transmissionTime == "" || signature == "" || certURL == "" || authAlgo == "" {
		return errors.New("missing paypal webhook signature headers")
	}
	if !strings.HasPrefix(strings.ToLower(certURL), "https://") {
		return errors.New("paypal webhook certificate url must use https")
	}
	if len(payload) == 0 {
		return errors.New("empty paypal webhook payload")
	}

	body := map[string]interface{}{
		"auth_algo":         authAlgo,
		"cert_url":          certURL,
		"transmission_id":   transmissionID,
		"transmission_sig":  signature,
		"transmission_time": transmissionTime,
		"webhook_id":        webhookID,
		"webhook_event":     json.RawMessage(payload),
	}
	var out struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := call(context.Background(), http.MethodPost, "/v1/notify/webhook/verify-webhook-signature", body, &out); err != nil {
		return err
	}
	if out.VerificationStatus != "SUCCESS" {
		return errors.New("invalid paypal webhook signature")
	}
	return nil
}

// --- amounts ---

// FormatAmount renders minor units as the decimal string PayPal expects.
func FormatAmount(amountMinor int64, currency string) string {
	if zeroDecimalCurrencies[strings.ToUpper(strings.TrimSpace(currency))] {
		return strconv.FormatInt(amountMinor, 10)
	}
	neg := amountMinor < 0
	abs := amountMinor
	if neg {
		abs = -abs
	}
	out := fmt.Sprintf("%d.%02d", abs/100, abs%100)
	if neg {
		return "-" + out
	}
	return out
}

// ParseAmount converts a PayPal decimal amount string back to minor units, rejecting anything
// that is not a plain non-negative amount so a webhook can never widen the charged amount.
func ParseAmount(value, currency string) (int64, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, errors.New("amount is empty")
	}
	if strings.HasPrefix(v, "-") || strings.HasPrefix(v, "+") {
		return 0, errors.New("amount must be positive")
	}
	parts := strings.Split(v, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("malformed amount")
	}
	for _, r := range parts[0] {
		if r < '0' || r > '9' {
			return 0, errors.New("malformed amount")
		}
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("malformed amount")
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if zeroDecimalCurrencies[strings.ToUpper(strings.TrimSpace(currency))] {
		if frac != "" {
			return 0, errors.New("currency does not allow decimals")
		}
		return whole, nil
	}
	if len(frac) > 2 {
		return 0, errors.New("currency allows at most two decimals")
	}
	for _, r := range frac {
		if r < '0' || r > '9' {
			return 0, errors.New("malformed amount")
		}
	}
	for len(frac) < 2 {
		frac += "0"
	}
	cents, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, errors.New("malformed amount")
	}
	return whole*100 + cents, nil
}

// --- transport ---

func applicationContext() map[string]string {
	c := global.GVA_CONFIG.Payment.PayPal
	out := map[string]string{}
	if brand := strings.TrimSpace(c.BrandName); brand != "" {
		out["brand_name"] = brand
	}
	if ret := strings.TrimSpace(c.ReturnURL); ret != "" {
		out["return_url"] = ret
	}
	if cancel := strings.TrimSpace(c.CancelURL); cancel != "" {
		out["cancel_url"] = cancel
	}
	if len(out) > 0 {
		out["user_action"] = "PAY_NOW"
	}
	return out
}

func call(ctx context.Context, method, path string, body, out interface{}) error {
	base, err := baseURL()
	if err != nil {
		return err
	}
	token, err := accessToken(ctx)
	if err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method == http.MethodPost {
		// PayPal's idempotency key: a retried POST is not applied twice.
		req.Header.Set("PayPal-Request-Id", newRequestID())
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("paypal request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiError(resp.StatusCode, raw)
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return errors.New("malformed paypal response")
		}
	}
	return nil
}

func apiError(status int, raw []byte) error {
	var e struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &e)
	switch {
	case e.Name != "" && e.Message != "":
		return fmt.Errorf("paypal error %d %s: %s", status, e.Name, e.Message)
	case e.Name != "":
		return fmt.Errorf("paypal error %d %s", status, e.Name)
	default:
		return fmt.Errorf("paypal error %d", status)
	}
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b[:])
}

type orderPayload struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	PurchaseUnits []struct {
		Amount struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"amount"`
		Payments struct {
			Authorizations []struct {
				ID string `json:"id"`
			} `json:"authorizations"`
			Captures []struct {
				ID string `json:"id"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

func (o orderPayload) view() *OrderView {
	v := &OrderView{ID: o.ID, Status: o.Status}
	for _, l := range o.Links {
		if l.Rel == "approve" || l.Rel == "payer-action" {
			if v.ApprovalURL == "" {
				v.ApprovalURL = l.Href
			}
		}
	}
	if len(o.PurchaseUnits) > 0 {
		pu := o.PurchaseUnits[0]
		v.Currency = strings.ToUpper(pu.Amount.CurrencyCode)
		if minor, err := ParseAmount(pu.Amount.Value, v.Currency); err == nil {
			v.AmountMinor = minor
		}
		if n := len(pu.Payments.Authorizations); n > 0 {
			v.AuthorizationID = pu.Payments.Authorizations[n-1].ID
		}
		if n := len(pu.Payments.Captures); n > 0 {
			v.CaptureID = pu.Payments.Captures[n-1].ID
		}
	}
	return v
}
