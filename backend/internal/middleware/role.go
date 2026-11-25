package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "role tidak ditemukan",
			})
			c.Abort()
			return
		}

		userRole, ok := value.(string)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "role tidak valid",
			})
			c.Abort()
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "kamu tidak memiliki akses ke resource ini",
		})

		c.Abort()
	}
}
