package config

import (
	"os"
)

// SearchConfig holds search-related configuration
type SearchConfig struct {
	Host     string
	APIKey   string
	Indexes  []string
	Settings SearchSettings
}

// SearchSettings holds search-specific settings
type SearchSettings struct {
	DefaultLimit     int
	MaxLimit         int
	DefaultPage      int
	HighlightEnabled bool
	SuggestionsLimit int
}

// GetSearchConfig returns the search configuration
func GetSearchConfig() *SearchConfig {
	return &SearchConfig{
		Host:   getEnv("MEILISEARCH_HOST", "http://localhost:7700"),
		APIKey: getEnv("MEILISEARCH_API_KEY", ""),
		Indexes: []string{
			"posts",
			"users",
			"categories",
			"media",
		},
		Settings: SearchSettings{
			DefaultLimit:     20,
			MaxLimit:         100,
			DefaultPage:      1,
			HighlightEnabled: true,
			SuggestionsLimit: 10,
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
