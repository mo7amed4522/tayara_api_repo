package system

import "github.com/mo7amed4522/tayara_api/service"

type ApiGroup struct {
	SysAdminProfileApi
	CasbinApi
	ExploreTopicApi
}

var (
	sysAdminProfile = service.ServiceGroupApp.SystemServiceGroup
	systemSvc       = service.ServiceGroupApp.SystemServiceGroup
)
