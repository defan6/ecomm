package util

import (
	"os"
	"time"
)

func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetJWTSecret() []byte {
	return []byte(GetEnvOrDefault("JWT_SECRET", "mysecretkey"))
}

func GetAccessTokenExpiration() time.Duration {
	durationStr := GetEnvOrDefault("ACCESS_TOKEN_EXPIRATION", "15m")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return time.Minute * 15
	}
	return duration
}

func GetRefreshTokenExpiration() time.Duration {
	durationStr := GetEnvOrDefault("REFRESH_TOKEN_EXPIRATION", "7d")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return duration
}
