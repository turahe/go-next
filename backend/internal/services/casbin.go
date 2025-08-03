package services

import (
	"fmt"
	"go-next/pkg/casbin"
	"go-next/pkg/database"
	"go-next/pkg/logger"
	"os"

	casbinlib "github.com/casbin/casbin/v2"
	"github.com/google/uuid"
)

var Enforcer *casbinlib.Enforcer

// CasbinService provides Casbin RBAC functionality
type CasbinService struct {
	enforcer *casbinlib.Enforcer
}

// NewCasbinService creates a new Casbin service
func NewCasbinService() *CasbinService {
	return &CasbinService{
		enforcer: Enforcer,
	}
}

// InitCasbin initializes Casbin with custom database adapter
func InitCasbin() error {
	// Load RBAC model configuration using model loader
	modelLoader := casbin.NewModelLoader("config")

	// Get environment from environment variable, default to empty (uses default model)
	environment := os.Getenv("ENVIRONMENT")

	mconf, err := modelLoader.LoadModel(environment)
	if err != nil {
		return fmt.Errorf("failed to load RBAC model configuration: %w", err)
	}

	// Validate the model configuration
	if err := modelLoader.ValidateModel(mconf); err != nil {
		return fmt.Errorf("invalid RBAC model configuration: %w", err)
	}

	// Create custom GORM adapter using existing CasbinRule model
	adapter := casbin.NewCustomGormAdapter(database.DB, "casbin_rule")

	// Create enforcer with custom adapter
	e, err := adapter.CreateEnforcer(mconf)
	if err != nil {
		return fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	Enforcer = e

	// Initialize default policies
	if err := initializeDefaultPolicies(e); err != nil {
		return fmt.Errorf("failed to initialize default policies: %w", err)
	}

	logger.Infof("Casbin initialized successfully with custom GORM adapter (environment: %s)", environment)
	return nil
}

// initializeDefaultPolicies sets up default RBAC policies
func initializeDefaultPolicies(e *casbinlib.Enforcer) error {
	// Check if policies already exist
	policies, err := e.GetPolicy()
	if err != nil {
		return err
	}

	// Only add default policies if none exist
	if len(policies) == 0 {
		// Admin policies - full access (global domain)
		adminPolicies := [][]string{
			{"admin", "*", "/api/users", "GET"},
			{"admin", "*", "/api/users", "POST"},
			{"admin", "*", "/api/users", "PUT"},
			{"admin", "*", "/api/users", "DELETE"},
			{"admin", "*", "/api/roles", "GET"},
			{"admin", "*", "/api/roles", "POST"},
			{"admin", "*", "/api/roles", "PUT"},
			{"admin", "*", "/api/roles", "DELETE"},
			{"admin", "*", "/api/posts", "GET"},
			{"admin", "*", "/api/posts", "POST"},
			{"admin", "*", "/api/posts", "PUT"},
			{"admin", "*", "/api/posts", "DELETE"},
			{"admin", "*", "/api/categories", "GET"},
			{"admin", "*", "/api/categories", "POST"},
			{"admin", "*", "/api/categories", "PUT"},
			{"admin", "*", "/api/categories", "DELETE"},
			{"admin", "*", "/api/comments", "GET"},
			{"admin", "*", "/api/comments", "POST"},
			{"admin", "*", "/api/comments", "PUT"},
			{"admin", "*", "/api/comments", "DELETE"},
			{"admin", "*", "/api/media", "GET"},
			{"admin", "*", "/api/media", "POST"},
			{"admin", "*", "/api/media", "PUT"},
			{"admin", "*", "/api/media", "DELETE"},
			{"admin", "*", "/api/casbin/policies", "GET"},
			{"admin", "*", "/api/casbin/policies", "POST"},
			{"admin", "*", "/api/casbin/policies", "DELETE"},
			{"admin", "*", "/api/casbin/users", "GET"},
			{"admin", "*", "/api/casbin/users", "POST"},
			{"admin", "*", "/api/casbin/users", "DELETE"},
			{"admin", "*", "/api/organizations", "GET"},
			{"admin", "*", "/api/organizations", "POST"},
			{"admin", "*", "/api/organizations", "PUT"},
			{"admin", "*", "/api/organizations", "DELETE"},
		}

		// Editor policies - content management (domain-specific)
		editorPolicies := [][]string{
			{"editor", "*", "/api/posts", "GET"},
			{"editor", "*", "/api/posts", "POST"},
			{"editor", "*", "/api/posts", "PUT"},
			{"editor", "*", "/api/posts", "DELETE"},
			{"editor", "*", "/api/categories", "GET"},
			{"editor", "*", "/api/comments", "GET"},
			{"editor", "*", "/api/comments", "POST"},
			{"editor", "*", "/api/comments", "PUT"},
			{"editor", "*", "/api/comments", "DELETE"},
			{"editor", "*", "/api/media", "GET"},
			{"editor", "*", "/api/media", "POST"},
		}

		// Moderator policies - comment moderation
		moderatorPolicies := [][]string{
			{"moderator", "*", "/api/posts", "GET"},
			{"moderator", "*", "/api/categories", "GET"},
			{"moderator", "*", "/api/comments", "GET"},
			{"moderator", "*", "/api/comments", "POST"},
			{"moderator", "*", "/api/comments", "PUT"},
			{"moderator", "*", "/api/comments", "DELETE"},
		}

		// User policies - basic access
		userPolicies := [][]string{
			{"user", "*", "/api/posts", "GET"},
			{"user", "*", "/api/categories", "GET"},
			{"user", "*", "/api/comments", "GET"},
			{"user", "*", "/api/comments", "POST"},
			{"user", "*", "/api/media", "GET"},
		}

		// Guest policies - read-only access
		guestPolicies := [][]string{
			{"guest", "/api/posts", "GET"},
			{"guest", "/api/categories", "GET"},
		}

		// Add all policies
		allPolicies := append(adminPolicies, editorPolicies...)
		allPolicies = append(allPolicies, moderatorPolicies...)
		allPolicies = append(allPolicies, userPolicies...)
		allPolicies = append(allPolicies, guestPolicies...)

		for _, policy := range allPolicies {
			if _, err := e.AddPolicy(policy); err != nil {
				return fmt.Errorf("failed to add policy %v: %w", policy, err)
			}
		}

		// Save policies to database
		if err := e.SavePolicy(); err != nil {
			return fmt.Errorf("failed to save policies: %w", err)
		}

		logger.Infof("Initialized %d default Casbin policies", len(allPolicies))
	}

	return nil
}

// GetEnforcer returns the Casbin enforcer
func (cs *CasbinService) GetEnforcer() *casbinlib.Enforcer {
	return cs.enforcer
}

// AddPolicy adds a new policy with domain context
func (cs *CasbinService) AddPolicy(subject, domain, object, action string) error {
	_, err := cs.enforcer.AddPolicy(subject, domain, object, action)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}
	return cs.enforcer.SavePolicy()
}

