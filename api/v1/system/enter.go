package system

import "github.com/mo7amed4522/tayara_api/service"

type ApiGroup struct {
	SysAdminProfileApi
}

var (
	sysAdminProfile = service.ServiceGroupApp.SystemServiceGroup
)
