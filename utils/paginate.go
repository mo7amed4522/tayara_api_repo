package utils

import (
	"gorm.io/gorm"
)

// PaginateResult pagination result structure
type PaginateResult struct {
	Total     int64       `json:"total"`     // Total number of records
	Page      int         `json:"page"`      // Current page number
	PageSize  int         `json:"pageSize"`  // Page size
	TotalPage int64       `json:"totalPage"` // Total number of pages
	Data      interface{} `json:"data"`      // data
}

// Paginate GORM pagination query tool function
// db: GORM database connection
// page: page number (starts from 1)
// pageSize: page size
// result: pointer to a slice to store the query results
// returns: paginated result and error information
func Paginate(db *gorm.DB, page, pageSize int, result interface{}) (*PaginateResult, error) {
	// set default values
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// get total number of records
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// calculate offset
	offset := (page - 1) * pageSize

	// execute pagination query
	err = db.Offset(offset).Limit(pageSize).Find(result).Error
	if err != nil {
		return nil, err
	}

	// calculate total number of pages
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)

	// build pagination result
	paginateResult := &PaginateResult{
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
		Data:      result,
	}

	return paginateResult, nil
}

// PaginateWithCustomQuery custom query pagination tool function
// Suitable for scenarios requiring complex query conditions
func PaginateWithCustomQuery(countDB, queryDB *gorm.DB, page, pageSize int, result interface{}) (*PaginateResult, error) {
	// set default values
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// get total number of records (using countDB)
	var total int64
	err := countDB.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// calculate offset
	offset := (page - 1) * pageSize

	// execute pagination query (using queryDB)
	err = queryDB.Offset(offset).Limit(pageSize).Find(result).Error
	if err != nil {
		return nil, err
	}

	// calculate total number of pages
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)

	// build pagination result
	paginateResult := &PaginateResult{
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
		Data:      result,
	}

	return paginateResult, nil
}

// BuildPaginateResponse build pagination response data
// Used to return uniform pagination data in the API layer
func BuildPaginateResponse(data interface{}, total int64, page, pageSize int) map[string]interface{} {
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	return map[string]interface{}{
		"list":       data,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}
}
