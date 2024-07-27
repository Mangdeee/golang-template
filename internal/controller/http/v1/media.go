package v1

import (
	"net/http"
	"strings"
	"time"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/felixlambertv/go-cleanplate/internal/middleware"
	"github.com/felixlambertv/go-cleanplate/internal/service"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type mediaRoutes struct {
	l   logger.Interface
	cfg *config.Config
	ms  service.IMediaService
}

func newMediaRoutes(handler *gin.RouterGroup, l logger.Interface, db *gorm.DB, cfg *config.Config, ms service.IMediaService) {
	r := &mediaRoutes{l: l, cfg: cfg, ms: ms}

	h := handler.Group("media").Use(middleware.Timeout(time.Duration(cfg.App.Timeout) * time.Second))
	{
		h.POST("/upload", r.uploadMedia)
	}
}

func (r *mediaRoutes) uploadMedia(ctx *gin.Context) {
	var req request.MediaUploadRequest

	err := ctx.ShouldBind(&req)
	if err != nil {
		ve := utils.ValidationResponse(err)

		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request invalid",
			Debug:   nil,
			Errors:  ve,
		})
		return
	}

	extension := "." + req.Filename[strings.LastIndex(req.Filename, ".")+1:]
	if err := utils.ValidateExtension(extension); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request invalid",
			Debug:   nil,
			Errors:  err.Error(),
		})
		return
	}

	fileReader, err := utils.ReadRequestFile(req.File)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request invalid",
			Debug:   nil,
			Errors:  err.Error(),
		})
		return
	}

	err = utils.GetReadableFileSize(float64(fileReader.Size()), extension)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "request invalid",
			Debug:   nil,
			Errors:  err.Error(),
		})
		return
	}

	uploadedMediaUrl, err := r.ms.UploadMedia(req, ctx)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, utils.ErrorRes{
			Message: "something went wrong when uploading the media",
			Debug:   nil,
			Errors:  err.Error(),
		})
		return
	}

	response := response.MediaResponse{
		UploadedUrl: uploadedMediaUrl,
	}

	utils.SuccessResponse(ctx, http.StatusOK, utils.SuccessRes{
		Message: "Upload Successful",
		Data:    response,
	})
}
