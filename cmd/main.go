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
	fmt.Println("Successfully connected to Redis" + rdb.Options().Addr)

	// 4. Auto Migrate (สร้างตาราง)
	db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.Permission{},
	)
	SeedData(db)

	// --- Repository Init ---
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// --- Service Init ---
	// RBAC Service (ใช้ Redis และ Repo ครบชุด)
	rbacService := service.NewRBACService(userRepo, roleRepo, permissionRepo, rdb)
	if err := rbacService.LoadPolicy(); err != nil {
		log.Printf("⚠️ Warning: Failed to load RBAC policy: %v", err)
	}

	// Auth Service
	authService := service.NewAuthService(userRepo, cfg.Server.JWTSecret)

	// --- Handler Init ---
	authHandler := http.NewAuthHandler(authService)
	rbacHandler := http.NewRBACHandler(rbacService) // ✅ เพิ่มตรงนี้

	// --- Middleware Setup ---
	// สร้างฟังก์ชันเช็คสิทธิ์ (Guard)
	guard := http.NewRBACMiddleware(cfg, rbacService)

	// 5. Server Setup
	app := fiber.New()
	api := app.Group("/api")

	// --- Public Routes ---
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// --- Protected Routes ---

	// 1. Admin Dashboard (Test Permission)
	api.Get("/admin/dashboard", guard("dashboard:view"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello Admin! This is secret dashboard."})
	})

	// 2. User Profile (Test Permission)
	api.Get("/profile", guard("profile:view"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello User! This is your profile."})
	})

	// --- RBAC Management Routes (New!) ---
	// เฉพาะ Admin เท่านั้นที่เข้ามาจัดการ Role/Permission ได้
	// ต้องมีสิทธิ์ "system:admin" (ซึ่งเรา Seed ให้ role:admin มีสิทธิ์นี้แล้ว)
	adminPanel := api.Group("/admin/panel", guard("system:admin"))

	adminPanel.Post("/roles", rbacHandler.CreateRole)                     // สร้าง Role ใหม่
	adminPanel.Post("/permissions", rbacHandler.CreatePermission)         // สร้าง Permission ใหม่
	adminPanel.Post("/roles/assign-perm", rbacHandler.AssignPermission)   // จับคู่ Role <-> Permission
	adminPanel.Post("/users/assign-role", rbacHandler.AssignRole)         // จับคู่ User <-> Role
	adminPanel.Delete("/roles/remove-perm", rbacHandler.RemovePermission) // เอาสิทธิ์ออกจาก Role
	adminPanel.Delete("/users/remove-role", rbacHandler.RemoveRole)       // เอา Role ออกจาก User

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
	permSystemAdmin := domain.Permission{Name: "system:admin"} // ✅ เพิ่มสิทธิ์สำหรับจัดการระบบ
	db.Create(&permDashboard)
	db.Create(&permProfile)
	db.Create(&permSystemAdmin)

	// 2. สร้าง Roles และจับคู่ Permission
	// Admin ทำได้หมด รวมถึงจัดการระบบ
	roleAdmin := domain.Role{
		Name:        "admin",
		Permissions: []*domain.Permission{&permDashboard, &permProfile, &permSystemAdmin},
	}
	// User ดูได้แค่ Profile
	roleUser := domain.Role{
		Name:        "user",
		Permissions: []*domain.Permission{&permProfile},
	}

	db.Create(&roleAdmin)
	db.Create(&roleUser)

	fmt.Println("✅ Seed Data Completed!")
}
