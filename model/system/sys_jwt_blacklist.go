package system

import "github.com/mo7amed4522/tayara_api/global"

type JwtBlacklist struct {
	ID  uint   `json:"id" gorm:"primarykey;comment:id of JWT"` // UUID
	Jwt string `gorm:"type:text;comment:jwt"`
	global.GVA_MODEL
}
