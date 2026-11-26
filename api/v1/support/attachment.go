package support

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

// AttachmentApi exposes uploads and the gated download.
type AttachmentApi struct{}

// Upload
//
//	@Tags			Support
//	@Summary		Upload an attachment for a conversation (validated before storage)
//	@Accept			multipart/form-data
//	@Produce		application/json
//	@Param			conversationId	formData	int		true	"conversation id"
//	@Param			file			formData	file	true	"file"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/support/attachments [post]
func (AttachmentApi) Upload(c *gin.Context) {
	convID, err := strconv.ParseUint(c.PostForm("conversationId"), 10, 64)
	if err != nil || convID == 0 {
		response.FailWithError(response.ValidationErr("conversationId is required"), c)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(response.ValidationErr("file is required"), c)
		return
	}
	att, err := supportService.AttachmentServices.Upload(currentActor(c), uint(convID), file)
	if err != nil {
		respondErr(err, c)
		return
	}
	response.OkWithData(map[string]any{
		"attachmentUuid": att.AttachmentUUID,
		"kind":           att.Kind,
		"mimeType":       att.MimeType,
		"sizeBytes":      att.SizeBytes,
		"originalName":   att.OriginalName,
		"width":          att.Width,
		"height":         att.Height,
	}, c)
}

// Download
//
//	@Tags		Support
//	@Summary	Download an attachment the caller may read
//	@Produce	application/octet-stream
//	@Param		uuid	path		string	true	"attachment uuid"
//	@Success	200		{file}		binary
//	@Router		/api/v1/support/attachments/{uuid} [get]
func (AttachmentApi) Download(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		response.FailWithError(response.ValidationErr("attachment uuid is required"), c)
		return
	}
	att, err := supportService.AttachmentServices.ResolveForDownload(currentActor(c), uuid)
	if err != nil {
		respondErr(err, c)
		return
	}
	if localPath, ok := supportService.AttachmentServices.LocalPath(att); ok {
		c.File(localPath)
		return
	}
	// Remote backends cannot be streamed through the OSS interface: hand back the stored URL.
	c.Redirect(302, att.StorageKey)
}
