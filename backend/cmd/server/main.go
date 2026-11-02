package main

import (
	"log"

	"admin_daebak_music/backend/internal/config"
	"admin_daebak_music/backend/internal/database"
	"admin_daebak_music/backend/internal/handlers"
	"admin_daebak_music/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	authH := &handlers.AuthHandler{DB: db, Cfg: cfg}
	usersH := &handlers.UsersHandler{DB: db}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			// Register publik dimatikan. Admin dibuat via bootstrap / POST /api/users.
			auth.POST("/bootstrap-admin", authH.BootstrapAdmin)
			auth.POST("/login", authH.Login)
			auth.POST("/google", authH.GoogleLogin)
			auth.POST("/refresh", authH.Refresh)
			auth.POST("/logout", middleware.Auth(cfg), authH.Logout)
			auth.GET("/me", middleware.Auth(cfg), authH.Me)
		}

		users := api.Group("/users")
		users.Use(middleware.Auth(cfg))
		{
			users.GET("", middleware.RequireRole("admin"), usersH.List)
			users.GET("/:id", middleware.RequireRole("admin"), usersH.GetByID)
			users.POST("", middleware.RequireRole("admin"), usersH.Create)
			users.PATCH("/me", usersH.UpdateProfile)
		}
	}

	log.Printf("server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}