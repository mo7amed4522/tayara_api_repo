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
		// Degrade rather than abort: Firebase only gates optional provider login (Google/Apple/
		// phone), and every consumer already guards with fbutil.Enabled() and returns a clean
		// "not enabled" error when the client is absent. A missing/unreachable credentials file
		// (e.g. the gitignored service-account JSON) must not prevent the whole API from booting
		// against its databases in a fresh deploy.
		global.GVA_LOG.Warn("Firebase initialization failed; running degraded (provider login disabled)", zap.Error(err))
		return
	}
}
