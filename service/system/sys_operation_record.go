package system

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/system"
	systemReq "github.com/mo7amed4522/tayara_api/model/system/request"
)

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: CreateSysOperationRecord
//@description: Create record
//@param: sysOperationRecord model.SysOperationRecord
//@return: err error

type OperationRecordService struct{}

var OperationRecordServiceApp = new(OperationRecordService)

func (operationRecordService *OperationRecordService) CreateSysOperationRecord(sysOperationRecord system.SysOperationRecord) (err error) {
	err = global.GVA_DB.Create(&sysOperationRecord).Error
	return err
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: DeleteSysOperationRecordByIds
//@description: Batch delete record
//@param: ids request.IdsReq
//@return: err error

func (operationRecordService *OperationRecordService) DeleteSysOperationRecordByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]system.SysOperationRecord{}, "id in (?)", ids.Ids).Error
	return err
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: DeleteSysOperationRecord
//@description: Delete operation record
//@param: sysOperationRecord model.SysOperationRecord
//@return: err error

func (operationRecordService *OperationRecordService) DeleteSysOperationRecord(sysOperationRecord system.SysOperationRecord) (err error) {
	err = global.GVA_DB.Delete(&sysOperationRecord).Error
	return err
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: GetSysOperationRecord
//@description: Get single operation record by id
//@param: id uint
//@return: sysOperationRecord system.SysOperationRecord, err error

func (operationRecordService *OperationRecordService) GetSysOperationRecord(id uint) (sysOperationRecord system.SysOperationRecord, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sysOperationRecord).Error
	return
}

//@author: [mo7amed4522](https://github.com/mo7amed4522)
//@function: GetSysOperationRecordInfoList
//@description: Get operation record list by page
//@param: info systemReq.SysOperationRecordSearch
//@return: list interface{}, total int64, err error

func (operationRecordService *OperationRecordService) GetSysOperationRecordInfoList(info systemReq.SysOperationRecordSearch) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// Create db
	db := global.GVA_DB.Model(&system.SysOperationRecord{})
	var sysOperationRecords []system.SysOperationRecord
	// If there is a condition search, the search statement will be automatically created below
	if info.Method != "" {
		db = db.Where("method = ?", info.Method)
	}
	if info.Path != "" {
		db = db.Where("path LIKE ?", "%"+info.Path+"%")
	}
	if info.SysOperationRecord.Status != 0 {
		db = db.Where("status = ?", info.SysOperationRecord.Status)
	}
	if info.SysOperationRecord.UserID != 0 {
		db = db.Where("user_id = ?", info.SysOperationRecord.UserID)
	}
	if info.SysOperationRecord.Body != "" {
		db = db.Where("body LIKE ?", "%"+info.Body+"%")
	}
	if info.SysOperationRecord.ShortcutKey == 1 {
		//UTR order reset
		db = db.Where("path = ?", "/orderPayReversal/createOrderPayReversal")
	}
	if info.SysOperationRecord.ShortcutKey == 2 {
		//Withdrawal audit
		db = db.Where("path = ?", "/orderPayout/orderReset")
	}
	if info.SysOperationRecord.ShortcutKey == 3 {
		//Withdrawal replenishment
		db = db.Where("path = ?", "/orderPayout/replenishment")
	}
	if info.SysOperationRecord.ShortcutKey == 4 {
		//Withdrawal lock order
		db = db.Where("path = ?", "/orderPayout/lock")
	}

	if startTime, err := info.GetStartTime(); err == nil {
		db = db.Where("created_at >= ?", startTime)
	}

	if endTime, err := info.GetEndTime(); err == nil {
		db = db.Where("created_at <= ?", endTime)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("id desc").Limit(limit).Offset(offset).Preload("User").Find(&sysOperationRecords).Error
	return sysOperationRecords, total, err
}
