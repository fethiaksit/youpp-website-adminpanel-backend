package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
)

const (
	ContextUserID     = "userId"
	ContextGlobalRole = "globalRole"
)

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := utils.ParseToken(parts[1], secret)
		if err != nil || claims.Subject == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(ContextUserID, claims.Subject)
		c.Set(ContextGlobalRole, claims.GlobalRole)
		c.Next()
	}
}

func SuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get(ContextGlobalRole)
		if !ok || role.(string) != "superadmin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "superadmin access required"})
			return
		}
		c.Next()
	}
}
