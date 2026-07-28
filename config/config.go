package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port           string
	DBPath         string
	JWTSecret      string
	JWTExpiry      string
	AppURL         string
	AllowedOrigins string
}

var Config AppConfig 

func LoadConfig() {
	err:= godotenv.Load()
	if err != nil {
		    log.Println(" File .env tidak ditemukan, menggunakan default environtment.")
	}

	Config = AppConfig{
		Port:           getEnv("PORT", "5000"),
		DBPath:         getEnv("DATABASE_PATH", "./timbangan.db"),
		JWTSecret:      getEnv("JWT_SECRET", "timbang_sapi_iot_secure_jwt_secret_key_2026"),
		JWTExpiry:      getEnv("JWT_EXPIRY", "24h"),
		AppURL:         getEnv("APP_URL", "http://localhost:5000"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "https://timbangan-digital-project.vercel.app,http://localhost:5173,http://localhost:3000,*"),
	}

	log.Printf("✅ Konfigurasi dimuat — Port: %s, DB: %s, AppURL: %s", Config.Port, Config.DBPath, Config.AppURL)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value

}