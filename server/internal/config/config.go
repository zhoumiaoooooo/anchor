package config

import (
	"os"
	"strings"
)

type Config struct {
	Port            string
	DatabasePath    string
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	DeepSeekModel   string
	UploadDir       string
	AnchorGenTime   string // "14:30"
}

func Load() *Config {
	loadDotEnv(".env")

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabasePath:    getEnv("DATABASE_PATH", "./data/anchor.db"),
		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		UploadDir:       getEnv("UPLOAD_DIR", "./data/uploads"),
		AnchorGenTime:   getEnv("ANCHOR_GEN_TIME", "14:30"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		os.Setenv(key, value)
	}
}
