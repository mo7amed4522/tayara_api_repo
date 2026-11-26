package support

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/upload"
	"gorm.io/gorm"
)

// extMime is the storage layer's hard extension allowlist mapped to its canonical MIME type.
// It mirrors utils/upload/local.go's allowedUploadExts — we adapt to it, we do not change it.
var extMime = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

// AttachmentServices validates uploads, stores them through the existing OSS interface and owns
// the one-time link to a message plus the orphan sweeper.
type AttachmentServices struct{}

// Upload validates before a single byte is written, then stores via
// upload.NewOss().UploadFileWithPath(file, "support-chat").
func (s AttachmentServices) Upload(actor Actor, conversationID uint, file *multipart.FileHeader) (*supportModel.Attachment, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	settings := SettingsServices{}
	if !settings.AttachmentsEnabled() {
		return nil, response.ConflictErr("attachments are disabled")
	}
	access, ok, err := (AuthorizationServices{}).CanWrite(actor, conversationID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, response.ForbiddenErr("not allowed to upload into this conversation")
	}
	if access.Conversation.Status != constants.SupportConversationStatusActive {
		return nil, response.ConflictErr("conversation is closed")
	}
	if file == nil || file.Size <= 0 {
		return nil, response.ValidationErr("file is required")
	}
	if file.Size > settings.MaxFileBytes() {
		return nil, response.ValidationErr("file exceeds the maximum size")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	expected, known := extMime[ext]
	if !known {
		return nil, response.ValidationErr("file extension is not allowed")
	}
	allowed := settings.AllowedMime()
	if !containsString(allowed, expected) {
		return nil, response.ValidationErr("file type is not allowed")
	}

	sniffed, sha, width, height, err := inspectUpload(file)
	if err != nil {
		return nil, response.ValidationErr("file could not be read")
	}
	if !mimeAgrees(expected, sniffed) {
		return nil, response.ValidationErr("file content does not match its extension")
	}

	publicURL, storageName, err := upload.NewOss().UploadFileWithPath(file, "support-chat")
	if err != nil {
		return nil, response.InternalErr("failed to store attachment: " + err.Error())
	}

	kind := constants.SupportAttachmentKindFile
	if strings.HasPrefix(sniffed, "image/") {
		kind = constants.SupportAttachmentKindImage
	}
	att := supportModel.Attachment{
		AttachmentUUID: utils.NewUUIDv7(),
		ConversationID: conversationID,
		TicketID:       access.Conversation.TicketID,
		UploaderType:   senderTypeForActor(actor),
		UploaderID:     actor.ID,
		Kind:           kind,
		MimeType:       sniffed,
		SizeBytes:      file.Size,
		OriginalName:   utils.SanitizeFilename(file.Filename),
		StorageKey:     publicURL,
		StorageName:    storageName,
		Sha256:         sha,
		Width:          width,
		Height:         height,
		Status:         constants.SupportAttachmentStatusUploaded,
	}
	if err := global.GVA_DB.Create(&att).Error; err != nil {
		return nil, err
	}
	return &att, nil
}

// inspectUpload sniffs the content type, hashes the bytes and (for images) decodes dimensions.
func inspectUpload(file *multipart.FileHeader) (string, string, *int, *int, error) {
	f, err := file.Open()
	if err != nil {
		return "", "", nil, nil, err
	}
	defer func() { _ = f.Close() }()

	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", "", nil, nil, err
	}
	sniffed := http.DetectContentType(head[:n])

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", "", nil, nil, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", "", nil, nil, err
	}
	sum := hex.EncodeToString(h.Sum(nil))

	var width, height *int
	if strings.HasPrefix(sniffed, "image/") && sniffed != "image/webp" {
		if _, err := f.Seek(0, io.SeekStart); err == nil {
			if cfg, _, derr := image.DecodeConfig(f); derr == nil {
				w, ht := cfg.Width, cfg.Height
				width, height = &w, &ht
			}
		}
	}
	return sniffed, sum, width, height, nil
}

