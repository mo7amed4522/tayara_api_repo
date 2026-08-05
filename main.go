package main

import (
	"github.com/mo7amed4522/tayara_api/core"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/initialize"
	"github.com/mo7amed4522/tayara_api/utils"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

// @title                       Tayara Swagger API Interface Documentation
// @version                     v2.7.7
// @description                 Full-stack development basic platform using gin for extremely fast development
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /
func main() {
	global.GVA_VP = core.Viper() // Initialize Viper
	initialize.OtherInit()
	global.GVA_LOG = core.Zap() // Initialize zap log library
	zap.ReplaceGlobals(global.GVA_LOG)
	initialize.Firebase()
	global.GVA_DB = initialize.Gorm() // Initialize MySQL for Tayara project
	initialize.Timer()                // Initialize cron
	utils.Init()
	initialize.Redis()
	initialize.DBList()
	// MongoDB (qmgo) powers the payment ledger. Non-fatal: the app still boots if Mongo is
	// not configured, but payment lifecycle relies on it, so it must be up in production.
	if err := initialize.Mongo.Initialization(); err != nil {
		global.GVA_LOG.Error("mongodb initialization failed", zap.Error(err))
	}
	if global.GVA_DB != nil {
		initialize.RegisterTables() // Initialize tables
		initialize.InitializeData() // Initialize data
		// Close database connection before program ends
		db, _ := global.GVA_DB.DB()
		defer db.Close()
	}
	core.RunWindowsServer()
}
