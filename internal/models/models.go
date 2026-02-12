package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	GlobalRole   string             `bson:"globalRole" json:"globalRole"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Site struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Slug        string                 `bson:"slug" json:"slug"`
	Status      string                 `bson:"status" json:"status"`
	Content     map[string]interface{} `bson:"content" json:"content"`
	CreatedAt   time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time              `bson:"updatedAt" json:"updatedAt"`
	PublishedAt *time.Time             `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
}

type SitePermission struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID    primitive.ObjectID `bson:"siteId" json:"siteId"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Role      string             `bson:"role" json:"role"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Legacy types still used by existing provisioning flows.
type ProvisionCodePayload struct {
	SiteName string `bson:"siteName" json:"siteName"`
	SiteSlug string `bson:"siteSlug" json:"siteSlug"`
}

type ProvisionCode struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	CodeHash  string               `bson:"codeHash" json:"codeHash"`
	ExpiresAt time.Time            `bson:"expiresAt" json:"expiresAt"`
	UsedAt    *time.Time           `bson:"usedAt,omitempty" json:"usedAt,omitempty"`
	UsedByIP  string               `bson:"usedByIp,omitempty" json:"usedByIp,omitempty"`
	Payload   ProvisionCodePayload `bson:"payload" json:"payload"`
	CreatedAt time.Time            `bson:"createdAt" json:"createdAt"`
}
