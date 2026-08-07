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
func (FcmSender) Send(ctx context.Context, tokens []string, title, body string) error {
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
			Title: title,
			Body:  body,
		},
	})
	return err
}

func messagingClient(ctx context.Context) (*messaging.Client, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, err
	}
	return app.Messaging(ctx)
}
