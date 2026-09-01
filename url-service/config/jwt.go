package config

func GetJWTSecret() string {
	return getEnv("JWT_SECRET", "your-secret-key")
}
