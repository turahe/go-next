package services

import (
	"fmt"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationService provides organization management functionality
type OrganizationService struct {
	db *gorm.DB
}

// NewOrganizationService creates a new organization service
func NewOrganizationService() *OrganizationService {
	return &OrganizationService{
		db: database.DB,
	}
}

// CreateOrganization creates a new organization
func (os *OrganizationService) CreateOrganization(org *models.Organization) error {
	if err := os.db.Create(org).Error; err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	return nil
}

// GetOrganizationByID gets an organization by ID
func (os *OrganizationService) GetOrganizationByID(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	if err := os.db.Preload("Children").Preload("Users").First(&org, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &org, nil
}

// GetOrganizationBySlug gets an organization by slug
func (os *OrganizationService) GetOrganizationBySlug(slug string) (*models.Organization, error) {
	var org models.Organization
	if err := os.db.Preload("Children").Preload("Users").Where("slug = ?", slug).First(&org).Error; err != nil {
		return nil, fmt.Errorf("failed to get organization by slug: %w", err)
	}
	return &org, nil
}

// GetAllOrganizations gets all organizations
func (os *OrganizationService) GetAllOrganizations() ([]models.Organization, error) {
	var orgs []models.Organization
	if err := os.db.Preload("Children").Preload("Users").Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	return orgs, nil
}

// UpdateOrganization updates an organization
func (os *OrganizationService) UpdateOrganization(org *models.Organization) error {
	if err := os.db.Save(org).Error; err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

// DeleteOrganization deletes an organization
func (os *OrganizationService) DeleteOrganization(id uuid.UUID) error {
	if err := os.db.Delete(&models.Organization{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

// AddUserToOrganization adds a user to an organization
func (os *OrganizationService) AddUserToOrganization(userID, orgID uuid.UUID) error {
	// Check if user is already in organization
	var existing models.OrganizationUser
	if err := os.db.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&existing).Error; err == nil {
		return fmt.Errorf("user is already a member of this organization")
	}

	// Add user to organization using the join table model
	orgUser := &models.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "member", // Default role
		IsActive:       true,
	}

	if err := os.db.Create(orgUser).Error; err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Add user role in Casbin for the organization
	casbinService := NewCasbinService()
	if err := casbinService.AddRoleForUser(userID, "user", orgID.String()); err != nil {
		logger.Warnf("Failed to add user role in Casbin for organization: %v", err)
	}

	return nil
}

// RemoveUserFromOrganization removes a user from an organization
func (os *OrganizationService) RemoveUserFromOrganization(userID, orgID uuid.UUID) error {
	// Remove user from organization using the join table model
	if err := os.db.Where("user_id = ? AND organization_id = ?", userID, orgID).Delete(&models.OrganizationUser{}).Error; err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}

	// Remove user role in Casbin for the organization
	casbinService := NewCasbinService()
	if err := casbinService.RemoveRoleForUser(userID, "user", orgID.String()); err != nil {
		logger.Warnf("Failed to remove user role in Casbin for organization: %v", err)
	}

	return nil
}

// GetUsersInOrganization gets all users in an organization
func (os *OrganizationService) GetUsersInOrganization(orgID uuid.UUID) ([]models.User, error) {
	var users []models.User
	if err := os.db.Joins("JOIN organization_users ON users.id = organization_users.user_id").
		Where("organization_users.organization_id = ? AND organization_users.is_active = ?", orgID, true).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users in organization: %w", err)
	}
	return users, nil
}

// GetOrganizationsForUser gets all organizations a user belongs to
func (os *OrganizationService) GetOrganizationsForUser(userID uuid.UUID) ([]models.Organization, error) {
	var orgs []models.Organization
	if err := os.db.Joins("JOIN organization_users ON organizations.id = organization_users.organization_id").
		Where("organization_users.user_id = ? AND organization_users.is_active = ?", userID, true).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to get organizations for user: %w", err)
	}
	return orgs, nil
}

// AddOrganizationPolicy adds a policy for an organization
func (os *OrganizationService) AddOrganizationPolicy(role, object, action, orgID string) error {
	casbinService := NewCasbinService()
	return casbinService.AddPolicy(role, orgID, object, action)
}

// RemoveOrganizationPolicy removes a policy for an organization
func (os *OrganizationService) RemoveOrganizationPolicy(role, object, action, orgID string) error {
	casbinService := NewCasbinService()
	return casbinService.RemovePolicy(role, orgID, object, action)
}

// GetOrganizationPolicies gets all policies for an organization
func (os *OrganizationService) GetOrganizationPolicies(orgID string) ([][]string, error) {
	casbinService := NewCasbinService()
	return casbinService.GetFilteredPolicies(1, orgID) // domain is at index 1
}

// GetUserRoleInOrganization gets the role of a user in a specific organization
func (os *OrganizationService) GetUserRoleInOrganization(userID, orgID uuid.UUID) (string, error) {
	var orgUser models.OrganizationUser
	if err := os.db.Where("user_id = ? AND organization_id = ? AND is_active = ?", userID, orgID, true).First(&orgUser).Error; err != nil {
		return "", fmt.Errorf("user not found in organization: %w", err)
	}
	return orgUser.Role, nil
}

// UpdateUserRoleInOrganization updates the role of a user in a specific organization
func (os *OrganizationService) UpdateUserRoleInOrganization(userID, orgID uuid.UUID, role string) error {
	if err := os.db.Model(&models.OrganizationUser{}).
		Where("user_id = ? AND organization_id = ? AND is_active = ?", userID, orgID, true).
		Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update user role in organization: %w", err)
	}
	return nil
}

