package request

// CasbinInfo Casbin info structure
type CasbinInfo struct {
	Path   string `json:"path"`   // Path
	Method string `json:"method"` // Method
}

// CasbinInReceive Casbin structure for input parameters
type CasbinInReceive struct {
	CasbinInfos []CasbinInfo `json:"casbinInfos"`
}

// CasbinSyncReq admin replaces the full route matrix for an authority (role).
type CasbinSyncReq struct {
	AuthorityID uint         `json:"authorityId" binding:"required"`
	Policies    []CasbinInfo `json:"policies"`
}

// CasbinUpdateReq admin updates a single route rule (path+method) for all authorities using it.
type CasbinUpdateReq struct {
	OldPath   string `json:"oldPath" binding:"required"`
	NewPath   string `json:"newPath" binding:"required"`
	OldMethod string `json:"oldMethod" binding:"required"`
	NewMethod string `json:"newMethod" binding:"required"`
}

func DefaultCasbin() []CasbinInfo {
	return []CasbinInfo{
		{Path: "/menu/getMenu", Method: "POST"},
		{Path: "/jwt/jsonInBlacklist", Method: "POST"},
		{Path: "/base/login", Method: "POST"},
		{Path: "/user/changePassword", Method: "POST"},
		{Path: "/user/setUserAuthority", Method: "POST"},
		{Path: "/user/getUserInfo", Method: "GET"},
		{Path: "/user/setSelfInfo", Method: "PUT"},
		{Path: "/fileUploadAndDownload/upload", Method: "POST"},
		{Path: "/sysDictionary/findSysDictionary", Method: "GET"},
	}
}
