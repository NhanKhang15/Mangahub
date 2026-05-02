package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPPort  string
	MongoURI  string
	MongoDB   string
	JWTSecret string
	JWTAccess time.Duration
	JWTRefresh time.Duration

	MangaDexBase  string
	MangaDexToken string

	MALBase     string
	MALClientID string

	AniListBase  string
	AniListToken string

	AdminToken string

	PollInterval time.Duration

	Env string // dev | prod
}

func Load() Config {
	_ = loadDotEnv(".env")
	return Config{
		HTTPPort:      getenv("HTTP_PORT", "8080"),
		MongoURI:      getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:       getenv("MONGO_DB", "mangahub"),
		JWTSecret:     getenv("JWT_SECRET", "change-me"),
		JWTAccess:     getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefresh:    getDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		MangaDexBase:  getenv("MANGADEX_BASE", "https://api.mangadex.org"),
		MangaDexToken: getenv("MANGADEX_TOKEN", ""),
		MALBase:       getenv("MAL_BASE", "https://api.myanimelist.net/v2"),
		MALClientID:   getenv("MAL_CLIENT_ID", ""),
		AniListBase:   getenv("ANILIST_BASE", "https://graphql.anilist.co"),
		AniListToken:  getenv("ANILIST_TOKEN", ""),
		AdminToken:    getenv("ADMIN_TOKEN", ""),
		PollInterval:  getDuration("POLL_INTERVAL", 5*time.Minute),
		Env:           getenv("ENV", "dev"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getDuration(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func loadDotEnv(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return nil
}
