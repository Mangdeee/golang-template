package request

import "mime/multipart"

type (
	MediaUploadRequest struct {
		Filename string                `form:"filename" binding:"required" example:"scenarioImage"`
		File     *multipart.FileHeader `form:"file" binding:"required,file"`
	}
)
