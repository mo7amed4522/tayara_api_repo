package initialize

import (
	// blank-import domain seed initializers so their init() registers into the system chain
	_ "github.com/mo7amed4522/tayara_api/service/base"
)

func init() {
	// do nothing,only import source package so that inits can be registered
}
