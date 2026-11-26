package initialize

import (
	// blank-import domain seed initializers so their init() registers into the system chain
	_ "github.com/mo7amed4522/tayara_api/service/base"

	// blank-import demo/test seed packages (source/*) so their init() registers into the same chain.
	_ "github.com/mo7amed4522/tayara_api/source/base"
	_ "github.com/mo7amed4522/tayara_api/source/booking"
	_ "github.com/mo7amed4522/tayara_api/source/bxapi"
	_ "github.com/mo7amed4522/tayara_api/source/catalog"
	_ "github.com/mo7amed4522/tayara_api/source/dxapi"
	_ "github.com/mo7amed4522/tayara_api/source/engage"
	_ "github.com/mo7amed4522/tayara_api/source/payment"
	_ "github.com/mo7amed4522/tayara_api/source/support"
	_ "github.com/mo7amed4522/tayara_api/source/system"
)

func init() {
	// do nothing,only import source package so that inits can be registered
}
