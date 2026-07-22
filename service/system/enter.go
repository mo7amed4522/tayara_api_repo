package system

type ServiceGroup struct {
	JwtService
	CasbinService
	AutoCodeService
	SystemConfigService
	OperationRecordService
}
