package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	// db.AutoMigrate(
	// 	&domain.User{},
	// 	&domain.Role{},
	// 	&domain.Permission{},
	// )
	// SeedData(db)

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
	api.Get("/admin/dashboard", guard("dashboard:view"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello Admin! This is secret dashboard."})
	})
	api.Get("/profile", guard("profile:view"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello User! This is your profile."})
	})

	// --- RBAC Management Routes ---
	adminPanel := api.Group("/admin/panel", guard("system:admin"))

	// GET Routes สำหรับดูข้อมูล (เพิ่มเข้ามาใหม่)
	adminPanel.Get("/roles", rbacHandler.GetRoles)
	adminPanel.Get("/permissions", rbacHandler.GetPermissions)
	adminPanel.Get("/users/:id/roles", rbacHandler.GetUserRoles) // สังเกตการใช้ :id

	// POST / DELETE Routes (ของเดิม)
	adminPanel.Post("/roles", rbacHandler.CreateRole)
	adminPanel.Post("/permissions", rbacHandler.CreatePermission)
	adminPanel.Post("/roles/assign-perm", rbacHandler.AssignPermission)
	adminPanel.Post("/users/assign-role", rbacHandler.AssignRole)
	adminPanel.Delete("/roles/remove-perm", rbacHandler.RemovePermission)
	adminPanel.Delete("/users/remove-role", rbacHandler.RemoveRole)

	// ==========================================
	// 🛑 Graceful Shutdown Setup
	// ==========================================

	// สร้าง Channel ไว้รอรับสัญญาณการปิดโปรแกรม (เช่น กด Ctrl+C หรือ Docker สั่งหยุด)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// สั่งให้ Goroutine รอฟังเสียงสัญญาณ
	go func() {
		<-c // รอจนกว่าจะมีสัญญาณเข้ามา
		fmt.Println("\n🛑 Gracefully shutting down server...")

		// ปิด Fiber App อย่างนุ่มนวล (รอให้ Request ที่ค้างอยู่ ทำงานเสร็จก่อน)
		if err := app.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}

		// (Optional) สั่งปิด Database และ Redis
		sqlDB, _ := db.DB()
		sqlDB.Close()
		rdb.Close()
		fmt.Println("✅ All connections closed. Goodbye!")
	}()

	// Start Server (เปลี่ยนจาก log.Fatal เป็นการเช็ค err ธรรมดา เพื่อให้บรรทัดข้างบนได้ทำงาน)
	fmt.Printf("🚀 Server is starting on port %s\n", cfg.Server.Port)
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Panic(err) // ถ้า Port ชน หรือ Start ไม่ขึ้นตั้งแต่แรก ค่อย Panic
	}
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
