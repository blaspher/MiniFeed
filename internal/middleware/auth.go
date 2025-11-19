package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	jwtUtil "minifeed/pkg/jwt"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "missing or invalid token",
			})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		claims, err := jwtUtil.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)

		c.Next()

	}
}