// mimeAgrees checks that the sniffed type matches the extension's canonical type. Some platforms
// report generic types for documents, so a small compatibility window is allowed.
func mimeAgrees(expected, sniffed string) bool {
	if expected == sniffed {
		return true
	}
	switch expected {
	case "application/pdf":
		return sniffed == "application/pdf"
	case "application/msword":
		return sniffed == "application/msword" || sniffed == "application/CDFV2" || sniffed == "application/octet-stream"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return sniffed == "application/zip" || sniffed == "application/octet-stream"
	}
	return false
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// claim validates the attachment uuids a message intends to attach and returns the rows. It does
// not link them; linkToMessage performs the single legal promotion inside the same transaction.
func (s AttachmentServices) claim(tx *gorm.DB, actor Actor, conv *supportModel.Conversation, uuids []string) ([]supportModel.Attachment, error) {
	if len(uuids) == 0 {
		return nil, nil
	}
	settings := SettingsServices{}
	if !settings.AttachmentsEnabled() {
		return nil, response.ConflictErr("attachments are disabled")
	}
	seen := map[string]bool{}
	list := make([]string, 0, len(uuids))
	for _, u := range uuids {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		list = append(list, u)
	}
	if len(list) == 0 {
		return nil, response.ValidationErr("attachmentUuids is empty")
	}
	if len(list) > settings.MaxFilesPerMessage() {
		return nil, response.ValidationErr("too many attachments")
	}
	var rows []supportModel.Attachment
	if err := tx.Where("attachment_uuid IN ? AND status = ?", list, constants.SupportAttachmentStatusUploaded).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != len(list) {
		return nil, response.ValidationErr("one or more attachments are unavailable")
	}
	var total int64
	for _, a := range rows {
		if a.ConversationID != conv.ID {
			return nil, response.ForbiddenErr("attachment belongs to another conversation")
		}
		if a.UploaderType != senderTypeForActor(actor) || a.UploaderID != actor.ID {
			return nil, response.ForbiddenErr("attachment was uploaded by another actor")
		}
		total += a.SizeBytes
	}
	if total > settings.MaxTotalBytesPerMessage() {
		return nil, response.ValidationErr("attachments exceed the per-message size limit")
	}
	return rows, nil
}

// linkToMessage performs the uploaded -> attached promotion. This is the only UPDATE the
// attachment table ever permits.
func (s AttachmentServices) linkToMessage(tx *gorm.DB, ids []uint, messageID uint) error {
	for _, id := range ids {
		var att supportModel.Attachment
		if err := tx.First(&att, id).Error; err != nil {
			return err
		}
		att.MessageID = &messageID
		att.Status = constants.SupportAttachmentStatusAttached
		if err := tx.Save(&att).Error; err != nil {
			return err
		}
	}
	return nil
}

// ResolveForDownload authorizes a download by UUID and re-checks conversation access.
func (s AttachmentServices) ResolveForDownload(actor Actor, attachmentUUID string) (*supportModel.Attachment, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var att supportModel.Attachment
	if err := global.GVA_DB.Where("attachment_uuid = ?", attachmentUUID).First(&att).Error; err != nil {
		return nil, response.NotFoundErr("attachment not found")
	}
	if att.Status == constants.SupportAttachmentStatusRejected {
		return nil, response.NotFoundErr("attachment not found")
	}
	if _, ok, err := (AuthorizationServices{}).CanRead(actor, att.ConversationID); err != nil {
		return nil, err
	} else if !ok {
		return nil, response.ForbiddenErr("not allowed to read this attachment")
	}
	return &att, nil
}

// LocalPath returns the on-disk path for a locally stored attachment. Remote backends cannot be
// streamed through the OSS interface, so the handler redirects to StorageKey instead.
func (s AttachmentServices) LocalPath(att *supportModel.Attachment) (string, bool) {
	if strings.ToLower(global.GVA_CONFIG.System.OssType) != "local" {
		return "", false
	}
	if att.StorageName == "" {
		return "", false
	}
	return filepath.Join(global.GVA_CONFIG.Local.StorePath, "support-chat", att.StorageName), true
}

// OrphanCleanup marks old, never-linked uploads as rejected. The file itself is left in place:
// the OSS interface's DeleteFile cannot address a customPath upload.
func (s AttachmentServices) OrphanCleanup() error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	hours := global.GVA_CONFIG.Support.Chat.Attachments.OrphanCleanupHours
	if hours <= 0 {
		return nil
	}
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	var rows []supportModel.Attachment
	if err := global.GVA_DB.Where("status = ? AND created_at < ? AND message_id IS NULL",
		constants.SupportAttachmentStatusUploaded, cutoff).Limit(500).Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		rows[i].Status = constants.SupportAttachmentStatusRejected
		rows[i].MessageID = nil
		if err := global.GVA_DB.Save(&rows[i]).Error; err != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Warn("support orphan attachment sweep failed")
			}
		}
	}
	return nil
}
