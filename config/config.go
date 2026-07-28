package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port string
	DBPath string
	JWTSecret string
	JWTExpiry string
}

var Config AppConfig 

func LoadConfig() {
	err:= godotenv.Load()
	if err != nil {
		    log.Println(" File .env tidak ditemukan, menggunakan default environtment.")
	}

	Config = AppConfig{
		Port:      getEnv("PORT", "5000"),
		DBPath:    getEnv("DATABASE_PATH", "./timbangan.db"),
		JWTSecret: getEnv("JWT_SECRET", "default_secret_key"),
		JWTExpiry: getEnv("JWT_EXPIRY", "24h"),
	}

	log.Printf("✅ Konfigurasi dimuat — Port: %s, DB: %s", Config.Port, Config.DBPath)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value

}