// DeactivateUserInOrganization deactivates a user in an organization (soft delete)
func (os *OrganizationService) DeactivateUserInOrganization(userID, orgID uuid.UUID) error {
	if err := os.db.Model(&models.OrganizationUser{}).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate user in organization: %w", err)
	}
	return nil
}

// GetOrganizationUsersWithRoles gets all users in an organization with their roles
func (os *OrganizationService) GetOrganizationUsersWithRoles(orgID uuid.UUID) ([]models.OrganizationUser, error) {
	var orgUsers []models.OrganizationUser
	if err := os.db.Where("organization_id = ? AND is_active = ?", orgID, true).
		Preload("User").
		Find(&orgUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to get organization users with roles: %w", err)
	}
	return orgUsers, nil
}

// Nested Set Model methods for Organization

// GetOrganizationDescendants gets all descendants of an organization using nested set model
func (os *OrganizationService) GetOrganizationDescendants(id uuid.UUID) ([]models.Organization, error) {
	var org models.Organization
	if err := os.db.First(&org, id).Error; err != nil {
		return nil, err
	}

	var descendants []models.Organization
	err := os.db.Where("record_left > ? AND record_right < ?", org.RecordLeft, org.RecordRight).
		Order("record_left ASC").Find(&descendants).Error
	return descendants, err
}

// GetOrganizationAncestors gets all ancestors of an organization using nested set model
func (os *OrganizationService) GetOrganizationAncestors(id uuid.UUID) ([]models.Organization, error) {
	var org models.Organization
	if err := os.db.First(&org, id).Error; err != nil {
		return nil, err
	}

	var ancestors []models.Organization
	err := os.db.Where("record_left < ? AND record_right > ?", org.RecordLeft, org.RecordRight).
		Order("record_left ASC").Find(&ancestors).Error
	return ancestors, err
}

// GetOrganizationSiblings gets all siblings of an organization using nested set model
func (os *OrganizationService) GetOrganizationSiblings(id uuid.UUID) ([]models.Organization, error) {
	var org models.Organization
	if err := os.db.First(&org, id).Error; err != nil {
		return nil, err
	}

	var siblings []models.Organization
	err := os.db.Where("record_left > ? AND record_right < ? AND record_dept = ?",
		org.RecordLeft, org.RecordRight, org.RecordDept).
		Order("record_left ASC").Find(&siblings).Error
	return siblings, err
}

// MoveOrganization moves an organization and its subtree to a new parent using nested set model
func (os *OrganizationService) MoveOrganization(id uuid.UUID, newParentID uuid.UUID) error {
	return os.db.Transaction(func(tx *gorm.DB) error {
		var org models.Organization
		if err := tx.First(&org, id).Error; err != nil {
			return err
		}

		return os.moveOrganizationSubtree(tx, &org, newParentID)
	})
}

