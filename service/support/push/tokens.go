package push

import (
	"errors"
	"strings"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	engageModel "github.com/mo7amed4522/tayara_api/model/engage"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
)

// engageUserID maps a support actor onto the user id that model/engage's DeviceToken stores.
//
// Only customers share that id space (engage_device_token.user_id is a bx_user_profile.id, and a
// support customer's Actor.ID is the same row). Staff ids are sys_admin_profile ids, which can
// coincide numerically with a customer id, so they are deliberately never looked up here — a
// wrong match would deliver one person's support chat to another person.
func engageUserID(actorType string, actorID uint) (uint, bool) {
	if actorType != constants.ActorCustomer || actorID == 0 {
		return 0, false
	}
	return actorID, true
}

// activeDeviceTokens returns every token a participant can be reached on: the support
// registrations (append-only, all actor types) plus the device tokens the engage module already
// owns for customers. Duplicates are dropped; first occurrence wins.
func activeDeviceTokens(p Participant) ([]string, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	out := make([]string, 0, 2)
	seen := map[string]bool{}
	appendToken := func(tok string) {
		tok = strings.TrimSpace(tok)
		if tok == "" || seen[tok] {
			return
		}
		seen[tok] = true
		out = append(out, tok)
	}

	var rows []supportModel.PushToken
	if err := activeTokensQuery().
		Where("actor_type = ? AND actor_id = ?", p.ActorType, p.ActorID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		appendToken(r.Token)
	}

	if uid, ok := engageUserID(p.ActorType, p.ActorID); ok {
		var devices []engageModel.DeviceToken
		if err := global.GVA_DB.Where("user_id = ? AND is_active = ?", uid, true).
			Find(&devices).Error; err != nil {
			return nil, err
		}
		for _, d := range devices {
			appendToken(d.Token)
		}
	}
	return out, nil
}

// activeTokenValue resolves a previously dispatched token hash back to the token value. Only the
// hash is stored on a dispatch row, and the token may have come from either source.
func activeTokenValue(p Participant, tokenHash string) (string, bool) {
	if tokenHash == "" {
		return "", false
	}
	tokens, err := activeDeviceTokens(p)
	if err != nil {
		return "", false
	}
	for _, t := range tokens {
		if sha256Hex(t) == tokenHash {
			return t, true
		}
	}
	return "", false
}
