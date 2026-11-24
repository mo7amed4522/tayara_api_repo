package engage

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	"gorm.io/gorm/clause"
)

// UpsertPushToken atomically creates or refreshes a user's token for an app installation.
func (u UserEngageServices) UpsertPushToken(userID uint, req engageReq.PushTokenReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}

	now := time.Now()
	token := engage.PushToken{
		UserID:     userID,
		DeviceID:   req.DeviceID,
		Platform:   req.Platform,
		Token:      req.Token,
		AppVersion: req.AppVersion,
		LastSeenAt: now,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// The unique user/device key makes concurrent requests converge on one row.
	return global.GVA_DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"token",
			"platform",
			"app_version",
			"last_seen_at",
			"is_active",
			"updated_at",
		}),
	}).Create(&token).Error
}
