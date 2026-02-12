package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SiteHandler struct {
	Sites           *mongo.Collection
	SitePermissions *mongo.Collection
}

type createSiteRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type updateContentRequest struct {
	Content map[string]interface{} `json:"content" binding:"required"`
}

func (h *SiteHandler) List(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err.Error())
		return
	}
	globalRole, _ := getGlobalRole(c)

	filter := bson.M{}
	if globalRole != "superadmin" {
		siteIDs, err := h.getPermittedSiteIDs(c, userID)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to fetch permissions")
			return
		}
		if len(siteIDs) == 0 {
			c.JSON(http.StatusOK, []models.Site{})
			return
		}
		filter = bson.M{"_id": bson.M{"$in": siteIDs}}
	}

	cursor, err := h.Sites.Find(c, filter)
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
	globalRole, _ := getGlobalRole(c)
	if globalRole != "superadmin" {
		respondError(c, http.StatusForbidden, "superadmin access required")
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

	now := time.Now().UTC()
	site := models.Site{Name: req.Name, Slug: slug, Status: "draft", Content: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now}
	result, err := h.Sites.InsertOne(c, site)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondError(c, http.StatusConflict, "slug already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create site")
		return
	}
	site.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, site)
}

func (h *SiteHandler) Get(c *gin.Context) {
	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}
	allowed, err := h.canReadCurrentUser(c, siteID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to check permission")
		return
	}
	if !allowed {
		respondError(c, http.StatusForbidden, "no access to site")
		return
	}

	var site models.Site
	if err := h.Sites.FindOne(c, bson.M{"_id": siteID}).Decode(&site); err != nil {
		respondError(c, http.StatusNotFound, "site not found")
		return
	}
	c.JSON(http.StatusOK, site)
}

func (h *SiteHandler) UpdateContent(c *gin.Context) {
	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}
	allowed, err := h.canWriteCurrentUser(c, siteID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to check permission")
		return
	}
	if !allowed {
		respondError(c, http.StatusForbidden, "write access required")
		return
	}

	var req updateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.Sites.UpdateOne(c, bson.M{"_id": siteID}, bson.M{"$set": bson.M{"content": req.Content, "updatedAt": time.Now().UTC()}})
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

func (h *SiteHandler) Publish(c *gin.Context)   { h.togglePublish(c, true) }
func (h *SiteHandler) Unpublish(c *gin.Context) { h.togglePublish(c, false) }

func (h *SiteHandler) togglePublish(c *gin.Context, publish bool) {
	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}
	allowed, err := h.canWriteCurrentUser(c, siteID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to check permission")
		return
	}
	if !allowed {
		respondError(c, http.StatusForbidden, "write access required")
		return
	}

	now := time.Now().UTC()
	status := "draft"
	var publishedAt *time.Time
	if publish {
		status = "published"
		publishedAt = &now
	}

	result, err := h.Sites.UpdateOne(c, bson.M{"_id": siteID}, bson.M{"$set": bson.M{"status": status, "publishedAt": publishedAt, "updatedAt": now}})
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

func (h *SiteHandler) canReadCurrentUser(c *gin.Context, siteID primitive.ObjectID) (bool, error) {
	userID, err := getUserID(c)
	if err != nil {
		return false, err
	}
	role, _ := getGlobalRole(c)
	if role == "superadmin" {
		return true, nil
	}
	return h.canReadSite(c, userID, siteID)
}

func (h *SiteHandler) canWriteCurrentUser(c *gin.Context, siteID primitive.ObjectID) (bool, error) {
	userID, err := getUserID(c)
	if err != nil {
		return false, err
	}
	role, _ := getGlobalRole(c)
	if role == "superadmin" {
		return true, nil
	}
	return h.canWriteSite(c, userID, siteID)
}

func (h *SiteHandler) canReadSite(c *gin.Context, userID, siteID primitive.ObjectID) (bool, error) {
	count, err := h.SitePermissions.CountDocuments(c, bson.M{"userId": userID, "siteId": siteID, "role": bson.M{"$in": []string{"viewer", "editor", "owner"}}})
	return count > 0, err
}

func (h *SiteHandler) canWriteSite(c *gin.Context, userID, siteID primitive.ObjectID) (bool, error) {
	count, err := h.SitePermissions.CountDocuments(c, bson.M{"userId": userID, "siteId": siteID, "role": bson.M{"$in": []string{"editor", "owner"}}})
	return count > 0, err
}

func (h *SiteHandler) getPermittedSiteIDs(c *gin.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	cursor, err := h.SitePermissions.Find(c, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var perms []models.SitePermission
	if err := cursor.All(c, &perms); err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, 0, len(perms))
	for _, p := range perms {
		ids = append(ids, p.SiteID)
	}
	return ids, nil
}
