package main

import (
	"fmt"
	"log"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/config"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/handler/http"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/storage/postgres"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/storage/postgres/repository"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/service"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("Config loaded successfully")

	// 2. Connect Database
	db, err := postgres.NewPostgresDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Successfully connected to Database" + db.Name())

	// 3. Connect Redis
	// rdb, err := redis.NewRedisClient(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to connect to Redis: %v", err)
	// }

	// 4. Auto Migrate (สร้างตาราง)
	db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.Permission{},
	)

	// 5. Dependency Injection (Wiring)
	// Repo -> Service -> Handler
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.Server.JWTSecret)
	authHandler := http.NewAuthHandler(authService)

	// 6. Server Setup
	app := fiber.New()

	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	log.Fatal(app.Listen(":" + cfg.Server.Port))
}
