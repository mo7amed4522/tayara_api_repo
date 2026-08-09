package response

type PageResult struct {
	List      interface{} `json:"list"`
	Total     int64       `json:"total"`
	TotalPage int64       `json:"totalPage"`
	Page      int         `json:"page"`
	PageSize  int         `json:"pageSize"`
}
