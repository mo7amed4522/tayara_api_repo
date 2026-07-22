package core

import (
	"fmt"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/initialize"
	"github.com/mo7amed4522/tayara_api/service/system"
	"go.uber.org/zap"
)

type server interface {
	ListenAndServe() error
}

func RunWindowsServer() {

	// Load JWT data from DB
	if global.GVA_DB != nil {
		system.LoadAll()
	}

	Router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)
	s := initServer(address, Router)

	global.GVA_LOG.Info("server run success on ", zap.String("address", address))

	fmt.Printf(`Welcome to use Tayara API Server, server is running on %s\n`, address)
	global.GVA_LOG.Error(s.ListenAndServe().Error())
}
