package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port      string
	JWTSecret string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string // เพิ่ม
	TimeZone string // เพิ่ม
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func LoadConfig() (*Config, error) {
	// บอก Viper ว่าไฟล์ชื่ออะไร อยู่ที่ไหน
	viper.SetConfigName("/config/config") // ชื่อไฟล์ config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // ให้หาที่ root folder

	// เผื่ออยาก override ด้วย Environment Variable (เช่น SERVER_PORT=8080)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// อ่านไฟล์
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// แปลงข้อมูลใส่ Struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
