package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

func extractToken(c *gin.Context) (string, error) {
	bearerToken := c.Request.Header.Get("Authorization")
	err := errors.New("no Authorization token detected")

	// Apple already reserved header for Authorization
	// https://developer.apple.com/documentation/foundation/nsurlrequest
	if bearerToken == "" {
		bearerToken = c.Request.Header.Get("X-Authorization")
	}

	if len(strings.Split(bearerToken, " ")) == 2 {
		bearerToken = strings.Split(bearerToken, " ")[1]
	}

	if bearerToken == "" {
		return "", err
	}

	return bearerToken, nil
}

func JWTAuthMiddleware(cfg *config.Config, allowedLevel ...uint) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		extractedToken, err := extractToken(ctx)
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
				Message: "Invalid token",
				Debug:   err,
				Errors:  err.Error(),
			})
			ctx.Abort()
			return
		}

		parsedToken, err := utils.ParseToken(extractedToken, cfg.App.Secret)
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
				Message: "Invalid token",
				Debug:   err,
				Errors:  err.Error(),
			})
			ctx.Abort()
			return
		}

		if !slices.Contains(allowedLevel, parsedToken.User.UserLevel) || (time.Now().Unix() >= parsedToken.Expire) {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
				Message: "Invalid token",
				Debug:   nil,
				Errors:  "You're not authorized to access this",
			})
			ctx.Abort()
			return
		}

		if !utils.CheckWhitelistUrl(ctx.Request.URL.Path) {
			if parsedToken.User.ConfirmedAt == (time.Time{}) && !strings.Contains(ctx.Request.URL.Path, "verify") {
				utils.ErrorResponse(ctx, http.StatusUnauthorized, utils.ErrorRes{
					Message: "Invalid token",
					Debug:   nil,
					Errors:  "This account is not verified",
				})
				ctx.Abort()
				return
			}
		}

		ctx.Set("user", *parsedToken.User)
		ctx.Next()
	}
}
