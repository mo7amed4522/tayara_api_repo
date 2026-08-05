package response

import "github.com/mo7amed4522/tayara_api/model/bxapi"

// UserLoginResponse returned on successful user login (token + profile).
type UserLoginResponse struct {
	User      bxapi.BxsUserProfile `json:"user"`
	Token     string               `json:"token"`
	ExpiresAt int64                `json:"expiresAt"`
}
