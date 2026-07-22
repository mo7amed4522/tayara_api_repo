package response

import "github.com/mo7amed4522/tayara_api/model/system"

type SysLoginResponse struct {
	Admin     system.SysAdminProfile `json:"admin"`
	Token     string                 `json:"token"`
	ExpiresAt int64                  `json:"expiresAt"`
}
