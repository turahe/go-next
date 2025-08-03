package casbin

import (
	"fmt"
	"go-next/internal/models"
	"go-next/pkg/logger"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// CustomGormAdapter is a custom GORM adapter that uses the existing CasbinRule model
type CustomGormAdapter struct {
	db        *gorm.DB
	tableName string
}

// NewCustomGormAdapter creates a new custom GORM adapter
func NewCustomGormAdapter(db *gorm.DB, tableName string) *CustomGormAdapter {
	return &CustomGormAdapter{
		db:        db,
		tableName: tableName,
	}
}

// LoadPolicy loads all policy rules from the storage.
func (a *CustomGormAdapter) LoadPolicy(model model.Model) error {
	var rules []models.CasbinRule
	if err := a.db.Table(a.tableName).Find(&rules).Error; err != nil {
		return err
	}

	for _, rule := range rules {
		line := rule.Ptype
		if rule.V0 != "" {
			line += ", " + rule.V0
		}
		if rule.V1 != "" {
			line += ", " + rule.V1
		}
		if rule.V2 != "" {
			line += ", " + rule.V2
		}
		if rule.V3 != "" {
			line += ", " + rule.V3
		}
		if rule.V4 != "" {
			line += ", " + rule.V4
		}
		if rule.V5 != "" {
			line += ", " + rule.V5
		}

		persist.LoadPolicyLine(line, model)
	}

	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *CustomGormAdapter) SavePolicy(model model.Model) error {
	// Clear existing policies
	if err := a.db.Table(a.tableName).Where("1 = 1").Delete(&models.CasbinRule{}).Error; err != nil {
		return err
	}

	// Save policies
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := fmt.Sprintf("%s, %s", ptype, rule[0])
			for i := 1; i < len(rule); i++ {
				line += ", " + rule[i]
			}
			if err := a.savePolicyLine(line); err != nil {
				return err
			}
		}
	}

	// Save role inheritance rules
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := fmt.Sprintf("%s, %s", ptype, rule[0])
			for i := 1; i < len(rule); i++ {
				line += ", " + rule[i]
			}
			if err := a.savePolicyLine(line); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *CustomGormAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := fmt.Sprintf("%s, %s", ptype, rule[0])
	for i := 1; i < len(rule); i++ {
		line += ", " + rule[i]
	}
	return a.savePolicyLine(line)
}

// RemovePolicy removes a policy rule from the storage.
func (a *CustomGormAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := fmt.Sprintf("%s, %s", ptype, rule[0])
	for i := 1; i < len(rule); i++ {
		line += ", " + rule[i]
	}
	return a.removePolicyLine(line)
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *CustomGormAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	query := a.db.Table(a.tableName).Where("ptype = ?", ptype)
	for i, fieldValue := range fieldValues {
		if fieldValue != "" {
			query = query.Where(fmt.Sprintf("v%d = ?", i), fieldValue)
		}
	}
	return query.Delete(&models.CasbinRule{}).Error
}

// savePolicyLine saves a policy line to the database
func (a *CustomGormAdapter) savePolicyLine(line string) error {
	rule := a.parsePolicyLine(line)
	return a.db.Table(a.tableName).Create(&rule).Error
}

// removePolicyLine removes a policy line from the database
func (a *CustomGormAdapter) removePolicyLine(line string) error {
	rule := a.parsePolicyLine(line)
	return a.db.Table(a.tableName).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
		rule.Ptype, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5).Delete(&models.CasbinRule{}).Error
}

// parsePolicyLine parses a policy line into a CasbinRule struct
func (a *CustomGormAdapter) parsePolicyLine(line string) models.CasbinRule {
	rule := models.CasbinRule{}
	parts := strings.Split(line, ", ")
	if len(parts) > 0 {
		rule.Ptype = parts[0]
	}
	if len(parts) > 1 {
		rule.V0 = parts[1]
	}
	if len(parts) > 2 {
		rule.V1 = parts[2]
	}
	if len(parts) > 3 {
		rule.V2 = parts[3]
	}
	if len(parts) > 4 {
		rule.V3 = parts[4]
	}
	if len(parts) > 5 {
		rule.V4 = parts[5]
	}
	if len(parts) > 6 {
		rule.V5 = parts[6]
	}
	return rule
}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *CustomGormAdapter) IsFiltered() bool {
	return false
}

// UpdatePolicy updates a policy rule from storage.
func (a *CustomGormAdapter) UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	oldLine := fmt.Sprintf("%s, %s", ptype, oldRule[0])
	for i := 1; i < len(oldRule); i++ {
		oldLine += ", " + oldRule[i]
	}

	newLine := fmt.Sprintf("%s, %s", ptype, newRule[0])
	for i := 1; i < len(newRule); i++ {
		newLine += ", " + newRule[i]
	}

	oldRuleStruct := a.parsePolicyLine(oldLine)
	newRuleStruct := a.parsePolicyLine(newLine)

	return a.db.Table(a.tableName).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
		oldRuleStruct.Ptype, oldRuleStruct.V0, oldRuleStruct.V1, oldRuleStruct.V2, oldRuleStruct.V3, oldRuleStruct.V4, oldRuleStruct.V5).
		Updates(&newRuleStruct).Error
}

