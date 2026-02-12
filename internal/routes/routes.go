package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/config"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/handlers"
	"github.com/youpp/youpp-website-adminpanel-backend/internal/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(router *gin.Engine, db *mongo.Database, cfg *config.Config) {
	authHandler := &handlers.AuthHandler{
		Users: db.Collection("users"),
		Cfg:   cfg,
	}
	provisionHandler := &handlers.ProvisionHandler{
		Cfg:            cfg,
		ProvisionCodes: db.Collection("provision_codes"),
		Organizations:  db.Collection("organizations"),
		Sites:          db.Collection("sites"),
		Tenants:        db.Collection("tenants"),
		Users:          db.Collection("users"),
	}
	siteHandler := &handlers.SiteHandler{
		Sites: db.Collection("sites"),
	}
	publicHandler := &handlers.PublicHandler{
		Sites: db.Collection("sites"),
	}

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/setup-login", provisionHandler.SetupLogin)

			setup := auth.Group("")
			setup.Use(middleware.SetupTokenRequired(cfg.JWTSecret))
			{
				setup.POST("/setup-register", provisionHandler.SetupRegister)
			}
		}

		provision := api.Group("/provision")
		provision.Use(middleware.ProvisionAPIKeyRequired(cfg.ProvisionAPIKey))
		{
			provision.POST("/request-code", provisionHandler.RequestCode)
		}

		api.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)

		secured := api.Group("")
		secured.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			secured.GET("/sites", siteHandler.List)
			secured.POST("/sites", siteHandler.Create)
			secured.GET("/sites/:id", siteHandler.Get)
			secured.PUT("/sites/:id/content", siteHandler.UpdateContent)
			secured.POST("/sites/:id/publish", siteHandler.Publish)
			secured.POST("/sites/:id/unpublish", siteHandler.Unpublish)
		}
	}

	router.GET("/s/:slug", publicHandler.GetPublishedSite)
}
