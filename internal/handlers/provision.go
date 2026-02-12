package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	provisionCodeTTL = 15 * time.Minute
	setupTokenTTL    = 15 * time.Minute
)

type ProvisionHandler struct {
	Cfg            *config.Config
	ProvisionCodes *mongo.Collection
	Organizations  *mongo.Collection
	Sites          *mongo.Collection
	Tenants        *mongo.Collection
	Users          *mongo.Collection
}

type requestCodeRequest struct {
	SiteName string `json:"siteName" binding:"required"`
	SiteSlug string `json:"siteSlug" binding:"required"`
}

type setupLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type setupRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name"`
}

type conflictError struct{ message string }

func (e conflictError) Error() string { return e.message }

func (h *ProvisionHandler) RequestCode(c *gin.Context) {
	var req requestCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	slug := utils.NormalizeSlug(req.SiteSlug)
	if !utils.IsValidSlug(slug) {
		respondError(c, http.StatusBadRequest, "invalid site slug")
		return
	}

	existingSiteCount, err := h.Sites.CountDocuments(c, bson.M{"slug": slug})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to validate site slug")
		return
	}
	if existingSiteCount > 0 {
		respondError(c, http.StatusConflict, "site slug already exists")
		return
	}

	code, err := generateHumanCode()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to generate setup code")
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(provisionCodeTTL)
	provisionCode := models.ProvisionCode{
		CodeHash:  hashCode(code),
		ExpiresAt: expiresAt,
		Payload: models.ProvisionCodePayload{
			SiteName: req.SiteName,
			SiteSlug: slug,
		},
		CreatedAt: now,
	}

	if _, err := h.ProvisionCodes.InsertOne(c, provisionCode); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create provision code")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": code, "expiresAt": expiresAt})
}

func (h *ProvisionHandler) SetupLogin(c *gin.Context) {
	var req setupLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	now := time.Now().UTC()
	var provision models.ProvisionCode
	err := h.ProvisionCodes.FindOne(c, bson.M{
		"codeHash":  hashCode(req.Code),
		"usedAt":    bson.M{"$exists": false},
		"expiresAt": bson.M{"$gt": now},
	}).Decode(&provision)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid or expired setup code")
		return
	}

	setupToken, tokenExpiresAt, err := utils.CreateSetupToken(provision.ID.Hex(), provision.Payload.SiteSlug, "", h.Cfg.JWTSecret, setupTokenTTL)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create setup token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"setupToken": setupToken,
		"expiresAt":  tokenExpiresAt,
		"siteId":     "",
		"tenantId":   "",
	})
}

