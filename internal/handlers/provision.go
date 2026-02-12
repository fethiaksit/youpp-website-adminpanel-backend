package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type ProvisionHandler struct {
	Cfg   *config.Config
	Users *mongo.Collection
}

func (h *ProvisionHandler) Bootstrap(c *gin.Context) {
	email := strings.ToLower(strings.TrimSpace(h.Cfg.SuperAdminEmail))
	password := h.Cfg.SuperAdminPassword
	if email == "" || password == "" {
		respondError(c, http.StatusBadRequest, "SUPERADMIN_EMAIL and SUPERADMIN_PASSWORD are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}
	now := time.Now().UTC()
	_, err = h.Users.UpdateOne(c,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"passwordHash": string(hash), "globalRole": "superadmin", "updatedAt": now}, "$setOnInsert": bson.M{"createdAt": now}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to bootstrap superadmin")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "bootstrapped"})
}
