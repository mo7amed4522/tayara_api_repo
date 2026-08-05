package v1

import (
	"github.com/mo7amed4522/tayara_api/api/v1/booking"
	"github.com/mo7amed4522/tayara_api/api/v1/bxapi"
	"github.com/mo7amed4522/tayara_api/api/v1/catalog"
	"github.com/mo7amed4522/tayara_api/api/v1/dxapi"
	"github.com/mo7amed4522/tayara_api/api/v1/engage"
	"github.com/mo7amed4522/tayara_api/api/v1/payment"
	"github.com/mo7amed4522/tayara_api/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup system.ApiGroup
	BxApiGroup     bxapi.ApiGroup
	ManageApiGroup dxapi.ApiGroup
	CatalogApiGroup catalog.ApiGroup
	BookingApiGroup booking.ApiGroup
	PaymentApiGroup payment.ApiGroup
	EngageApiGroup  engage.ApiGroup
}