// RemovePolicy removes a policy with domain context
func (cs *CasbinService) RemovePolicy(subject, domain, object, action string) error {
	_, err := cs.enforcer.RemovePolicy(subject, domain, object, action)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}
	return cs.enforcer.SavePolicy()
}

// AddRoleForUser adds a role for a user with domain context
func (cs *CasbinService) AddRoleForUser(userID uuid.UUID, role, domain string) error {
	_, err := cs.enforcer.AddGroupingPolicy(userID.String(), role, domain)
	if err != nil {
		return fmt.Errorf("failed to add role for user: %w", err)
	}
	return cs.enforcer.SavePolicy()
}

// RemoveRoleForUser removes a role from a user with domain context
func (cs *CasbinService) RemoveRoleForUser(userID uuid.UUID, role, domain string) error {
	_, err := cs.enforcer.RemoveFilteredGroupingPolicy(0, userID.String(), role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return cs.enforcer.SavePolicy()
}

// GetUserRoles gets all roles for a user with domain context
func (cs *CasbinService) GetUserRoles(userID uuid.UUID) ([]string, error) {
	return cs.enforcer.GetRolesForUser(userID.String())
}

// GetUserRolesInDomain gets all roles for a user in a specific domain
func (cs *CasbinService) GetUserRolesInDomain(userID uuid.UUID, domain string) ([]string, error) {
	policies, err := cs.enforcer.GetFilteredGroupingPolicy(0, userID.String(), "", domain)
	if err != nil {
		return nil, err
	}

	var roles []string
	for _, policy := range policies {
		if len(policy) >= 2 {
			roles = append(roles, policy[1]) // role is at index 1
		}
	}
	return roles, nil
}

// Enforce checks if a user has permission
func (cs *CasbinService) Enforce(userID uuid.UUID, object, action string) (bool, error) {
	// Get user's roles
	roles, err := cs.GetUserRoles(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check each role
	for _, role := range roles {
		allowed, err := cs.enforcer.Enforce(role, "*", object, action) // Use wildcard for domain
		if err != nil {
			return false, fmt.Errorf("failed to enforce policy: %w", err)
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// EnforceWithDomain checks if a user has permission in a specific domain
func (cs *CasbinService) EnforceWithDomain(userID uuid.UUID, object, action, domain string) (bool, error) {
	// Get user's roles in the specific domain
	roles, err := cs.GetUserRolesInDomain(userID, domain)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles in domain: %w", err)
	}

	// Check each role
	for _, role := range roles {
		allowed, err := cs.enforcer.Enforce(role, domain, object, action)
		if err != nil {
			return false, fmt.Errorf("failed to enforce policy: %w", err)
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// GetAllPolicies gets all policies
func (cs *CasbinService) GetAllPolicies() ([][]string, error) {
	return cs.enforcer.GetPolicy()
}

// GetFilteredPolicies gets filtered policies
func (cs *CasbinService) GetFilteredPolicies(fieldIndex int, fieldValues ...string) ([][]string, error) {
	return cs.enforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
}
