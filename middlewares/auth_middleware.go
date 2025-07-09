package middlewares

import (
	"net/http"

	"github.com/dipankarupd/text-editor/utils"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		if ctx.Request.URL.Path == "/refresh" {
			ctx.Next()
			return
		}

		// middleware code:
		token := ctx.Request.Header.Get("token")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
			ctx.Abort()
			return
		}

		claims, msg := utils.ValidateToken(token)
		if msg != "" {

			if msg == "token expired" {
				ctx.JSON(
					http.StatusUnauthorized,
					gin.H{"error": "TOKEN EXPIRED"},
				)
				ctx.Abort()
				return
			}
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			ctx.Abort()
			return
		}

		ctx.Set("name", claims.Name)
		ctx.Set("email", claims.Email)
		ctx.Set("userid", claims.UserId)
	}
}
