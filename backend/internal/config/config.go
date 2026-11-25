package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	CORSAllowedOrigins string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret            string
	InitialAdminPassword string

	LiveKitURL             string
	LiveKitAPIKey          string
	LiveKitAPISecret       string
	MeetingMaxParticipants int
}

func Load() *Config {
	// Support running the server from either the backend folder or repository root.
	_ = godotenv.Load(".env", "backend/.env")

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	liveKitURL := os.Getenv("LIVEKIT_URL")
	if liveKitURL == "" {
		liveKitURL = "ws://localhost:7880"
	}

	liveKitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	if liveKitAPIKey == "" {
		liveKitAPIKey = "devkey"
	}

	liveKitAPISecret := os.Getenv("LIVEKIT_API_SECRET")
	if liveKitAPISecret == "" {
		liveKitAPISecret = "secret"
	}

	maxParticipants := envInt("MEETING_MAX_PARTICIPANTS", 50)
	if maxParticipants <= 0 {
		maxParticipants = 50
	}

	return &Config{
		AppPort:            appPort,
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),

		JWTSecret:            os.Getenv("JWT_SECRET"),
		InitialAdminPassword: os.Getenv("ADMIN_INITIAL_PASSWORD"),

		LiveKitURL:             liveKitURL,
		LiveKitAPIKey:          liveKitAPIKey,
		LiveKitAPISecret:       liveKitAPISecret,
		MeetingMaxParticipants: maxParticipants,
	}
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
