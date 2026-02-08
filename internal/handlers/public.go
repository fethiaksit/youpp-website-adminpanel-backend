package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PublicHandler struct {
	Sites *mongo.Collection
}

func (h *PublicHandler) GetPublishedSite(c *gin.Context) {
	slug := c.Param("slug")
	var site models.Site
	if err := h.Sites.FindOne(c, bson.M{"slug": slug, "status": "published"}).Decode(&site); err != nil {
		respondError(c, http.StatusNotFound, "site not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slug":    site.Slug,
		"content": site.Content,
		"updated": site.UpdatedAt,
	})
}
