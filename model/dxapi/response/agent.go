package response

import "github.com/mo7amed4522/tayara_api/model/dxapi"

// AgentLoginResponse returned on successful agent login (token + profile).
type AgentLoginResponse struct {
	Agent     dxapi.DxAgentProfile `json:"agent"`
	Token     string               `json:"token"`
	ExpiresAt int64                `json:"expiresAt"`
}

// AgentDetail an agent profile plus its documents and notes (admin view).
type AgentDetail struct {
	dxapi.DxAgentProfile
	Documents []dxapi.AgentDocument `json:"documents"`
	Notes     []dxapi.AgentNote     `json:"notes"`
}
