package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
)

// bnplProvider is a Buy-Now-Pay-Later adapter shared by Tabby and Tamara. It talks to each
// provider's sandbox HTTP API; card data is never involved. Real credentials come from env
// via config.yaml; without them Enabled() is false and calls fail cleanly.
type bnplProvider struct {
	name string
	cfg  config.BNPLProvider
}

// TabbyProvider returns the Tabby BNPL provider configured from global config.
func TabbyProvider() Provider {
	return bnplProvider{name: constants.PaymentProviderTabby, cfg: global.GVA_CONFIG.Payment.Tabby}
}

// TamaraProvider returns the Tamara BNPL provider configured from global config.
func TamaraProvider() Provider {
	return bnplProvider{name: constants.PaymentProviderTamara, cfg: global.GVA_CONFIG.Payment.Tamara}
}

func (b bnplProvider) Name() string  { return b.name }
func (b bnplProvider) Enabled() bool { return b.cfg.Enabled && b.cfg.SecretKey != "" }

func (b bnplProvider) currency() string {
	if b.cfg.Currency != "" {
		return b.cfg.Currency
	}
	return constants.CurrencyAED
}

func (b bnplProvider) createPath() string {
	if b.name == constants.PaymentProviderTabby {
		return "/api/v1/checkout"
	}
	return "/v1/orders"
}

func (b bnplProvider) refundPath() string {
	if b.name == constants.PaymentProviderTabby {
		return "/api/v1/refunds"
	}
	return "/v1/refunds"
}

func (b bnplProvider) CreatePayment(ctx context.Context, p CreatePaymentParams) (*CreatePaymentResult, error) {
	if !b.Enabled() {
		return nil, errors.New(b.name + " provider is not configured")
	}
	payload := map[string]interface{}{
		"amount":       p.Amount,
		"currency":     b.currency(),
		"reference_id": p.BookingCode,
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := b.postJSON(ctx, b.cfg.BaseURL+b.createPath(), payload, &out); err != nil {
		return nil, err
	}
	if out.ID == "" {
		return nil, errors.New(b.name + " created no session id")
	}
	return &CreatePaymentResult{Provider: b.name, IntentID: out.ID, Status: "requires_session"}, nil
}

// Capture is implicit for BNPL (settles when the customer completes the plan); nothing to call.
func (b bnplProvider) Capture(ctx context.Context, intentID string, amount *int64) error {
	return nil
}

func (b bnplProvider) Refund(ctx context.Context, intentID string, amount *int64, reason string) (string, error) {
	payload := map[string]interface{}{"order_id": intentID, "amount": amount, "reason": reason}
	var out struct {
		ID string `json:"id"`
	}
	if err := b.postJSON(ctx, b.cfg.BaseURL+b.refundPath(), payload, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (b bnplProvider) postJSON(ctx context.Context, url string, body interface{}, out interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.cfg.SecretKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s provider error %d: %s", b.name, resp.StatusCode, string(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}
