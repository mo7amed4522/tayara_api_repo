package dispatch

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/mo7amed4522/tayara_api/global"
	"go.uber.org/zap"
)

// FcmSender delivers push notifications via Firebase Cloud Messaging. It uses the same Firebase
// app initialized for auth. When Firebase messaging is unavailable (no creds), Enabled() is false
// and Send becomes a no-op so the app still boots and the notification is simply logged as pending.
type FcmSender struct{}

// Enabled reports whether Firebase messaging can be used.
func (FcmSender) Enabled() bool {
	return global.GVA_CONFIG.Firebase.Enabled
}

func fcmEnabled() bool {
	return global.GVA_CONFIG.Firebase.Enabled
}

// Send pushes a message to the given device tokens.
func (FcmSender) Send(ctx context.Context, tokens []string, title, body, imageURL string) error {
	if len(tokens) == 0 || !fcmEnabled() {
		return nil
	}
	client, err := messagingClient(ctx)
	if err != nil {
		global.GVA_LOG.Warn("fcm messaging client unavailable", zap.Error(err))
		return err
	}
	_, err = client.SendEachForMulticast(ctx, &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageURL,
		},
	})
	return err
}

// DataMessage is one push payload for a single device token. Silent messages carry data only (no
// notification block) and ask both platforms to wake the app without a banner.
type DataMessage struct {
	Token  string
	Title  string
	Body   string
	Data   map[string]string
	Silent bool
}

// SendDataMessage delivers a single data-carrying message. It is additive: Send (the existing
// multicast notification path) is untouched. A missing credential or empty token is a no-op.
func (FcmSender) SendDataMessage(ctx context.Context, msg DataMessage) error {
	if msg.Token == "" || !fcmEnabled() {
		return nil
	}
	client, err := messagingClient(ctx)
	if err != nil {
		global.GVA_LOG.Warn("fcm messaging client unavailable", zap.Error(err))
		return err
	}
	m := &messaging.Message{
		Token: msg.Token,
		Data:  msg.Data,
		Android: &messaging.AndroidConfig{
			Priority: "normal",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{Aps: &messaging.Aps{ContentAvailable: true}},
		},
	}
	if !msg.Silent {
		m.Notification = &messaging.Notification{Title: msg.Title, Body: msg.Body}
		m.Android.Priority = "high"
	}
	_, err = client.Send(ctx, m)
	return err
}

// IsInvalidToken reports whether FCM rejected the device token itself. Callers revoke the token
// instead of retrying it forever.
func IsInvalidToken(err error) bool {
	return messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err)
}

func messagingClient(ctx context.Context) (*messaging.Client, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, err
	}
	return app.Messaging(ctx)
}
