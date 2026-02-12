package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	Users           *mongo.Collection
	Sites           *mongo.Collection
	SitePermissions *mongo.Collection
}

type grantSiteRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Role            string `json:"role" binding:"required"`
	CreateIfMissing bool   `json:"createIfMissing"`
}

type createUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	GlobalRole string `json:"globalRole"`
}

func (h *AdminHandler) ListSites(c *gin.Context) {
	(&SiteHandler{Sites: h.Sites, SitePermissions: h.SitePermissions}).List(c)
}
func (h *AdminHandler) CreateSite(c *gin.Context) {
	(&SiteHandler{Sites: h.Sites, SitePermissions: h.SitePermissions}).Create(c)
}

func (h *AdminHandler) GrantSiteAccess(c *gin.Context) {
	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}

	var req grantSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role != "owner" && role != "editor" && role != "viewer" {
		respondError(c, http.StatusBadRequest, "invalid role")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	var user models.User
	err = h.Users.FindOne(c, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			respondError(c, http.StatusInternalServerError, "failed to query user")
			return
		}
		if !req.CreateIfMissing {
			respondError(c, http.StatusNotFound, "user not found")
			return
		}
		now := time.Now().UTC()
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(primitive.NewObjectID().Hex()), bcrypt.DefaultCost)
		user = models.User{Email: email, PasswordHash: string(passwordHash), GlobalRole: "user", CreatedAt: now, UpdatedAt: now}
		res, insertErr := h.Users.InsertOne(c, user)
		if insertErr != nil {
			respondError(c, http.StatusInternalServerError, "failed to create user")
			return
		}
		user.ID = res.InsertedID.(primitive.ObjectID)
	}

	now := time.Now().UTC()
	_, err = h.SitePermissions.UpdateOne(c,
		bson.M{"siteId": siteID, "userId": user.ID},
		bson.M{"$set": bson.M{"role": role, "updatedAt": now}, "$setOnInsert": bson.M{"createdAt": now}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to grant access")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "granted"})
}

func (h *AdminHandler) ListSiteUsers(c *gin.Context) {
	siteID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid site id")
		return
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"siteId": siteID}}},
		{{Key: "$lookup", Value: bson.M{"from": "users", "localField": "userId", "foreignField": "_id", "as": "user"}}},
		{{Key: "$unwind", Value: "$user"}},
		{{Key: "$project", Value: bson.M{"_id": 0, "userId": "$user._id", "email": "$user.email", "globalRole": "$user.globalRole", "role": "$role"}}},
	}
	cursor, err := h.SitePermissions.Aggregate(c, pipeline)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to fetch site users")
		return
	}
	defer cursor.Close(c)
	var out []bson.M
	if err := cursor.All(c, &out); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to decode site users")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}
	role := strings.ToLower(strings.TrimSpace(req.GlobalRole))
	if role == "" {
		role = "user"
	}
	if role != "user" && role != "superadmin" {
		respondError(c, http.StatusBadRequest, "invalid globalRole")
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}
	now := time.Now().UTC()
	user := models.User{Email: strings.ToLower(req.Email), PasswordHash: string(passwordHash), GlobalRole: role, CreatedAt: now, UpdatedAt: now}
	res, err := h.Users.InsertOne(c, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondError(c, http.StatusConflict, "email already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create user")
		return
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	user.PasswordHash = ""
	c.JSON(http.StatusCreated, user)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	cursor, err := h.Users.Find(c, bson.M{})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	defer cursor.Close(c)
	var users []models.User
	if err := cursor.All(c, &users); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to decode users")
		return
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) CreateSiteDirect(c *gin.Context) {
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
	res, err := h.Sites.InsertOne(c, site)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondError(c, http.StatusConflict, "slug already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create site")
		return
	}
	site.ID = res.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, site)
}
