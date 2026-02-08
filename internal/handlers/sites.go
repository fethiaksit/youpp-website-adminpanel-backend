package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/middleware"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SiteHandler struct {
	Sites *mongo.Collection
}

type createSiteRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type updateContentRequest struct {
	Content map[string]interface{} `json:"content" binding:"required"`
}

func (h *SiteHandler) List(c *gin.Context) {
	orgObjectID, err := getOrgID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	cursor, err := h.Sites.Find(c, bson.M{"orgId": orgObjectID})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to fetch sites")
		return
	}
	defer cursor.Close(c)

	var sites []models.Site
	if err := cursor.All(c, &sites); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to decode sites")
		return
	}

	c.JSON(http.StatusOK, sites)
}

func (h *SiteHandler) Create(c *gin.Context) {
	orgObjectID, err := getOrgID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var req createSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	slug := utils.NormalizeSlug(req.Slug)
	if !utils.IsValidSlug(slug) {
		respondError(c, http.StatusBadRequest, "invalid slug")
		return
	}

	site := models.Site{
		OrganizationID: orgObjectID,
		Name:           req.Name,
		Slug:           slug,
		Status:         "draft",
		Content:        map[string]interface{}{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := h.Sites.InsertOne(c, site)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create site")
		return
	}

	site.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, site)
}

func (h *SiteHandler) Get(c *gin.Context) {
	orgObjectID, err := getOrgID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}

	var site models.Site
	if err := h.Sites.FindOne(c, bson.M{"_id": siteID, "orgId": orgObjectID}).Decode(&site); err != nil {
		respondError(c, http.StatusNotFound, "site not found")
		return
	}

	c.JSON(http.StatusOK, site)
}

func (h *SiteHandler) UpdateContent(c *gin.Context) {
	orgObjectID, err := getOrgID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}

	var req updateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"content":   req.Content,
			"updatedAt": time.Now(),
		},
	}

	result, err := h.Sites.UpdateOne(c, bson.M{"_id": siteID, "orgId": orgObjectID}, update)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update content")
		return
	}
	if result.MatchedCount == 0 {
		respondError(c, http.StatusNotFound, "site not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *SiteHandler) Publish(c *gin.Context) {
	h.togglePublish(c, true)
}

func (h *SiteHandler) Unpublish(c *gin.Context) {
	h.togglePublish(c, false)
}

func (h *SiteHandler) togglePublish(c *gin.Context, publish bool) {
	orgObjectID, err := getOrgID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}

	status := "draft"
	var publishedAt *time.Time
	if publish {
		status = "published"
		now := time.Now()
		publishedAt = &now
	}

	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"publishedAt": publishedAt,
			"updatedAt":   time.Now(),
		},
	}

	result, err := h.Sites.UpdateOne(c, bson.M{"_id": siteID, "orgId": orgObjectID}, update)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update status")
		return
	}
	if result.MatchedCount == 0 {
		respondError(c, http.StatusNotFound, "site not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

func getOrgID(c *gin.Context) (primitive.ObjectID, error) {
	orgID, ok := c.Get(middleware.ContextOrgID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("missing org context")
	}

	orgObjectID, err := primitive.ObjectIDFromHex(orgID.(string))
	if err != nil {
		return primitive.NilObjectID, err
	}

	return orgObjectID, nil
}
