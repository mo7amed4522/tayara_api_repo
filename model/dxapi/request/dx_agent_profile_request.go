package request

type DxCreateAgentProfileRequest struct {
	CompanyName       string  `json:"companyName" form:"companyName"`
	Password          string  `json:"password" form:"password"`
	ResponsibleName   string  `json:"responsibleName" form:"responsibleName"`
	HeaderImg         string  `json:"headerImg" form:"headerImg"`
	PhoneCountryCode  string  `json:"phoneCountryCode" form:"phoneCountryCode"`
	Phone             string  `json:"phone" form:"phone"`
	BusinessEmail     string  `json:"businessEmail" form:"businessEmail"`
	Country           string  `json:"country" form:"country"`
	CompanyAddress    string  `json:"companyAddress" form:"companyAddress"`
	LicenseNumber     string  `json:"licenseNumber" form:"licenseNumber"`
	PlanID            *uint   `json:"planId" form:"planId"`
	ShortDescription  *string `json:"shortDescription" form:"shortDescription"`
	LongDescription   *string `json:"longDescription" form:"longDescription"`
	CommissionPercent int64   `json:"commissionPercent" form:"commissionPercent"`
	SLAHours          int     `json:"slaHours" form:"slaHours"`
}

type DxLoginAgentRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}
