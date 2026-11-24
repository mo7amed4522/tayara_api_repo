package engage_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/global"
	engageModel "github.com/mo7amed4522/tayara_api/model/engage"
	engageService "github.com/mo7amed4522/tayara_api/service/engage"
	"gorm.io/gorm"
)

func TestEmitWithImagePersistsNotification(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&engageModel.Notification{}, &engageModel.NotificationTemplate{}); err != nil {
		t.Fatalf("migrate notification models: %v", err)
	}

	previousDB, previousConfig := global.GVA_DB, global.GVA_CONFIG
	global.GVA_DB = db
	global.GVA_CONFIG = config.Server{}
	defer func() {
		global.GVA_DB = previousDB
		global.GVA_CONFIG = previousConfig
	}()

	const (
		title    = "Welcome at Taiyara Marketplace"
		body     = "We here to help you in any trip and booking you do as you wish"
		imageURL = "https://raw.githubusercontent.com/tayara_app/refs/heads/master/assets/images/app_logo.png"
	)
	(engageService.NotificationServices{}).EmitWithImage(
		"welcome", "user", 51, "push", title, body, imageURL,
	)

	var notification engageModel.Notification
	if err := db.Where("recipient_id = ?", 51).First(&notification).Error; err != nil {
		t.Fatalf("load saved notification: %v", err)
	}
	if notification.Title != title || notification.Body != body || notification.ImageURL != imageURL {
		t.Fatalf("saved notification does not match request: %+v", notification)
	}
	if notification.Status != "pending" {
		t.Fatalf("notification without an enabled push sender should remain pending, got %q", notification.Status)
	}
}
