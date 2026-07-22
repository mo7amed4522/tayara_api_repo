// Automatically generated template SysOperationRecord
package system

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// If time.Time is contained, please import the time package yourself
type SysOperationRecord struct {
	global.GVA_MODEL
	Ip           string        `json:"ip" form:"ip" gorm:"column:ip;comment:Request ip"`                                     //	 Request ip
	Method       string        `json:"method" form:"method" gorm:"column:method;comment:Request method"`                     // Request method
	Path         string        `json:"path" form:"path" gorm:"column:path;comment:Request path"`                             // Request path
	Status       int           `json:"status" form:"status" gorm:"column:status;comment:Request status"`                     // Request status
	Latency      time.Duration `json:"latency" form:"latency" gorm:"column:latency;comment:Latency" swaggertype:"string"`    // Latency
	Agent        string        `json:"agent" form:"agent" gorm:"type:text;column:agent;comment:Agent"`                       // Agent
	ErrorMessage string        `json:"error_message" form:"error_message" gorm:"column:error_message;comment:Error message"` // Error message
	Body         string        `json:"body" form:"body" gorm:"type:text;column:body;comment:Request body"`                   // Request body
	Resp         string        `json:"resp" form:"resp" gorm:"type:text;column:resp;comment:Response body"`                  // Response body
	UserID       int           `json:"user_id" form:"user_id" gorm:"column:user_id;comment:User id"`                         // User id
	ShortcutKey  int           `json:"shortcutKey" form:"shortcutKey" gorm:"-"`                                              // Shortcut search
}
