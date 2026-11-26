package support

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/service"
	supportpkg "github.com/mo7amed4522/tayara_api/service/support"
	"github.com/mo7amed4522/tayara_api/utils"
)

// ApiGroup groups the support api handlers; embedded into api.ApiGroupApp.
type ApiGroup struct {
	TicketApi
	MessageApi
	AttachmentApi
	PresenceApi
	FaqApi
	AdminApi
	SseClientApi
}

// supportService reaches the support services via the service group var.
var supportService = service.ServiceGroupApp.SupportServiceGroup

// currentActor resolves the authenticated caller from the guard group's context. The client never
// supplies its own actor id.
func currentActor(c *gin.Context) supportpkg.Actor {
	if id := utils.GetBxUserID(c); id > 0 {
		return supportpkg.Actor{Type: constants.ActorCustomer, ID: id}
	}
	if id := utils.GetUserID(c); id > 0 {
		authority := utils.GetUserAuthorityId(c)
		actorType := supportpkg.AuthorityActorType(authority)
		if actorType == "" {
			actorType = constants.ActorAdmin
		}
		return supportpkg.Actor{Type: actorType, ID: id, AuthorityID: authority}
	}
	return supportpkg.Actor{}
}
