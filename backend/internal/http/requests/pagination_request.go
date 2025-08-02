package requests

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PaginationRequest represents the standard pagination parameters for list endpoints
type PaginationRequest struct {
	Page    int    `json:"page" form:"page" query:"page" validate:"min=1" default:"1"`
	PerPage int    `json:"perPage" form:"perPage" query:"perPage" validate:"min=1,max=100" default:"10"`
	Search  string `json:"search,omitempty" form:"search" query:"search"`
}

// NewPaginationRequest creates a new PaginationRequest with default values
func NewPaginationRequest() *PaginationRequest {
	return &PaginationRequest{
		Page:    1,
		PerPage: 10,
	}
}

// ParseFromQuery parses pagination parameters from gin context query parameters
func (p *PaginationRequest) ParseFromQuery(c *gin.Context) error {
	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			p.Page = page
		}
	}

	// Parse perPage parameter
	if perPageStr := c.Query("perPage"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil {
			p.PerPage = perPage
		}
	}

	// Parse search parameter
	p.Search = c.Query("search")

	// Validate the parsed values
	validate := validator.New()
	return validate.Struct(p)
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// GetLimit returns the limit for database queries
func (p *PaginationRequest) GetLimit() int {
	return p.PerPage
}

// IsValid checks if the pagination parameters are valid
func (p *PaginationRequest) IsValid() bool {
	return p.Page >= 1 && p.PerPage >= 1 && p.PerPage <= 100
}

// ParsePaginationFromQuery is a helper function to parse pagination from gin context
// Returns the pagination request and any validation error
func ParsePaginationFromQuery(c *gin.Context) (*PaginationRequest, error) {
	pagination := NewPaginationRequest()
	err := pagination.ParseFromQuery(c)
	return pagination, err
}