// UpdatePolicies updates some policy rules to storage, like db, redis.
func (a *CustomGormAdapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	for i, oldRule := range oldRules {
		if i < len(newRules) {
			if err := a.UpdatePolicy(sec, ptype, oldRule, newRules[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (a *CustomGormAdapter) UpdateFilteredPolicies(sec string, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) (bool, error) {
	// Remove filtered policies
	if err := a.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
		return false, err
	}

	// Add new policies
	for _, policy := range newPolicies {
		if err := a.AddPolicy(sec, ptype, policy); err != nil {
			return false, err
		}
	}

	return true, nil
}

// CreateEnforcer creates a Casbin enforcer with the custom adapter
func (a *CustomGormAdapter) CreateEnforcer(modelConfig string) (*casbin.Enforcer, error) {
	// Parse model configuration
	m, err := model.NewModelFromString(modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin model: %w", err)
	}

	// Create enforcer with custom adapter
	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Enable auto-save
	e.EnableAutoSave(true)

	logger.Infof("Casbin enforcer created successfully with custom GORM adapter")
	return e, nil
}

// GetPolicyCount returns the total number of policies
func (a *CustomGormAdapter) GetPolicyCount() (int64, error) {
	var count int64
	err := a.db.Table(a.tableName).Count(&count).Error
	return count, err
}

// GetRoleCount returns the total number of role assignments
func (a *CustomGormAdapter) GetRoleCount() (int64, error) {
	var count int64
	err := a.db.Table(a.tableName).Where("ptype = ?", "g").Count(&count).Error
	return count, err
}

// GetPolicyCountByRole returns the number of policies for a specific role
func (a *CustomGormAdapter) GetPolicyCountByRole(role string) (int64, error) {
	var count int64
	err := a.db.Table(a.tableName).Where("ptype = ? AND v0 = ?", "p", role).Count(&count).Error
	return count, err
}

// GetUserRoleCount returns the number of role assignments for a user
func (a *CustomGormAdapter) GetUserRoleCount(userID string) (int64, error) {
	var count int64
	err := a.db.Table(a.tableName).Where("ptype = ? AND v1 = ?", "g", userID).Count(&count).Error
	return count, err
}

// ClearAllPolicies removes all policies and role assignments
func (a *CustomGormAdapter) ClearAllPolicies() error {
	return a.db.Table(a.tableName).Where("1 = 1").Delete(&models.CasbinRule{}).Error
}

// ClearPolicies removes all policies (keeps role assignments)
func (a *CustomGormAdapter) ClearPolicies() error {
	return a.db.Table(a.tableName).Where("ptype = ?", "p").Delete(&models.CasbinRule{}).Error
}

// ClearRoles removes all role assignments (keeps policies)
func (a *CustomGormAdapter) ClearRoles() error {
	return a.db.Table(a.tableName).Where("ptype = ?", "g").Delete(&models.CasbinRule{}).Error
}

// GetDatabaseStats returns database statistics
func (a *CustomGormAdapter) GetDatabaseStats() (map[string]interface{}, error) {
	policyCount, err := a.GetPolicyCount()
	if err != nil {
		return nil, err
	}

	roleCount, err := a.GetRoleCount()
	if err != nil {
		return nil, err
	}

	// Get table size information
	var tableSize int64
	err = a.db.Raw(`
		SELECT pg_total_relation_size(?) as size
	`, a.tableName).Scan(&tableSize).Error
	if err != nil {
		logger.Warnf("Failed to get table size: %v", err)
		tableSize = -1
	}

	return map[string]interface{}{
		"total_policies":   policyCount,
		"total_roles":      roleCount,
		"table_name":       a.tableName,
		"table_size_bytes": tableSize,
		"adapter_type":     "custom_gorm",
	}, nil
}

// BackupPolicies creates a backup of all policies
func (a *CustomGormAdapter) BackupPolicies() ([][]string, error) {
	var rules []models.CasbinRule
	err := a.db.Table(a.tableName).Find(&rules).Error
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
func (a *CustomGormAdapter) RestorePolicies(policies [][]string) error {
	// Clear existing policies
	if err := a.ClearAllPolicies(); err != nil {
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

		if err := a.db.Table(a.tableName).Create(&rule).Error; err != nil {
			return fmt.Errorf("failed to restore policy %v: %w", policy, err)
		}
	}

	return nil
}

// ValidatePolicy validates a policy before insertion
func (a *CustomGormAdapter) ValidatePolicy(policy []string) error {
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
