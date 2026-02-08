package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/middleware"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Users *mongo.Collection
	Cfg   *config.Config
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	var user models.User
	if err := h.Users.FindOne(c, bson.M{"email": req.Email}).Decode(&user); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, err := utils.CreateToken(user.ID.Hex(), user.OrganizationID.Hex(), user.Role, h.Cfg.JWTSecret, time.Duration(h.Cfg.AccessTTLMinutes)*time.Minute)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create access token")
		return
	}

	refreshToken, err := utils.CreateToken(user.ID.Hex(), user.OrganizationID.Hex(), user.Role, h.Cfg.JWTRefreshSecret, time.Duration(h.Cfg.RefreshTTLDays)*24*time.Hour)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	claims, err := utils.ParseToken(req.RefreshToken, h.Cfg.JWTRefreshSecret)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	accessToken, err := utils.CreateToken(claims.UserID, claims.OrgID, claims.Role, h.Cfg.JWTSecret, time.Duration(h.Cfg.AccessTTLMinutes)*time.Minute)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create access token")
		return
	}

	refreshToken, err := utils.CreateToken(claims.UserID, claims.OrgID, claims.Role, h.Cfg.JWTRefreshSecret, time.Duration(h.Cfg.RefreshTTLDays)*24*time.Hour)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	orgID, ok := c.Get(middleware.ContextOrgID)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing org context")
		return
	}

	userID, ok := c.Get(middleware.ContextUserID)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing user context")
		return
	}

	orgObjectID, err := primitive.ObjectIDFromHex(orgID.(string))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid org id")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var user models.User
	if err := h.Users.FindOne(c, bson.M{"_id": userObjectID, "orgId": orgObjectID}).Decode(&user); err != nil {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID.Hex(),
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
		"orgId": user.OrganizationID.Hex(),
	})
}
