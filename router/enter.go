package router

import (
	"github.com/mo7amed4522/tayara_api/router/analytics"
	"github.com/mo7amed4522/tayara_api/router/base"
	"github.com/mo7amed4522/tayara_api/router/booking"
	"github.com/mo7amed4522/tayara_api/router/bxapi"
	"github.com/mo7amed4522/tayara_api/router/catalog"
	"github.com/mo7amed4522/tayara_api/router/dxapi"
	"github.com/mo7amed4522/tayara_api/router/engage"
	"github.com/mo7amed4522/tayara_api/router/payment"
	"github.com/mo7amed4522/tayara_api/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System    system.RouterGroup
	BxApi     bxapi.RouterGroup
	Manage    dxapi.RouterGroup
	Catalog   catalog.RouterGroup
	Booking   booking.RouterGroup
	Payment   payment.RouterGroup
	Engage    engage.RouterGroup
	Analytics analytics.RouterGroup
	Base      base.RouterGroup
}
