package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Organization struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	BrandName string             `bson:"brandName" json:"brandName"`
	Slug      string             `bson:"slug" json:"slug"`
	PlanID    primitive.ObjectID `bson:"planId" json:"planId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID primitive.ObjectID `bson:"orgId" json:"orgId"`
	Email          string             `bson:"email" json:"email"`
	Name           string             `bson:"name" json:"name"`
	Role           string             `bson:"role" json:"role"`
	PasswordHash   string             `bson:"passwordHash" json:"-"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Site struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	OrganizationID primitive.ObjectID     `bson:"orgId" json:"orgId"`
	Name           string                 `bson:"name" json:"name"`
	Slug           string                 `bson:"slug" json:"slug"`
	Status         string                 `bson:"status" json:"status"`
	Content        map[string]interface{} `bson:"content" json:"content"`
	CreatedAt      time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time              `bson:"updatedAt" json:"updatedAt"`
	PublishedAt    *time.Time             `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
}

type PageContent struct {
	Blocks []map[string]interface{} `bson:"blocks" json:"blocks"`
	Meta   map[string]interface{}   `bson:"meta" json:"meta"`
}

type Plan struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	PriceCents int64              `bson:"priceCents" json:"priceCents"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Subscription struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganizationID primitive.ObjectID `bson:"orgId" json:"orgId"`
	PlanID         primitive.ObjectID `bson:"planId" json:"planId"`
	Status         string             `bson:"status" json:"status"`
	RenewsAt       time.Time          `bson:"renewsAt" json:"renewsAt"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}
