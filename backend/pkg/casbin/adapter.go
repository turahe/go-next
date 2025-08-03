package casbin

import (
	"fmt"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/logger"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// AdapterConfig holds configuration for the Casbin GORM adapter
type AdapterConfig struct {
	TableName    string // Custom table name (default: "casbin_rule")
	AutoMigrate  bool   // Whether to auto-migrate the table
	BatchSize    int    // Batch size for operations (default: 1000)
	MaxRetries   int    // Maximum retries for database operations
	EnableCache  bool   // Enable adapter-level caching
	CacheTimeout int    // Cache timeout in seconds
}

// DefaultAdapterConfig returns default adapter configuration
func DefaultAdapterConfig() *AdapterConfig {
	return &AdapterConfig{
		TableName:    "casbin_rule",
		AutoMigrate:  true,
		BatchSize:    1000,
		MaxRetries:   3,
		EnableCache:  true,
		CacheTimeout: 300, // 5 minutes
	}
}

// CasbinAdapter provides enhanced Casbin GORM adapter functionality
type CasbinAdapter struct {
	adapter *gormadapter.Adapter
	config  *AdapterConfig
	db      *gorm.DB
}

// NewCasbinAdapter creates a new Casbin GORM adapter
func NewCasbinAdapter(config *AdapterConfig) (*CasbinAdapter, error) {
	if config == nil {
		config = DefaultAdapterConfig()
	}

	// Create GORM adapter with custom configuration
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(
		database.DB,
		config.TableName,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Auto-migrate if enabled
	if config.AutoMigrate {
		if err := database.DB.AutoMigrate(&models.CasbinRule{}); err != nil {
			logger.Warnf("Failed to auto-migrate Casbin table: %v", err)
		}
	}

	return &CasbinAdapter{
		adapter: adapter,
		config:  config,
		db:      database.DB,
	}, nil
}

// GetAdapter returns the underlying GORM adapter
func (ca *CasbinAdapter) GetAdapter() *gormadapter.Adapter {
	return ca.adapter
}

// CreateEnforcer creates a Casbin enforcer with the adapter
func (ca *CasbinAdapter) CreateEnforcer(modelConfig string) (*casbin.Enforcer, error) {
	// Parse model configuration
	m, err := model.NewModelFromString(modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin model: %w", err)
	}

	// Create enforcer with adapter
	e, err := casbin.NewEnforcer(m, ca.adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Enable auto-save if configured
	e.EnableAutoSave(true)

	logger.Infof("Casbin enforcer created successfully with GORM adapter")
	return e, nil
}

// GetPolicyCount returns the total number of policies
func (ca *CasbinAdapter) GetPolicyCount() (int64, error) {
	var count int64
	err := ca.db.Table(ca.config.TableName).Count(&count).Error
	return count, err
}

// GetRoleCount returns the total number of role assignments
func (ca *CasbinAdapter) GetRoleCount() (int64, error) {
	var count int64
	err := ca.db.Table(ca.config.TableName).Where("ptype = ?", "g").Count(&count).Error
	return count, err
}

// GetPolicyCountByRole returns the number of policies for a specific role
func (ca *CasbinAdapter) GetPolicyCountByRole(role string) (int64, error) {
	var count int64
	err := ca.db.Table(ca.config.TableName).Where("ptype = ? AND v0 = ?", "p", role).Count(&count).Error
	return count, err
}

// GetUserRoleCount returns the number of role assignments for a user
func (ca *CasbinAdapter) GetUserRoleCount(userID string) (int64, error) {
	var count int64
	err := ca.db.Table(ca.config.TableName).Where("ptype = ? AND v1 = ?", "g", userID).Count(&count).Error
	return count, err
}

// ClearAllPolicies removes all policies and role assignments
func (ca *CasbinAdapter) ClearAllPolicies() error {
	return ca.db.Table(ca.config.TableName).Where("1 = 1").Delete(&models.CasbinRule{}).Error
}

// ClearPolicies removes all policies (keeps role assignments)
func (ca *CasbinAdapter) ClearPolicies() error {
	return ca.db.Table(ca.config.TableName).Where("ptype = ?", "p").Delete(&models.CasbinRule{}).Error
}

// ClearRoles removes all role assignments (keeps policies)
func (ca *CasbinAdapter) ClearRoles() error {
	return ca.db.Table(ca.config.TableName).Where("ptype = ?", "g").Delete(&models.CasbinRule{}).Error
}

// GetDatabaseStats returns database statistics
func (ca *CasbinAdapter) GetDatabaseStats() (map[string]interface{}, error) {
	policyCount, err := ca.GetPolicyCount()
	if err != nil {
		return nil, err
	}

	roleCount, err := ca.GetRoleCount()
	if err != nil {
		return nil, err
	}

	// Get table size information
	var tableSize int64
	err = ca.db.Raw(`
		SELECT pg_total_relation_size(?) as size
	`, ca.config.TableName).Scan(&tableSize).Error
	if err != nil {
		logger.Warnf("Failed to get table size: %v", err)
		tableSize = -1
	}

	return map[string]interface{}{
		"total_policies":   policyCount,
		"total_roles":      roleCount,
		"table_name":       ca.config.TableName,
		"table_size_bytes": tableSize,
		"auto_migrate":     ca.config.AutoMigrate,
		"batch_size":       ca.config.BatchSize,
		"max_retries":      ca.config.MaxRetries,
		"cache_enabled":    ca.config.EnableCache,
		"cache_timeout":    ca.config.CacheTimeout,
	}, nil
}

// BackupPolicies creates a backup of all policies
func (ca *CasbinAdapter) BackupPolicies() ([][]string, error) {
	var rules []models.CasbinRule
	err := ca.db.Table(ca.config.TableName).Find(&rules).Error
	if err != nil {
		return nil, err
	}

	var policies [][]string
	for _, rule := range rules {
		policy := []string{rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5}
		// Remove empty fields
		for i := len(policy) - 1; i >= 0; i-- {
			if policy[i] == "" {
				policy = policy[:i]
			} else {
				break
			}
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// RestorePolicies restores policies from backup
func (ca *CasbinAdapter) RestorePolicies(policies [][]string) error {
	// Clear existing policies
	if err := ca.ClearAllPolicies(); err != nil {
		return err
	}

	// Insert new policies
	for _, policy := range policies {
		rule := models.CasbinRule{
			Ptype: "p",
			V0:    policy[0],
			V1:    policy[1],
			V2:    policy[2],
		}
		if len(policy) > 3 {
			rule.V3 = policy[3]
		}
		if len(policy) > 4 {
			rule.V4 = policy[4]
		}
		if len(policy) > 5 {
			rule.V5 = policy[5]
		}

		if err := ca.db.Table(ca.config.TableName).Create(&rule).Error; err != nil {
			return fmt.Errorf("failed to restore policy %v: %w", policy, err)
		}
	}

	return nil
}

// ValidatePolicy validates a policy before insertion
func (ca *CasbinAdapter) ValidatePolicy(policy []string) error {
	if len(policy) < 3 {
		return fmt.Errorf("policy must have at least 3 elements: subject, object, action")
	}

	// Check for empty required fields
	if policy[0] == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if policy[1] == "" {
		return fmt.Errorf("object cannot be empty")
	}
	if policy[2] == "" {
		return fmt.Errorf("action cannot be empty")
	}

	return nil
}

// GetAdapterConfig returns the current adapter configuration
func (ca *CasbinAdapter) GetAdapterConfig() *AdapterConfig {
	return ca.config
}
