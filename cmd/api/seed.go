package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/db"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/models"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func runSeed(ctx context.Context, cfg *config.Config) error {
	mongoConn, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return fmt.Errorf("mongo error: %w", err)
	}
	defer func() { _ = mongoConn.Client.Disconnect(ctx) }()

	users := mongoConn.DB.Collection("users")
	sites := mongoConn.DB.Collection("sites")
	permissions := mongoConn.DB.Collection("site_permissions")

	var demoUserID primitive.ObjectID
	if cfg.SuperAdminEmail != "" && cfg.SuperAdminPassword != "" {
		if _, err := upsertUser(ctx, users, cfg.SuperAdminEmail, cfg.SuperAdminPassword, "superadmin"); err != nil {
			return err
		}
	}
	if cfg.DemoEmail != "" && cfg.DemoPassword != "" {
		id, err := upsertUser(ctx, users, cfg.DemoEmail, cfg.DemoPassword, "user")
		if err != nil {
			return err
		}
		demoUserID = id
	}

	demoSiteID, err := upsertSite(ctx, sites, cfg.DemoSiteSlug)
	if err != nil {
		return err
	}

	if demoUserID != primitive.NilObjectID {
		now := time.Now().UTC()
		_, err := permissions.UpdateOne(ctx,
			bson.M{"siteId": demoSiteID, "userId": demoUserID},
			bson.M{"$set": bson.M{"role": "editor", "updatedAt": now}, "$setOnInsert": bson.M{"createdAt": now}},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			return fmt.Errorf("grant demo permission: %w", err)
		}
	}

	fmt.Println("Seed completed")
	return nil
}

func upsertUser(ctx context.Context, users *mongo.Collection, email, password, role string) (primitive.ObjectID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("hash password: %w", err)
	}
	now := time.Now().UTC()
	email = strings.ToLower(strings.TrimSpace(email))
	_, err = users.UpdateOne(ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"passwordHash": string(hash), "globalRole": role, "updatedAt": now}, "$setOnInsert": bson.M{"createdAt": now}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("upsert user: %w", err)
	}
	var user models.User
	if err := users.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return primitive.NilObjectID, fmt.Errorf("find user: %w", err)
	}
	return user.ID, nil
}

func upsertSite(ctx context.Context, sites *mongo.Collection, slug string) (primitive.ObjectID, error) {
	now := time.Now().UTC()
	slug = utils.NormalizeSlug(slug)
	_, err := sites.UpdateOne(ctx,
		bson.M{"slug": slug},
		bson.M{"$set": bson.M{"name": "Demo Site", "status": "draft", "updatedAt": now}, "$setOnInsert": bson.M{"content": bson.M{}, "createdAt": now}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("upsert site: %w", err)
	}
	var site models.Site
	if err := sites.FindOne(ctx, bson.M{"slug": slug}).Decode(&site); err != nil {
		return primitive.NilObjectID, fmt.Errorf("find site: %w", err)
	}
	return site.ID, nil
}
