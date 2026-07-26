package system

import api "github.com/mo7amed4522/tayara_api/api/v1"

type RouterGroup struct {
	SysAdminProfileRouter
}

var (
	sysAdminProfileApi = api.ApiGroupApp.SystemApiGroup.SysAdminProfileApi
)
