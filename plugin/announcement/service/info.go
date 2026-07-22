package service

import (
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/plugin/announcement/model"
	"github.com/mo7amed4522/tayara_api/plugin/announcement/model/request"
)

var Info = new(info)

type info struct{}

// CreateInfo Create Announcement Record
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (s *info) CreateInfo(info *model.Info) (err error) {
	err = global.GVA_DB.Create(info).Error
	return err
}

// DeleteInfo Delete announcement record
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (s *info) DeleteInfo(ID string) (err error) {
	err = global.GVA_DB.Delete(&model.Info{}, "id = ?", ID).Error
	return err
}

// DeleteInfoByIds Batch delete announcement record
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (s *info) DeleteInfoByIds(IDs []string) (err error) {
	err = global.GVA_DB.Delete(&[]model.Info{}, "id in ?", IDs).Error
	return err
}

// UpdateInfo Update announcement record
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (s *info) UpdateInfo(info model.Info) (err error) {
	err = global.GVA_DB.Model(&model.Info{}).Where("id = ?", info.CreatedAt).Updates(&info).Error
	return err
}

// GetInfo Get announcement record by ID
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (s *info) GetInfo(ID string) (info model.Info, err error) {
	err = global.GVA_DB.Where("id = ?", ID).First(&info).Error
	return
}

// GetInfoInfoList Get announcement list by page
// Author [mo7amed4522](https://github.com/mo7amed4522)
func (s *info) GetInfoInfoList(info request.InfoSearch) (list []model.Info, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// Create db
	db := global.GVA_DB.Model(&model.Info{})
	var infos []model.Info
	// If there is a condition search, the search statement will be automatically created below
	if info.StartCreatedAt != nil && info.EndCreatedAt != nil {
		db = db.Where("created_at BETWEEN ? AND ?", info.StartCreatedAt, info.EndCreatedAt)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	err = db.Find(&infos).Error
	return infos, total, err
}
func (s *info) GetInfoDataSource() (res map[string][]map[string]any, err error) {
	res = make(map[string][]map[string]any)

	userID := make([]map[string]any, 0)
	global.GVA_DB.Table("sys_users").Select("nick_name as label,id as value").Scan(&userID)
	res["userID"] = userID
	return
}
