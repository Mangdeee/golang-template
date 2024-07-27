package v1

import (
	"fmt"
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"

	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/di"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/middleware"

	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(handler *gin.Engine, l logger.Interface, db *gorm.DB, cfg *config.Config, di *di.DependencyInjection) {
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())
	handler.Use(middleware.CORSMiddleware())
	if cfg.App.Env != "local" {
		handler.Use(sentrygin.New(sentrygin.Options{}))
	}

	handler.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "oks"})
	})

	handler.GET("/app/reset-password/:token", func(context *gin.Context) {
		var req request.ResetPasswordRedirectRequest

		if err := context.ShouldBindUri(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		deeplinkUrl := fmt.Sprintf("%sreset-password?token=%s", cfg.App.DeeplinkUrl, req.ResetToken)
		http.Redirect(context.Writer, context.Request, deeplinkUrl, http.StatusSeeOther)
	})

	h := handler.Group("api/v1")
	{
		newAuthRoutes(h, l, cfg, di.AuthService, di.MailService)
		newUserRoutes(h, l, db, di.UserService, cfg)
		newMediaRoutes(h, l, db, cfg, di.MediaService)
	}
}
