package handlers

import (
	"net/http"
	"strings"
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
	if err := h.Users.FindOne(c, bson.M{"email": strings.ToLower(req.Email)}).Decode(&user); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, err := utils.CreateToken(user.ID.Hex(), user.GlobalRole, h.Cfg.JWTSecret, time.Duration(h.Cfg.AccessTTLMinutes)*time.Minute)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create access token")
		return
	}

	refreshToken, err := utils.CreateToken(user.ID.Hex(), user.GlobalRole, h.Cfg.JWTRefreshSecret, time.Duration(h.Cfg.RefreshTTLDays)*24*time.Hour)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
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

	accessToken, err := utils.CreateToken(claims.Subject, claims.GlobalRole, h.Cfg.JWTSecret, time.Duration(h.Cfg.AccessTTLMinutes)*time.Minute)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create access token")
		return
	}

	refreshToken, err := utils.CreateToken(claims.Subject, claims.GlobalRole, h.Cfg.JWTRefreshSecret, time.Duration(h.Cfg.RefreshTTLDays)*24*time.Hour)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err.Error())
		return
	}

	globalRole, err := getGlobalRole(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err.Error())
		return
	}

	var user models.User
	if err := h.Users.FindOne(c, bson.M{"_id": userID}).Decode(&user); err != nil {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID.Hex(),
		"email":      user.Email,
		"globalRole": globalRole,
	})
}

func getUserID(c *gin.Context) (primitive.ObjectID, error) {
	userID, ok := c.Get(middleware.ContextUserID)
	if !ok {
		return primitive.NilObjectID, errMissingUserContext
	}
	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		return primitive.NilObjectID, errInvalidUserContext
	}
	return objID, nil
}

func getGlobalRole(c *gin.Context) (string, error) {
	role, ok := c.Get(middleware.ContextGlobalRole)
	if !ok {
		return "", errMissingRoleContext
	}
	return role.(string), nil
}
