package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/handlers"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(router *gin.Engine, db *mongo.Database, cfg *config.Config) {
	frontendOrigins := cfg.FrontendOrigins
	if len(frontendOrigins) == 0 {
		frontendOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     frontendOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authHandler := &handlers.AuthHandler{Users: db.Collection("users"), Cfg: cfg}
	siteHandler := &handlers.SiteHandler{Sites: db.Collection("sites"), SitePermissions: db.Collection("site_permissions")}
	adminHandler := &handlers.AdminHandler{Users: db.Collection("users"), Sites: db.Collection("sites"), SitePermissions: db.Collection("site_permissions")}
	provisionHandler := &handlers.ProvisionHandler{Cfg: cfg, Users: db.Collection("users")}
	publicHandler := &handlers.PublicHandler{Sites: db.Collection("sites")}

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)

		provision := api.Group("/provision")
		provision.Use(middleware.ProvisionAPIKeyRequired(cfg.ProvisionAPIKey))
		provision.POST("/bootstrap", provisionHandler.Bootstrap)

		api.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)

		secured := api.Group("")
		secured.Use(middleware.AuthRequired(cfg.JWTSecret))
		secured.GET("/sites", siteHandler.List)
		secured.POST("/sites", siteHandler.Create)
		secured.GET("/sites/:id", siteHandler.Get)
		secured.PUT("/sites/:id/content", siteHandler.UpdateContent)
		secured.POST("/sites/:id/publish", siteHandler.Publish)
		secured.POST("/sites/:id/unpublish", siteHandler.Unpublish)

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.SuperAdminRequired())
		admin.GET("/sites", adminHandler.ListSites)
		admin.POST("/sites", adminHandler.CreateSiteDirect)
		admin.POST("/sites/:id/grant", adminHandler.GrantSiteAccess)
		admin.GET("/sites/:id/users", adminHandler.ListSiteUsers)
		admin.POST("/users", adminHandler.CreateUser)
		admin.GET("/users", adminHandler.ListUsers)
	}

	router.GET("/s/:slug", publicHandler.GetPublishedSite)
}
