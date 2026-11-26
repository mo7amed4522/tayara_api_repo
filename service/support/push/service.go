package push

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
)

// PushServices owns token registration, revocation and the failed-dispatch sweeper.
type PushServices struct{}

var validPlatforms = map[string]bool{
	constants.PushPlatformIOS:     true,
	constants.PushPlatformAndroid: true,
	constants.PushPlatformWeb:     true,
}

// RegisterToken appends a device token registration. It is idempotent: registering the same token
// twice returns the existing active row, and a token that moved to another actor revokes the
// previous registration first (append-only in both cases).
func (s PushServices) RegisterToken(actorType string, actorID uint, req supportReq.RegisterPushTokenReq) (*supportModel.PushToken, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if actorType == "" || actorID == 0 {
		return nil, response.ForbiddenErr("not authenticated")
	}
	platform := strings.TrimSpace(strings.ToLower(req.Platform))
	if !validPlatforms[platform] {
		return nil, response.ValidationErr("platform must be ios, android or web")
	}
	token := strings.TrimSpace(req.Token)
	if len(token) < 16 || len(token) > 512 {
		return nil, response.ValidationErr("token length must be 16-512")
	}
	hash := sha256Hex(token)

	var existing []supportModel.PushToken
	if err := global.GVA_DB.Where("token_hash = ?", hash).Order("id desc").Limit(1).Find(&existing).Error; err != nil {
		return nil, err
	}
	if len(existing) > 0 && existing[0].RevokedAt == nil {
		if existing[0].ActorType == actorType && existing[0].ActorID == actorID {
			return &existing[0], nil // idempotent re-registration
		}
		if err := revokeTokenRow(token); err != nil {
			return nil, err
		}
	}

	row := supportModel.PushToken{
		ActorType: actorType, ActorID: actorID, Platform: platform,
		Token: token, TokenHash: hash, AppVersion: strings.TrimSpace(req.AppVersion),
	}
	if err := global.GVA_DB.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// RevokeToken appends a revocation row for every active token the actor has on the platform.
func (s PushServices) RevokeToken(actorType string, actorID uint, req supportReq.RevokePushTokenReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	if actorType == "" || actorID == 0 {
		return response.ForbiddenErr("not authenticated")
	}
	platform := strings.TrimSpace(strings.ToLower(req.Platform))
	if !validPlatforms[platform] {
		return response.ValidationErr("platform must be ios, android or web")
	}
	var rows []supportModel.PushToken
	if err := activeTokensQuery().
		Where("actor_type = ? AND actor_id = ? AND platform = ?", actorType, actorID, platform).
		Find(&rows).Error; err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, r := range rows {
		if err := global.GVA_DB.Create(&supportModel.PushToken{
			ActorType: r.ActorType, ActorID: r.ActorID, Platform: r.Platform,
			Token: r.Token, TokenHash: r.TokenHash, AppVersion: r.AppVersion,
			RevokedAt: &now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// DispatchForEvent forwards to the process-wide dispatcher.
func (s PushServices) DispatchForEvent(ctx context.Context, env *supportKafka.Envelope) error {
	return DispatcherInstance().DispatchForEvent(ctx, env)
}

// RetryFailed re-attempts failed dispatches whose attempt budget is not exhausted. It appends a
// new row per attempt; the previous failure is never updated.
func (s PushServices) RetryFailed() error {
	if global.GVA_DB == nil || !Enabled() {
		return nil
	}
	limit := maxAttempts()
	var rows []supportModel.PushDispatch
	if err := global.GVA_DB.Where("status = ?", constants.PushStatusFailed).
		Order("id desc").Limit(500).Find(&rows).Error; err != nil {
		return err
	}
	seen := map[string]bool{}
	d := DispatcherInstance()
	for _, r := range rows {
		if seen[r.IdempotencyKey] {
			continue // only the newest attempt per dispatch is retried
		}
		seen[r.IdempotencyKey] = true
		attempts, settled := dispatchState(r.IdempotencyKey)
		if settled || attempts >= limit || attempts != r.Attempts {
			continue
		}
		participant := Participant{ActorType: r.RecipientType, ActorID: r.RecipientID}
		token, ok := activeTokenValue(participant, r.DeviceTokenHash)
		if !ok {
			continue
		}
		var payload Payload
		if err := json.Unmarshal([]byte(r.Payload), &payload); err != nil {
			continue
		}
		d.sendAttempt(context.Background(), attempt{
			MessageUUID:    r.MessageUUID,
			ConversationID: r.ConversationID,
			TicketID:       r.TicketID,
			Recipient:      participant,
			Token:          token,
			Payload:        payload,
			Attempt:        attempts + 1,
		})
	}
	return nil
}
