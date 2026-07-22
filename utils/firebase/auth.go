package firebase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/mo7amed4522/tayara_api/global"
	"google.golang.org/api/option"
)

const (
	SignInProviderGoogle   = "google.com"
	SignInProviderApple    = "apple.com"
	SignInProviderFacebook = "facebook.com"
	SignInProviderPhone    = "phone"
)

var (
	authClient *auth.Client
	initOnce   sync.Once
	initErr    error
)

func Init() error {
	initOnce.Do(func() {
		cfg := global.GVA_CONFIG.Firebase
		if !cfg.Enabled {
			return
		}
		if cfg.CredentialsFile == "" {
			initErr = errors.New("firebase credentials-file is required when firebase is enabled")
			return
		}
		if cfg.ProjectID == "" {
			initErr = errors.New("firebase project-id is required when firebase is enabled")
			return
		}

		app, err := firebase.NewApp(context.Background(), &firebase.Config{
			ProjectID: cfg.ProjectID,
		}, option.WithCredentialsFile(cfg.CredentialsFile))
		if err != nil {
			initErr = fmt.Errorf("firebase app init: %w", err)
			return
		}

		client, err := app.Auth(context.Background())
		if err != nil {
			initErr = fmt.Errorf("firebase auth client init: %w", err)
			return
		}
		authClient = client
	})
	return initErr
}

func Enabled() bool {
	return global.GVA_CONFIG.Firebase.Enabled && authClient != nil
}

func AuthClient() (*auth.Client, error) {
	if !Enabled() {
		return nil, errors.New("firebase auth is not enabled or not initialized")
	}
	return authClient, nil
}

func VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	client, err := AuthClient()
	if err != nil {
		return nil, err
	}
	return client.VerifyIDToken(ctx, idToken)
}

func SignInProvider(token *auth.Token) string {
	if token == nil {
		return ""
	}
	firebaseClaims, ok := token.Claims["firebase"].(map[string]interface{})
	if !ok {
		return ""
	}
	provider, _ := firebaseClaims["sign_in_provider"].(string)
	return provider
}

func ProviderEnabled(provider string) bool {
	cfg := global.GVA_CONFIG.Firebase.Providers
	switch provider {
	case SignInProviderGoogle:
		return cfg.Google
	case SignInProviderApple:
		return cfg.Apple
	case SignInProviderFacebook:
		return cfg.Facebook
	case SignInProviderPhone:
		return cfg.Phone
	default:
		return false
	}
}

func ValidateProvider(token *auth.Token, expectedProvider string) error {
	actual := SignInProvider(token)
	if actual != expectedProvider {
		return fmt.Errorf("token provider mismatch: expected %s, got %s", expectedProvider, actual)
	}
	if !ProviderEnabled(expectedProvider) {
		return fmt.Errorf("sign-in provider %s is disabled", expectedProvider)
	}
	return nil
}
