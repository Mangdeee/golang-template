package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE, UPDATE")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Authorization, Refresh-Token, X-Retry")
		ctx.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Refresh-Token")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")

		if ctx.Request.Method == "OPTIONS" {
			log.Println("OPTIONS")
			ctx.AbortWithStatus(200)
		} else {
			ctx.Next()
		}
	}
}
