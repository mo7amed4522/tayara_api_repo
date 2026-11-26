package request

// RegisterPushTokenReq registers (or refreshes) one device token for the calling actor.
// Registration is idempotent: the same token twice returns the existing row.
type RegisterPushTokenReq struct {
	Platform   string `json:"platform" form:"platform" binding:"required,oneof=ios android web"`
	Token      string `json:"token" form:"token" binding:"required,min=16,max=512"`
	AppVersion string `json:"appVersion" form:"appVersion" binding:"max=32"`
}

// RevokePushTokenReq revokes every active token of the caller on one platform by appending a
// revocation row per token.
type RevokePushTokenReq struct {
	Platform string `json:"platform" form:"platform" binding:"required,oneof=ios android web"`
}
