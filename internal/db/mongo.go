package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func Connect(ctx context.Context, uri, dbName string) (*Mongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Mongo{Client: client, DB: client.Database(dbName)}, nil
}

func EnsureIndexes(ctx context.Context, database *mongo.Database) error {
	if err := validateDuplicateSlugs(ctx, database.Collection("sites")); err != nil {
		return err
	}

	sites := database.Collection("sites")
	if _, err := sites.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("slug_1"),
	}); err != nil {
		return fmt.Errorf("create sites slug index: %w", err)
	}

	users := database.Collection("users")
	if _, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("email_1"),
	}); err != nil {
		return fmt.Errorf("create users email index: %w", err)
	}

	permissions := database.Collection("site_permissions")
	if _, err := permissions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "siteId", Value: 1}}, Options: options.Index().SetName("siteId_1")},
		{Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetName("userId_1")},
		{Keys: bson.D{{Key: "siteId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("siteId_1_userId_1")},
	}); err != nil {
		return fmt.Errorf("create site_permissions indexes: %w", err)
	}

	return nil
}

func validateDuplicateSlugs(ctx context.Context, sites *mongo.Collection) error {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$slug", "count": bson.M{"$sum": 1}}}},
		{{Key: "$match", Value: bson.M{"_id": bson.M{"$ne": ""}, "count": bson.M{"$gt": 1}}}},
	}
	cursor, err := sites.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("validate duplicate slugs: %w", err)
	}
	defer cursor.Close(ctx)
	var dups []struct {
		Slug  string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err := cursor.All(ctx, &dups); err != nil {
		return fmt.Errorf("decode duplicate slugs: %w", err)
	}
	if len(dups) == 0 {
		return nil
	}
	parts := make([]string, 0, len(dups))
	for _, d := range dups {
		parts = append(parts, fmt.Sprintf("%s(%d)", d.Slug, d.Count))
	}
	return fmt.Errorf("duplicate site slugs found, cannot ensure unique index: %s", strings.Join(parts, ", "))
}