func (h *ProvisionHandler) SetupRegister(c *gin.Context) {
	var req setupRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request")
		return
	}

	provisionIDValue, ok := c.Get(middleware.ContextProvisionID)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing provision context")
		return
	}
	provisionID, err := primitive.ObjectIDFromHex(provisionIDValue.(string))
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid provision context")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	now := time.Now().UTC()
	usedByIP := c.ClientIP()
	client := h.ProvisionCodes.Database().Client()
	session, err := client.StartSession()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to start session")
		return
	}
	defer session.EndSession(c)

	result := struct {
		siteID   primitive.ObjectID
		tenantID primitive.ObjectID
		userID   primitive.ObjectID
	}{}

	_, err = session.WithTransaction(c, func(sc mongo.SessionContext) (interface{}, error) {
		var provision models.ProvisionCode
		if err := h.ProvisionCodes.FindOne(sc, bson.M{"_id": provisionID}).Decode(&provision); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("invalid setup token")
			}
			return nil, err
		}
		if provision.UsedAt != nil {
			return nil, conflictError{message: "setup code already used"}
		}
		if !provision.ExpiresAt.After(now) {
			return nil, errors.New("setup code expired")
		}

		siteSlug := provision.Payload.SiteSlug
		var existingSite models.Site
		err := h.Sites.FindOne(sc, bson.M{"slug": siteSlug}).Decode(&existingSite)
		if err == nil {
			if existingSite.IsProvisioned {
				return nil, conflictError{message: "site already provisioned"}
			}
			adminCount, countErr := h.Users.CountDocuments(sc, bson.M{"siteId": existingSite.ID, "role": "TENANT_ADMIN"})
			if countErr != nil {
				return nil, countErr
			}
			if adminCount > 0 {
				return nil, conflictError{message: "first admin already exists for site"}
			}
			return nil, conflictError{message: "site slug already exists"}
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}

		org := models.Organization{
			Name:      provision.Payload.SiteName,
			BrandName: provision.Payload.SiteName,
			Slug:      siteSlug,
			CreatedAt: now,
			UpdatedAt: now,
		}
		orgInsert, err := h.Organizations.InsertOne(sc, org)
		if err != nil {
			return nil, err
		}
		orgID := orgInsert.InsertedID.(primitive.ObjectID)

		site := models.Site{
			OrganizationID: orgID,
			Name:           provision.Payload.SiteName,
			Slug:           siteSlug,
			IsProvisioned:  true,
			Status:         "draft",
			Content:        map[string]interface{}{},
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		siteInsert, err := h.Sites.InsertOne(sc, site)
		if err != nil {
			return nil, err
		}
		siteID := siteInsert.InsertedID.(primitive.ObjectID)

		tenant := models.Tenant{SiteID: siteID, PanelSlug: siteSlug, CreatedAt: now}
		tenantInsert, err := h.Tenants.InsertOne(sc, tenant)
		if err != nil {
			return nil, err
		}
		tenantID := tenantInsert.InsertedID.(primitive.ObjectID)

		adminCount, err := h.Users.CountDocuments(sc, bson.M{"siteId": siteID, "role": "TENANT_ADMIN"})
		if err != nil {
			return nil, err
		}
		if adminCount > 0 {
			return nil, conflictError{message: "first admin already exists for site"}
		}

		userName := strings.TrimSpace(req.Name)
		if userName == "" {
			userName = req.Email
		}
		user := models.User{
			OrganizationID: orgID,
			SiteID:         &siteID,
			TenantID:       &tenantID,
			Email:          strings.ToLower(req.Email),
			Name:           userName,
			Role:           "TENANT_ADMIN",
			PasswordHash:   string(passwordHash),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		userInsert, err := h.Users.InsertOne(sc, user)
		if err != nil {
			return nil, err
		}

		updateResult, err := h.ProvisionCodes.UpdateOne(sc, bson.M{"_id": provision.ID, "usedAt": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"usedAt": now, "usedByIp": usedByIP}})
		if err != nil {
			return nil, err
		}
		if updateResult.MatchedCount == 0 {
			return nil, conflictError{message: "setup code already used"}
		}

		result.siteID = siteID
		result.tenantID = tenantID
		result.userID = userInsert.InsertedID.(primitive.ObjectID)
		return nil, nil
	}, options.Transaction())
	if err != nil {
		var conflict conflictError
		switch {
		case errors.As(err, &conflict):
			respondError(c, http.StatusConflict, conflict.message)
		case strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "invalid setup token"):
			respondError(c, http.StatusUnauthorized, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "failed to complete setup")
		}
		return
	}

	accessToken, err := utils.CreateToken(result.userID.Hex(), result.tenantID.Hex(), "TENANT_ADMIN", h.Cfg.JWTSecret, time.Duration(h.Cfg.AccessTTLMinutes)*time.Minute)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create access token")
		return
	}
	refreshToken, err := utils.CreateToken(result.userID.Hex(), result.tenantID.Hex(), "TENANT_ADMIN", h.Cfg.JWTRefreshSecret, time.Duration(h.Cfg.RefreshTTLDays)*24*time.Hour)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create refresh token")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"siteId":       result.siteID.Hex(),
		"tenantId":     result.tenantID.Hex(),
	})
}

func generateHumanCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const size = 8
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = alphabet[int(bytes[i])%len(alphabet)]
	}
	return fmt.Sprintf("%s-%s", string(bytes[:4]), string(bytes[4:])), nil
}

func hashCode(code string) string {
	hashed := sha256.Sum256([]byte(strings.TrimSpace(strings.ToUpper(code))))
	return hex.EncodeToString(hashed[:])
}
