package support

import (
	"errors"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
)

// LastSeenServices owns the append-only last_seen projection: one row per graceful disconnect
// and one per actor per flush interval, latest row wins on read.
type LastSeenServices struct{}

// Record appends one support_presence row.
func (s LastSeenServices) Record(actorType string, actorID uint, at time.Time) error {
	if global.GVA_DB == nil || actorID == 0 || actorType == "" {
		return errors.New("db not init")
	}
	return global.GVA_DB.Create(&supportModel.Presence{
		ActorType: actorType, ActorID: actorID, LastSeenAt: at.UTC(),
	}).Error
}

// Visible reports whether a last-seen value may be shown for this actor type. Applied
// server-side, before the DTO is built, so a hidden role's value is null on the wire.
func (s LastSeenServices) Visible(actorType string) bool {
	settings := SettingsServices{}
	if !settings.LastSeenDisplay() {
		return false
	}
	for _, hidden := range settings.HiddenLastSeenRoles() {
		if strings.EqualFold(hidden, actorType) {
			return false
		}
	}
	return true
}

// Latest returns the newest last_seen_at for an actor, or nil.
func (s LastSeenServices) Latest(actorType string, actorID uint) *time.Time {
	if global.GVA_DB == nil || actorID == 0 {
		return nil
	}
	var row supportModel.Presence
	if err := global.GVA_DB.Where("actor_type = ? AND actor_id = ?", actorType, actorID).
		Order("id desc").First(&row).Error; err != nil {
		return nil
	}
	t := row.LastSeenAt
	return &t
}

// LatestBatch returns the newest last_seen_at per actor using one portable MAX(id) subquery.
func (s LastSeenServices) LatestBatch() (map[string]*time.Time, error) {
	out := map[string]*time.Time{}
	if global.GVA_DB == nil {
		return out, nil
	}
	sub := global.GVA_DB.Model(&supportModel.Presence{}).Select("MAX(id)").Group("actor_type, actor_id")
	var rows []supportModel.Presence
	if err := global.GVA_DB.Where("id IN (?)", sub).Find(&rows).Error; err != nil {
		return out, err
	}
	for _, r := range rows {
		t := r.LastSeenAt
		out[r.ActorType+":"+uintString(r.ActorID)] = &t
	}
	return out, nil
}

// Flush appends a last_seen row for every actor currently holding a live connection.
func (s LastSeenServices) Flush() error {
	store := Store()
	if store == nil {
		return nil
	}
	actors, err := store.OnlineActors(contextBackground())
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, a := range actors {
		_ = s.Record(a.ActorType, a.ActorID, now)
	}
	return nil
}
