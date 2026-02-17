package main

import (
	"fmt"
	"log"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/config"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/handler/http"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/storage/postgres"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/storage/postgres/repository"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/adapter/storage/redis"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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
	rdb, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// 4. Auto Migrate (สร้างตาราง)
	db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.Permission{},
	)
	SeedData(db)

	// --- RBAC Setup ---
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	rbacService := service.NewRBACService(userRepo, roleRepo, permissionRepo, rdb)
	if err := rbacService.LoadPolicy(); err != nil {
		log.Printf("⚠️ Warning: Failed to load RBAC policy: %v", err)
	}

	// 5. Dependency Injection (Wiring)
	// Repo -> Service -> Handler
	authService := service.NewAuthService(userRepo, cfg.Server.JWTSecret)
	authHandler := http.NewAuthHandler(authService)

	// --- Middleware Setup ---
	// สร้างฟังก์ชันเช็คสิทธิ์ (Guard)
	guard := http.NewRBACMiddleware(cfg, rbacService)

	// 6. Server Setup
	app := fiber.New()

	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected Routes (ต้องมีสิทธิ์)
	// ตัวอย่าง 1: Admin เท่านั้นที่ดู Dashboard ได้
	api.Get("/admin/dashboard", guard("dashboard:view"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello Admin! This is secret dashboard."})
	})

	// ตัวอย่าง 2: User ทั่วไปก็ดู Profile ได้ (ถ้ามีสิทธิ์)
	api.Get("/profile", guard("profile:view"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello User! This is your profile."})
	})

	log.Fatal(app.Listen(":" + cfg.Server.Port))
}

func SeedData(db *gorm.DB) {
	var count int64
	db.Model(&domain.Role{}).Count(&count)
	if count > 0 {
		return // มีข้อมูลแล้ว ไม่สร้างซ้ำ
	}

	fmt.Println("Seeding Data...")

	// 1. สร้าง Permissions
	permDashboard := domain.Permission{Name: "dashboard:view"}
	permProfile := domain.Permission{Name: "profile:view"}
	db.Create(&permDashboard)
	db.Create(&permProfile)

	// 2. สร้าง Roles และจับคู่ Permission
	roleAdmin := domain.Role{Name: "admin", Permissions: []*domain.Permission{&permDashboard, &permProfile}} // Admin ทำได้หมด
	roleUser := domain.Role{Name: "user", Permissions: []*domain.Permission{&permProfile}}                   // User ดูได้แค่ Profile
	db.Create(&roleAdmin)
	db.Create(&roleUser)

	fmt.Println("✅ Seed Data Completed!")
	// หมายเหตุ: User ต้องไป Register ผ่าน API แล้วค่อยมาแก้ DB ผูก Role เอาเองนะครับ (หรือเขียน Seed User เพิ่มก็ได้)
}
