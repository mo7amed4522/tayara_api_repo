package request

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     int `json:"page" form:"page"`         // page number
	PageSize int `json:"pageSize" form:"pageSize"` // Page size
}

func (r *PageInfo) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.Page <= 0 {
			r.Page = 1
		}
		switch {
		case r.PageSize > 100:
			r.PageSize = 100
		case r.PageSize <= 0:
			r.PageSize = 10
		}
		offset := (r.Page - 1) * r.PageSize
		return db.Offset(offset).Limit(r.PageSize)
	}
}

// GetById Find by id structure
type GetById struct {
	ID int `json:"id" form:"id"` // Primary Key ID
}

func (r *GetById) Uint() uint {
	return uint(r.ID)
}

type IdsReq struct {
	Ids []int `json:"ids" form:"ids"`
}

// GetAuthorityId Get role by id structure
type GetAuthorityId struct {
	AuthorityId uint `json:"authorityId" form:"authorityId"` // Character ID
}

type Empty struct{}

type RangeTime struct {
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
}

func (r RangeTime) GetStartTime() (t time.Time, err error) {
	if r.StartTime == "" {
		err = errors.New("start time is empty")
		return
	}
	t, err = time.Parse("2006-01-02 15:04", r.StartTime)
	//t, err = time.Parse(time.DateOnly, r.StartTime)
	if err != nil {
		return
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
	return
}

func (r RangeTime) GetEndTime() (t time.Time, err error) {
	if r.EndTime == "" {
		err = errors.New("end time is empty")
		return
	}
	t, err = time.Parse("2006-01-02 15:04", r.EndTime)
	if err != nil {
		return
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, time.Local)
	return
}
