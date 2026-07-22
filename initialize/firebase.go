package initialize

import (
	"github.com/mo7amed4522/tayara_api/global"
	fbutil "github.com/mo7amed4522/tayara_api/utils/firebase"
	"go.uber.org/zap"
)

func Firebase() {
	if !global.GVA_CONFIG.Firebase.Enabled {
		return
	}
	if err := fbutil.Init(); err != nil {
		global.GVA_LOG.Error("Firebase initialization failed", zap.Error(err))
		panic(err)
	}
}
