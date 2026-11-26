package support

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"gorm.io/gorm"
)

// ErrAttachmentImmutable is returned when an attachment row is mutated outside the two legal
// transitions (uploaded -> attached, uploaded -> rejected).
var ErrAttachmentImmutable = errors.New("support_attachment only permits the one-time link")

// Attachment metadata for an uploaded file. Kafka never carries bytes; clients fetch the object
// through the gated download endpoint keyed by attachment_uuid.
type Attachment struct {
	ID             uint      `json:"id" gorm:"primarykey;comment:ID"`
	AttachmentUUID string    `json:"attachmentUuid" form:"attachmentUuid" gorm:"size:36;uniqueIndex;comment:UUIDv7"`
	ConversationID uint      `json:"conversationId" form:"conversationId" gorm:"index;comment:Conversation ID"`
	TicketID       uint      `json:"ticketId" form:"ticketId" gorm:"index;comment:Ticket ID (denormalized)"`
	MessageID      *uint     `json:"messageId" form:"messageId" gorm:"index;comment:NULL until linked to a message"`
	UploaderType   string    `json:"uploaderType" form:"uploaderType" gorm:"size:16;comment:customer|support|quality|admin"`
	UploaderID     uint      `json:"uploaderId" form:"uploaderId" gorm:"index;comment:Uploader ID"`
	Kind           string    `json:"kind" form:"kind" gorm:"size:16;comment:image|file"`
	MimeType       string    `json:"mimeType" form:"mimeType" gorm:"size:128;comment:Sniffed MIME type"`
	SizeBytes      int64     `json:"sizeBytes" form:"sizeBytes" gorm:"comment:File size in bytes"`
	OriginalName   string    `json:"originalName" form:"originalName" gorm:"size:255;comment:Sanitized original filename"`
	StorageKey     string    `json:"storageKey" form:"storageKey" gorm:"size:512;comment:First return of UploadFileWithPath"`
	StorageName    string    `json:"storageName" form:"storageName" gorm:"size:255;comment:Second return of UploadFileWithPath"`
	Sha256         string    `json:"sha256" form:"sha256" gorm:"size:64;comment:Content hash"`
	Width          *int      `json:"width" form:"width" gorm:"comment:Image width (images only)"`
	Height         *int      `json:"height" form:"height" gorm:"comment:Image height (images only)"`
	ThumbnailKey   *string   `json:"thumbnailKey" form:"thumbnailKey" gorm:"size:512;comment:Reserved; always NULL in v1"`
	Status         string    `json:"status" form:"status" gorm:"size:16;index;comment:uploaded|attached|rejected"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (Attachment) TableName() string { return "support_attachment" }

// BeforeUpdate enforces the two-transition rule by comparing the persisted row with the values
// GORM is about to write. Everything except message_id and status must be unchanged.
func (a *Attachment) BeforeUpdate(tx *gorm.DB) error {
	if a.ID == 0 {
		return ErrAttachmentImmutable
	}
	var old Attachment
	if err := tx.Session(&gorm.Session{NewDB: true}).First(&old, a.ID).Error; err != nil {
		return err
	}

	newMessageID := a.MessageID
	newStatus := a.Status
	switch dest := tx.Statement.Dest.(type) {
	case *Attachment:
		newMessageID, newStatus = dest.MessageID, dest.Status
	case Attachment:
		newMessageID, newStatus = dest.MessageID, dest.Status
	case map[string]interface{}:
		if v, ok := dest["message_id"]; ok {
			if v == nil {
				newMessageID = nil
			} else if id, ok := v.(uint); ok {
				newMessageID = &id
			}
		}
		if v, ok := dest["status"].(string); ok {
			newStatus = v
		}
	}

	linked := old.MessageID == nil && newMessageID != nil &&
		old.Status == constants.SupportAttachmentStatusUploaded && newStatus == constants.SupportAttachmentStatusAttached
	rejected := old.MessageID == nil && newMessageID == nil &&
		old.Status == constants.SupportAttachmentStatusUploaded && newStatus == constants.SupportAttachmentStatusRejected
	if !linked && !rejected {
		return ErrAttachmentImmutable
	}
	return nil
}

// BeforeDelete blocks removal: attachment metadata is append-only.
func (Attachment) BeforeDelete(tx *gorm.DB) error {
	return errors.New("support_attachment is append-only")
}
