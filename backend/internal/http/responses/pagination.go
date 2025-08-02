package responses

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationLinks represents the navigation links in Laravel-style pagination
type PaginationLinks struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

// PaginationMeta represents the metadata in Laravel-style pagination
type PaginationMeta struct {
	CurrentPage int64 `json:"current_page"`
	From        int64 `json:"from"`
	LastPage    int64 `json:"last_page"`
	PerPage     int64 `json:"per_page"`
	To          int64 `json:"to"`
	Total       int64 `json:"total"`
}

// LaravelPaginationResponse represents the complete Laravel-style pagination response
type LaravelPaginationResponse struct {
	Data  interface{}     `json:"data"`
	Links PaginationLinks `json:"links"`
	Meta  PaginationMeta  `json:"meta"`
}

// PaginationParams holds pagination query parameters
type PaginationParams struct {
	Page    int
	PerPage int
}

// PaginateResult holds paginated data and meta info
type PaginateResult struct {
	Data         interface{}
	TotalCount   int64
	TotalPage    int64
	CurrentPage  int64
	LastPage     int64
	PerPage      int64
	NextPage     int64
	PreviousPage int64
}

// CreateLaravelPaginationResponse creates a Laravel-style pagination response
func CreateLaravelPaginationResponse(c *gin.Context, data interface{}, total int64, currentPage, perPage int64) LaravelPaginationResponse {
	lastPage := (total + perPage - 1) / perPage
	if lastPage < 1 {
		lastPage = 1
	}

	// Calculate from and to
	from := (currentPage-1)*perPage + 1
	if from > total {
		from = 0
	}
	to := currentPage * perPage
	if to > total {
		to = total
	}

	// Build base URL
	baseURL := getBaseURL(c)

	// Create links
	links := PaginationLinks{
		First: buildPaginationURL(baseURL, c.Request.URL.Query(), 1),
		Last:  buildPaginationURL(baseURL, c.Request.URL.Query(), lastPage),
	}

	// Add prev/next links
	if currentPage > 1 {
		links.Prev = buildPaginationURL(baseURL, c.Request.URL.Query(), currentPage-1)
	}
	if currentPage < lastPage {
		links.Next = buildPaginationURL(baseURL, c.Request.URL.Query(), currentPage+1)
	}

	// Create meta
	meta := PaginationMeta{
		CurrentPage: currentPage,
		From:        from,
		LastPage:    lastPage,
		PerPage:     perPage,
		To:          to,
		Total:       total,
	}

	return LaravelPaginationResponse{
		Data:  data,
		Links: links,
		Meta:  meta,
	}
}

// SendLaravelPagination sends a Laravel-style pagination response
func SendLaravelPagination(c *gin.Context, data interface{}, total int64, currentPage, perPage int64) {
	response := CreateLaravelPaginationResponse(c, data, total, currentPage, perPage)
	c.JSON(http.StatusOK, response)
}

// SendLaravelPaginationWithMessage sends a Laravel-style pagination response with a custom message
func SendLaravelPaginationWithMessage(c *gin.Context, message string, data interface{}, total int64, currentPage, perPage int64) {
	response := CreateLaravelPaginationResponse(c, data, total, currentPage, perPage)

	// Add message to response
	fullResponse := gin.H{
		"message": message,
		"data":    response.Data,
		"links":   response.Links,
		"meta":    response.Meta,
	}

	c.JSON(http.StatusOK, fullResponse)
}

// getBaseURL extracts the base URL from the request
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	host := c.Request.Host
	if host == "" {
		host = c.Request.Header.Get("Host")
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, c.Request.URL.Path)
}

// buildPaginationURL builds a pagination URL with the given page number
func buildPaginationURL(baseURL string, query url.Values, page int64) string {
	// Create a copy of the query parameters
	newQuery := make(url.Values)
	for key, values := range query {
		newQuery[key] = values
	}

	// Set the page parameter
	newQuery.Set("page", strconv.FormatInt(page, 10))

	// Build the URL
	if len(newQuery) > 0 {
		return baseURL + "?" + newQuery.Encode()
	}
	return baseURL
}

// ParsePaginationParams parses pagination parameters from the request
func ParsePaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))

	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}
	if perPage > 100 {
		perPage = 100
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}

// Legacy PaginationResponse for backward compatibility
type PaginationResponse struct {
	Data         any   `json:"data"`
	TotalCount   int64 `json:"totalCount"`
	TotalPage    int64 `json:"totalPage"`
	CurrentPage  int64 `json:"currentPage"`
	LastPage     int64 `json:"lastPage"`
	PerPage      int64 `json:"perPage"`
	NextPage     int64 `json:"nextPage"`
	PreviousPage int64 `json:"previousPage"`
}
