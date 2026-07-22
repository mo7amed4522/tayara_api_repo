package request

import (
	"github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.RangeTime
	request.PageInfo
}
