package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
