package bxapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/mo7amed4522/tayara_api/model/bxapi/request"
	bxapiReq "github.com/mo7amed4522/tayara_api/model/bxapi/request"
)

func HomeSearchKey(req bxapiReq.CustomerSearchReq) string {
	return cacheKey("api:v1:customer:search", struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Keyword  string `json:"keyword"`
	}{req.Page, req.PageSize, req.Keyword})
}

func ExploreTripsKey(req bxapiReq.CustomerHomeExploreReq) string {
	return cacheKey("api:v1:customer:explore:trips", struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Category string `json:"category"`
	}{req.Page, req.PageSize, req.Category})
}

func DestinationsKey(page, pageSize int) string {
	return cacheKey("api:v1:customer:home:destinations", struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}{page, pageSize})
}

func DealsKey(req bxapiReq.CustomerHomeDealsReq) string {
	return cacheKey("api:v1:customer:home:deals", struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Category string `json:"category"`
	}{req.Page, req.PageSize, req.Category})
}

func TopRatedAgentsKey(page, pageSize int) string {
	return cacheKey("api:v1:customer:home:agents", struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}{page, pageSize})
}

func HomeAdsKey(page, pageSize int) string {
	return cacheKey("api:v1:customer:home:ads", struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}{page, pageSize})
}

func ExploreTopicsKey(page, pageSize int) string {
	return cacheKey("api:v1:customer:explore:topics", struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}{page, pageSize})
}

func ExploreTopicDetailKey(code string) string {
	return cacheKey("api:v1:customer:explore:topics:detail", struct {
		Code string `json:"code"`
	}{code})
}

func ExploreTopicPackagesKey(code string, req request.CustomerExploreTopicPackagesReq) string {
	return cacheKey("api:v1:customer:explore:topics:packages", struct {
		Code     string `json:"code"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Category string `json:"category"`
	}{code, req.Page, req.PageSize, req.Category})
}

func AgentDetailKey(agentID uint, destCountryID *uint, page, pageSize int) string {
	destinationID := uint(0)
	if destCountryID != nil {
		destinationID = *destCountryID
	}
	return cacheKey("api:v1:customer:agents:detail", struct {
		AgentID       uint `json:"agentId"`
		DestinationID uint `json:"destinationCountryId"`
		Page          int  `json:"page"`
		PageSize      int  `json:"pageSize"`
	}{agentID, destinationID, page, pageSize})
}

func PackageDetailKey(id uint) string {
	return cacheKey("api:v1:customer:packages:detail", struct {
		ID uint `json:"id"`
	}{id})
}

func PackageReviewsKey(packageID uint, page, pageSize int) string {
	return cacheKey("api:v1:customer:packages:reviews", struct {
		PackageID uint `json:"packageId"`
		Page      int  `json:"page"`
		PageSize  int  `json:"pageSize"`
	}{packageID, page, pageSize})
}

func GlobalSearchKey(keyword string) string {
	return cacheKey("api:v1:admin:search:global", struct {
		Keyword string `json:"keyword"`
	}{keyword})
}

func RecommendedForYouKey(limit int) string {
	return cacheKey("api:v1:customer:home:recommended:public", struct {
		Limit int `json:"limit"`
	}{limit})
}

func AgentGetMyTopicsKey(agentID uint, page, pageSize int) string {
	return cacheKey("api:v1:agent:explore:topics", struct {
		AgentID  uint `json:"agentId"`
		Page     int  `json:"page"`
		PageSize int  `json:"pageSize"`
	}{agentID, page, pageSize})
}

func AgentGetPackageTopicsKey(agentID, packageID uint) string {
	return cacheKey("api:v1:agent:explore:packages:topics", struct {
		AgentID   uint `json:"agentId"`
		PackageID uint `json:"packageId"`
	}{agentID, packageID})
}

func cacheKey(endpoint string, parameters any) string {
	encoded, err := json.Marshal(parameters)
	if err != nil {
		panic("cache key parameters are not JSON-serializable: " + err.Error())
	}
	hash := sha256.Sum256(encoded)
	return "bxapi:" + strings.ToLower(endpoint) + ":" + hex.EncodeToString(hash[:])
}
