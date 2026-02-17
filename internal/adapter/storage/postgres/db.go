package postgres

import (
	"fmt"
	"time"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // ให้ปริ้น SQL ออกมาดูตอน Dev
	})

	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("cannot get database instance: %w", err)
	}

	// จูนตามความเหมาะสมของ Server
	sqlDB.SetMaxIdleConns(10)  // Connection ที่เปิดรอไว้
	sqlDB.SetMaxOpenConns(100) // Connection สูงสุดที่รับได้
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
