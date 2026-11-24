package engage_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/mo7amed4522/tayara_api/global"
	engageModel "github.com/mo7amed4522/tayara_api/model/engage"
	engageReq "github.com/mo7amed4522/tayara_api/model/engage/request"
	engageService "github.com/mo7amed4522/tayara_api/service/engage"
	"gorm.io/gorm"
)

func TestUpsertPushTokenUpdatesExistingInstallation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&engageModel.PushToken{}); err != nil {
		t.Fatalf("migrate push tokens: %v", err)
	}

	previousDB := global.GVA_DB
	global.GVA_DB = db
	defer func() { global.GVA_DB = previousDB }()

	service := engageService.UserEngageServices{}
	const userID = uint(42)
	req := engageReq.PushTokenReq{
		DeviceID: "installation-1",
		Platform: "android",
		Token:    "token-before",
	}
	if err := service.UpsertPushToken(userID, req); err != nil {
		t.Fatalf("create push token: %v", err)
	}

	var first engageModel.PushToken
	if err := db.Where("user_id = ? AND device_id = ?", userID, req.DeviceID).First(&first).Error; err != nil {
		t.Fatalf("load created push token: %v", err)
	}
	if err := db.Model(&first).Update("is_active", false).Error; err != nil {
		t.Fatalf("deactivate push token: %v", err)
	}

	req.Platform = "ios"
	req.Token = "token-after"
	req.AppVersion = "2.0.1"
	if err := service.UpsertPushToken(userID, req); err != nil {
		t.Fatalf("refresh push token: %v", err)
	}

	var refreshed engageModel.PushToken
	if err := db.Where("user_id = ? AND device_id = ?", userID, req.DeviceID).First(&refreshed).Error; err != nil {
		t.Fatalf("load refreshed push token: %v", err)
	}
	if refreshed.ID != first.ID {
		t.Fatalf("upsert created a second row: initial ID %d, refreshed ID %d", first.ID, refreshed.ID)
	}
	if refreshed.Token != req.Token || refreshed.Platform != req.Platform || refreshed.AppVersion != req.AppVersion {
		t.Fatalf("upsert did not refresh fields: %+v", refreshed)
	}
	if !refreshed.IsActive || refreshed.LastSeenAt.IsZero() {
		t.Fatalf("upsert did not reactivate/update last-seen: %+v", refreshed)
	}

	var count int64
	if err := db.Model(&engageModel.PushToken{}).Where("user_id = ? AND device_id = ?", userID, req.DeviceID).Count(&count).Error; err != nil {
		t.Fatalf("count push tokens: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one row per user and device, got %d", count)
	}
}
