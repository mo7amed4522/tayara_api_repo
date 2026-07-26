package response

import "github.com/mo7amed4522/tayara_api/model/dxapi"

type DxAgentLoginResponse struct {
	Agent     dxapi.DxAgentProfile `json:"admin"`
	Token     string               `json:"token"`
	ExpiresAt int64                `json:"expiresAt"`
}
