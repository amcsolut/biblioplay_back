package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL              string
	JWTSecret               string
	JWTAlgorithm            string
	AccessTokenExpireMinutes int
	RefreshTokenExpireDays   int
	CORSOrigins             []string
	SMTPHost                string
	SMTPPort                int
	SMTPUser                string
	SMTPPassword            string
	Environment             string
	// OAuth Configuration
	GoogleClientID     string
	GoogleClientSecret string
	FacebookAppID      string
	FacebookAppSecret  string
	// UploadsPath: diretório local (dev) ou volume montado (produção) para arquivos servidos em /uploads
	UploadsPath string
}

func Load() *Config {
	accessTokenExpire, _ := strconv.Atoi(getEnv("ACCESS_TOKEN_EXPIRE_MINUTES", "30"))
	// REFRESH_TOKEN_EXPIRE can be in days (default) or hours if specified with 'h' suffix
	refreshTokenExpireStr := getEnv("REFRESH_TOKEN_EXPIRE", "7")
	refreshTokenExpire, _ := strconv.Atoi(refreshTokenExpireStr)
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		JWTSecret:               getEnv("JWT_SECRET", "your-secret-key"),
		JWTAlgorithm:            getEnv("JWT_ALGORITHM", "HS256"),
		AccessTokenExpireMinutes: accessTokenExpire,
		RefreshTokenExpireDays:   refreshTokenExpire,
		SMTPHost:                getEnv("SMTP_HOST", ""),
		SMTPPort:                smtpPort,
		SMTPUser:                getEnv("SMTP_USER", ""),
		SMTPPassword:            getEnv("SMTP_PASSWORD", ""),
		Environment:             getEnv("ENVIRONMENT", "development"),
		GoogleClientID:          getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		FacebookAppID:           getEnv("FACEBOOK_APP_ID", ""),
		FacebookAppSecret:       getEnv("FACEBOOK_APP_SECRET", ""),
		UploadsPath:             getEnv("UPLOADS_PATH", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetAccessTokenDuration() time.Duration {
	return time.Duration(c.AccessTokenExpireMinutes) * time.Minute
}

func (c *Config) GetRefreshTokenDuration() time.Duration {
	return time.Duration(c.RefreshTokenExpireDays) * 24 * time.Hour
}

