package casbin

import (
	"fmt"
	"os"
	"path/filepath"
)

// ModelLoader provides functionality to load RBAC model configurations
type ModelLoader struct {
	configDir string
}

// NewModelLoader creates a new model loader
func NewModelLoader(configDir string) *ModelLoader {
	return &ModelLoader{
		configDir: configDir,
	}
}

// LoadModel loads the RBAC model configuration based on environment
func (ml *ModelLoader) LoadModel(environment string) (string, error) {
	var modelFile string

	switch environment {
	case "production", "prod":
		modelFile = "rbac_model_prod.conf"
	case "development", "dev":
		modelFile = "rbac_model_dev.conf"
	default:
		modelFile = "rbac_model.conf"
	}

	modelPath := filepath.Join(ml.configDir, modelFile)

	// Check if environment-specific file exists, fallback to default
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		modelPath = filepath.Join(ml.configDir, "rbac_model.conf")
	}

	content, err := os.ReadFile(modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to read RBAC model configuration from %s: %w", modelPath, err)
	}

	return string(content), nil
}

// LoadDefaultModel loads the default RBAC model configuration
func (ml *ModelLoader) LoadDefaultModel() (string, error) {
	return ml.LoadModel("")
}

// ValidateModel validates the RBAC model configuration
func (ml *ModelLoader) ValidateModel(modelContent string) error {
	// Basic validation - check if required sections exist
	requiredSections := []string{
		"[request_definition]",
		"[policy_definition]",
		"[role_definition]",
		"[policy_effect]",
		"[matchers]",
	}

	for _, section := range requiredSections {
		if !contains(modelContent, section) {
			return fmt.Errorf("missing required section: %s", section)
		}
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

// containsSubstring checks if a string contains a substring (simplified)
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