// CreateOrganizationNested creates a new organization with nested set model
func (os *OrganizationService) CreateOrganizationNested(org *models.Organization, parentID *uuid.UUID) error {
	return os.db.Transaction(func(tx *gorm.DB) error {
		if parentID != nil {
			var parent models.Organization
			if err := tx.First(&parent, parentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == 0 {
				return gorm.ErrRecordNotFound
			}

			// Update parent's right value
			tx.Model(&models.Organization{}).
				Where("record_right >= ?", parent.RecordRight).
				Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Organization{}).
				Where("record_left > ?", parent.RecordRight).
				Update("record_left", gorm.Expr("record_left + 2"))

			org.RecordLeft = parent.RecordRight
			org.RecordRight = parent.RecordRight + 1
			org.RecordDept = parent.RecordDept + 1
			org.ParentID = *parentID
		} else {
			// Create as root
			var maxRight int
			tx.Model(&models.Organization{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)

			org.RecordLeft = maxRight + 1
			org.RecordRight = maxRight + 2
			org.RecordDept = 0
		}

		return tx.Create(org).Error
	})
}

// DeleteOrganizationNested deletes an organization and its entire subtree using nested set model
func (os *OrganizationService) DeleteOrganizationNested(id uuid.UUID) error {
	return os.db.Transaction(func(tx *gorm.DB) error {
		var org models.Organization
		if err := tx.First(&org, id).Error; err != nil {
			return err
		}

		// Calculate the width of the subtree
		width := org.RecordRight - org.RecordLeft + 1

		// Delete the organization and all its descendants
		if err := tx.Where("record_left >= ? AND record_right <= ?", org.RecordLeft, org.RecordRight).Delete(&models.Organization{}).Error; err != nil {
			return err
		}

		// Update the left and right values of remaining nodes
		tx.Model(&models.Organization{}).
			Where("record_left > ?", org.RecordRight).
			Update("record_left", gorm.Expr("record_left - ?", width))

		tx.Model(&models.Organization{}).
			Where("record_right > ?", org.RecordRight).
			Update("record_right", gorm.Expr("record_right - ?", width))

		return nil
	})
}

// moveOrganizationSubtree moves an organization and its entire subtree to a new parent
func (os *OrganizationService) moveOrganizationSubtree(tx *gorm.DB, org *models.Organization, newParentID uuid.UUID) error {
	// Calculate the width of the subtree
	width := org.RecordRight - org.RecordLeft + 1

	// Get the new parent
	var newParent models.Organization
	if newParentID != uuid.Nil {
		if err := tx.First(&newParent, newParentID).Error; err != nil {
			return err
		}
	}

	// Calculate the new position
	var newLeft int
	if newParentID != uuid.Nil {
		newLeft = newParent.RecordRight
	} else {
		// Moving to root level
		var maxRight int
		tx.Model(&models.Organization{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)
		newLeft = maxRight + 1
	}

	// Calculate the offset
	offset := newLeft - org.RecordLeft

	// Update all nodes in the subtree
	tx.Model(&models.Organization{}).
		Where("record_left >= ? AND record_right <= ?", org.RecordLeft, org.RecordRight).
		Updates(map[string]interface{}{
			"record_left":  gorm.Expr("record_left + ?", offset),
			"record_right": gorm.Expr("record_right + ?", offset),
			"record_dept":  gorm.Expr("record_dept + ?", newParent.RecordDept-org.RecordDept+1),
		})

	// Update nodes to the right of the old position
	tx.Model(&models.Organization{}).
		Where("record_left > ?", org.RecordRight).
		Update("record_left", gorm.Expr("record_left - ?", width))

	tx.Model(&models.Organization{}).
		Where("record_right > ?", org.RecordRight).
		Update("record_right", gorm.Expr("record_right - ?", width))

	// Update nodes to the right of the new position
	tx.Model(&models.Organization{}).
		Where("record_left >= ?", newLeft).
		Update("record_left", gorm.Expr("record_left + ?", width))

	tx.Model(&models.Organization{}).
		Where("record_right >= ?", newLeft).
		Update("record_right", gorm.Expr("record_right + ?", width))

	return nil
}
