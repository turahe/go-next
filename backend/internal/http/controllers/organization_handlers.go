package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-next/internal/http/requests"
	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	OrganizationService *services.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{
		OrganizationService: services.NewOrganizationService(),
	}
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Create a new organization with the provided details
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body requests.OrganizationRequest true "Organization details"
// @Success 201 {object} responses.SuccessResponse{data=models.Organization}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations [post]
// @Security BearerAuth
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var input requests.OrganizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid request body",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	// Validate input
	validator := validation.NewValidator()
	result := validator.Validate(input)
	if !result.IsValid {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Validation failed",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   fmt.Sprintf("%v", result.Errors),
		})
		return
	}

	// Create organization
	org := &models.Organization{
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		Code:        input.Code,
		Type:        input.Type,
		ParentID:    input.ParentID,
	}

	if err := h.OrganizationService.CreateOrganization(org); err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to create organization",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse{
		Message:   "Organization created successfully",
		Data:      org,
		Status:    http.StatusCreated,
		Timestamp: time.Now(),
	})
}

// GetOrganization godoc
// @Summary Get organization by ID
// @Description Get organization details by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse{data=models.Organization}
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id} [get]
// @Security BearerAuth
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	org, err := h.OrganizationService.GetOrganizationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse{
			Message:   "Organization not found",
			Status:    http.StatusNotFound,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "Organization retrieved successfully",
		Data:      org,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// GetAllOrganizations godoc
// @Summary Get all organizations
// @Description Get all organizations with pagination
// @Tags organizations
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} responses.SuccessResponse{data=[]models.Organization}
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations [get]
// @Security BearerAuth
func (h *OrganizationHandler) GetAllOrganizations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orgs, err := h.OrganizationService.GetAllOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to get organizations",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	// Simple pagination
	start := (page - 1) * limit
	end := start + limit
	if start >= len(orgs) {
		start = len(orgs)
	}
	if end > len(orgs) {
		end = len(orgs)
	}

	paginatedOrgs := orgs[start:end]

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "Organizations retrieved successfully",
		Data:      paginatedOrgs,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// UpdateOrganization godoc
// @Summary Update organization
// @Description Update organization details
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param organization body requests.OrganizationRequest true "Organization details"
// @Success 200 {object} responses.SuccessResponse{data=models.Organization}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id} [put]
// @Security BearerAuth
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	var input requests.OrganizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid request body",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	// Validate input
	validator := validation.NewValidator()
	result := validator.Validate(input)
	if !result.IsValid {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Validation failed",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   fmt.Sprintf("%v", result.Errors),
		})
		return
	}

	// Get existing organization
	org, err := h.OrganizationService.GetOrganizationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse{
			Message:   "Organization not found",
			Status:    http.StatusNotFound,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	// Update fields
	org.Name = input.Name
	org.Slug = input.Slug
	org.Description = input.Description
	org.Code = input.Code
	org.Type = input.Type
	org.ParentID = input.ParentID

	if err := h.OrganizationService.UpdateOrganization(org); err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to update organization",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "Organization updated successfully",
		Data:      org,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// DeleteOrganization godoc
// @Summary Delete organization
// @Description Delete organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id} [delete]
// @Security BearerAuth
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	if err := h.OrganizationService.DeleteOrganization(id); err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to delete organization",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "Organization deleted successfully",
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// AddUserToOrganization godoc
// @Summary Add user to organization
// @Description Add a user to an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id}/users/{user_id} [post]
// @Security BearerAuth
func (h *OrganizationHandler) AddUserToOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid user ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	if err := h.OrganizationService.AddUserToOrganization(userID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to add user to organization",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "User added to organization successfully",
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// RemoveUserFromOrganization godoc
// @Summary Remove user from organization
// @Description Remove a user from an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id}/users/{user_id} [delete]
// @Security BearerAuth
func (h *OrganizationHandler) RemoveUserFromOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid user ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	if err := h.OrganizationService.RemoveUserFromOrganization(userID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to remove user from organization",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "User removed from organization successfully",
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// GetOrganizationPolicies godoc
// @Summary Get organization policies
// @Description Get all Casbin policies for an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse{data=[][]string}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id}/policies [get]
// @Security BearerAuth
func (h *OrganizationHandler) GetOrganizationPolicies(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	policies, err := h.OrganizationService.GetOrganizationPolicies(orgID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to get organization policies",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "Organization policies retrieved successfully",
		Data:      policies,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// GetUserRoleInOrganization godoc
// @Summary Get user role in organization
// @Description Get the role of a specific user in a specific organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse{data=string}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id}/users/{user_id}/role [get]
// @Security BearerAuth
func (h *OrganizationHandler) GetUserRoleInOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid user ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	role, err := h.OrganizationService.GetUserRoleInOrganization(userID, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse{
			Message:   "User role not found",
			Status:    http.StatusNotFound,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "User role retrieved successfully",
		Data:      role,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// UpdateUserRoleInOrganization godoc
// @Summary Update user role in organization
// @Description Update the role of a specific user in a specific organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param user_id path string true "User ID"
// @Param role body map[string]string true "Role information"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id}/users/{user_id}/role [put]
// @Security BearerAuth
func (h *OrganizationHandler) UpdateUserRoleInOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid user ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid request body",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	role, exists := input["role"]
	if !exists {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Role is required",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   "role field is required",
		})
		return
	}

	if err := h.OrganizationService.UpdateUserRoleInOrganization(userID, orgID, role); err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to update user role",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "User role updated successfully",
		Data:      nil,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}

// GetOrganizationUsersWithRoles godoc
// @Summary Get organization users with roles
// @Description Get all users in an organization with their roles
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse{data=[]models.OrganizationUser}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/organizations/{id}/users-with-roles [get]
// @Security BearerAuth
func (h *OrganizationHandler) GetOrganizationUsersWithRoles(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{
			Message:   "Invalid organization ID",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	orgUsers, err := h.OrganizationService.GetOrganizationUsersWithRoles(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
			Message:   "Failed to get organization users with roles",
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
			Details:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse{
		Message:   "Organization users with roles retrieved successfully",
		Data:      orgUsers,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})
}
