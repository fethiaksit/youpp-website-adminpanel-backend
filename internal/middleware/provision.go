package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
)

const (
	ContextProvisionID = "provisionId"
	ContextSetupSiteID = "setupSiteId"
	ContextSetupSlug   = "setupSiteSlug"
)

func ProvisionAPIKeyRequired(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "provision endpoint misconfigured"})
			return
		}

		if c.GetHeader("X-API-Key") != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		c.Next()
	}
}

func SetupTokenRequired(secret string) gin.HandlerFunc {
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
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if !claims.Setup || claims.ProvisionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid setup token"})
			return
		}

		c.Set(ContextProvisionID, claims.ProvisionID)
		c.Set(ContextSetupSiteID, claims.SiteID)
		c.Set(ContextSetupSlug, claims.SiteSlug)
		c.Next()
	}
}
