package router

import "github.com/mo7amed4522/tayara_api/plugin/announcement/api"

var (
	Router  = new(router)
	apiInfo = api.Api.Info
)

type router struct{ Info info }
