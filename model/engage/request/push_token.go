package request

// PushTokenReq registers or refreshes the push token for an app installation.
type PushTokenReq struct {
	DeviceID   string `json:"device_id" binding:"required"`
	Platform   string `json:"platform" binding:"required,oneof=ios android"`
	Token      string `json:"token" binding:"required"`
	AppVersion string `json:"app_version"`
}
