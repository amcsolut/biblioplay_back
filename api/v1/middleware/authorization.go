package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoleLevel cria um middleware que exige um nível mínimo de permissão
func RequireRoleLevel(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleLevel, exists := c.Get("role_level")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role level not found"})
			c.Abort()
			return
		}

		userLevel := roleLevel.(int)
		if userLevel < minLevel {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient permissions",
				"required_level": minLevel,
				"user_level":     userLevel,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin é um atalho para RequireRoleLevel(9)
func RequireAdmin() gin.HandlerFunc {
	return RequireRoleLevel(9)
}

