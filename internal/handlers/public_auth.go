package handlers

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type PublicAuthHandler struct {
	Users           *mongo.Collection
	Sites           *mongo.Collection
	SitePermissions *mongo.Collection
	Cfg             *config.Config
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *PublicAuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		respondError(c, http.StatusBadRequest, "invalid email")
		return
	}
	if len(req.Password) < 8 {
		respondError(c, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	now := time.Now().UTC()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		GlobalRole:   "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	userResult, err := h.Users.InsertOne(c, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondError(c, http.StatusConflict, "email already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create user")
		return
	}
	userID := userResult.InsertedID.(primitive.ObjectID)

	baseName := strings.Split(email, "@")[0]
	baseSlug := utils.NormalizeSlug(baseName)
	if baseSlug == "" {
		baseSlug = "site"
	}

	siteSlug, err := h.resolveUniqueSlug(c, baseSlug)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create site")
		return
	}

	site := models.Site{
		Name:   fmt.Sprintf("%s Site", titleize(baseName)),
		Slug:   siteSlug,
		Status: "draft",
		Content: map[string]interface{}{
			"sections": []map[string]interface{}{
				{"type": "hero", "data": map[string]interface{}{"title": "Hoş geldin", "subtitle": "Siten hazır"}},
				{"type": "cta", "data": map[string]interface{}{"title": "İletişim", "buttonText": "Teklif Al", "buttonHref": "#contact"}},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	siteResult, err := h.Sites.InsertOne(c, site)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create site")
		return
	}
	siteID := siteResult.InsertedID.(primitive.ObjectID)

	permission := models.SitePermission{SiteID: siteID, UserID: userID, Role: "owner", CreatedAt: now, UpdatedAt: now}
	if _, err = h.SitePermissions.InsertOne(c, permission); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create site permission")
		return
	}

	accessToken, err := utils.CreateToken(userID.Hex(), "user", h.Cfg.JWTSecret, time.Duration(h.Cfg.AccessTTLMinutes)*time.Minute)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create access token")
		return
	}
	refreshToken, err := utils.CreateToken(userID.Hex(), "user", h.Cfg.JWTRefreshSecret, time.Duration(h.Cfg.RefreshTTLDays)*24*time.Hour)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"site":         gin.H{"id": siteID.Hex(), "slug": siteSlug, "name": site.Name, "status": "draft"},
	})
}

func (h *PublicAuthHandler) resolveUniqueSlug(c *gin.Context, baseSlug string) (string, error) {
	for i := 0; i <= 200; i++ {
		slug := baseSlug
		if i > 0 {
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
		}
		count, err := h.Sites.CountDocuments(c, bson.M{"slug": slug})
		if err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
	}
	return "", fmt.Errorf("unable to resolve unique slug")
}

func titleize(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return "My"
	}
	return strings.ToUpper(clean[:1]) + clean[1:]
}
