package search

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	searchResp "github.com/mo7amed4522/tayara_api/model/search/response"
	bxapiSvc "github.com/mo7amed4522/tayara_api/service/bxapi"
	"github.com/mo7amed4522/tayara_api/utils/cache"
)

// SearchServices implements the admin global search across entities.
type SearchServices struct{}

// per-group caps keep a search result small and predictable.
const pageLimit = 10

// Global runs a LIKE search across users, agents, packages and bookings and aggregates the
// matches into groups. Admin-scoped read: no row-scoping is applied because an admin sees all.
func (s SearchServices) Global(keyword string) (*searchResp.SearchResult, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	kw := strings.TrimSpace(keyword)
	if kw == "" {
		return nil, errors.New("keyword required")
	}
	result, _, err := cache.GetOrLoadJSON(context.Background(), bxapiSvc.GlobalSearchKey(kw), time.Hour, func() (*searchResp.SearchResult, error) {
		return s.globalUncached(kw)
	})
	return result, err
}

func (s SearchServices) globalUncached(kw string) (*searchResp.SearchResult, error) {
	like := "%" + kw + "%"
	out := &searchResp.SearchResult{Keyword: kw}

	var total int64

	// end-users
	{
		var rows []bxapi.BxsUserProfile
		if err := global.GVA_DB.Where("full_name LIKE ? OR username LIKE ? OR email LIKE ? OR phone LIKE ?",
			like, like, like, like).Limit(pageLimit).Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			out.Users = append(out.Users, searchResp.UserHit{
				ID: rows[i].ID, Name: rows[i].FullName, Email: rows[i].Email,
				Phone: rows[i].Phone, Status: rows[i].Status,
			})
		}
		if err := global.GVA_DB.Model(&bxapi.BxsUserProfile{}).
			Where("full_name LIKE ? OR username LIKE ? OR email LIKE ? OR phone LIKE ?",
				like, like, like, like).Count(&total).Error; err != nil {
			return nil, err
		}
	}

	// agents
	{
		var a []dxapi.DxAgentProfile
		if err := global.GVA_DB.Where("company_name LIKE ? OR business_email LIKE ? OR license_number LIKE ?",
			like, like, like).Limit(pageLimit).Find(&a).Error; err != nil {
			return nil, err
		}
		for i := range a {
			out.Agents = append(out.Agents, searchResp.AgentHit{
				ID: a[i].ID, CompanyName: a[i].CompanyName,
				BusinessEmail: a[i].BusinessEmail, Status: a[i].Status,
			})
		}
		var n int64
		if err := global.GVA_DB.Model(&dxapi.DxAgentProfile{}).
			Where("company_name LIKE ? OR business_email LIKE ? OR license_number LIKE ?",
				like, like, like).Count(&n).Error; err != nil {
			return nil, err
		}
		total += n
	}

	// packages
	{
		var pkgs []catalog.Package
		if err := global.GVA_DB.Where("name LIKE ?", like).Limit(pageLimit).Find(&pkgs).Error; err != nil {
			return nil, err
		}
		for i := range pkgs {
			out.Packages = append(out.Packages, searchResp.PackageHit{
				ID: pkgs[i].ID, AgentID: pkgs[i].AgentID, Name: pkgs[i].Name,
				Status: pkgs[i].Status, DisplayPrice: pkgs[i].DisplayPrice,
			})
		}
		var n int64
		if err := global.GVA_DB.Model(&catalog.Package{}).Where("name LIKE ?", like).Count(&n).Error; err != nil {
			return nil, err
		}
		total += n
	}

	// bookings
	{
		var bks []booking.Booking
		if err := global.GVA_DB.Where("booking_code LIKE ?", like).Limit(pageLimit).Find(&bks).Error; err != nil {
			return nil, err
		}
		for i := range bks {
			out.Bookings = append(out.Bookings, searchResp.BookingHit{
				ID: bks[i].ID, BookingCode: bks[i].BookingCode,
				BookingStatus: bks[i].BookingStatus, AmountGross: bks[i].AmountGross,
			})
		}
		var n int64
		if err := global.GVA_DB.Model(&booking.Booking{}).Where("booking_code LIKE ?", like).Count(&n).Error; err != nil {
			return nil, err
		}
		total += n
	}

	total += int64(len(out.Users))
	out.Total = total
	return out, nil
}
