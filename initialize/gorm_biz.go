package initialize

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/system"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		// base / dictionaries
		&base.Country{},
		&base.City{},
		&base.Nationality{},
		&base.DocumentType{},
		&base.SystemSetting{},
		// user
		&bxapi.BxDict{},
		&bxapi.BxsUserProfile{},
		// agent
		&dxapi.DxAgentProfile{},
		&dxapi.AgentPlan{},
		&dxapi.AgentDocument{},
		&dxapi.AgentNote{},
		// system
		&system.SysAdminProfile{},
		&system.JwtBlacklist{},
	)
	if err != nil {
		return err
	}
	return nil
}
