package main

import (
	"context"
	"fmt"
	"time"

	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/db"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	seedOrgName      = "Demo Reseller"
	seedOrgBrandName = "Demo"
	seedUserEmail    = "demo@reseller.com"
	seedUserPassword = "ChangeMe123!"
	seedUserRole     = "reseller"
)

func runSeed(ctx context.Context, cfg *config.Config) error {
	mongoConn, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return fmt.Errorf("mongo error: %w", err)
	}
	defer func() {
		_ = mongoConn.Client.Disconnect(ctx)
	}()

	orgs := mongoConn.DB.Collection("organizations")
	users := mongoConn.DB.Collection("users")
	now := time.Now()

	orgFilter := bson.M{"name": seedOrgName}
	orgUpdate := bson.M{
		"$set": bson.M{
			"name":      seedOrgName,
			"brandName": seedOrgBrandName,
			"slug":      utils.NormalizeSlug(seedOrgName),
			"updatedAt": now,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	if _, err := orgs.UpdateOne(ctx, orgFilter, orgUpdate, options.Update().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert org: %w", err)
	}

	var org models.Organization
	if err := orgs.FindOne(ctx, orgFilter).Decode(&org); err != nil {
		return fmt.Errorf("fetch org: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seedUserPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	userFilter := bson.M{"email": seedUserEmail}
	userUpdate := bson.M{
		"$set": bson.M{
			"email":        seedUserEmail,
			"orgId":        org.ID,
			"role":         seedUserRole,
			"passwordHash": string(passwordHash),
			"updatedAt":    now,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	if _, err := users.UpdateOne(ctx, userFilter, userUpdate, options.Update().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	var user models.User
	if err := users.FindOne(ctx, userFilter).Decode(&user); err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}

	fmt.Printf("Seeded organization ID: %s\n", org.ID.Hex())
	fmt.Printf("Seeded user ID: %s\n", user.ID.Hex())
	return nil
}
