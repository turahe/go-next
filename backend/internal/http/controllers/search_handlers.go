package controllers

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-next/internal/http/responses"
	"go-next/internal/services"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
	BaseHandler
	searchService *services.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search performs a search across all indexes
// @Summary Search across all content
// @Description Search across posts, users, categories, and media
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Results per page (default: 20)"
// @Param indexes query string false "Comma-separated list of indexes to search (posts,users,categories,media)"
// @Param filters query string false "JSON string of filters"
// @Param sort_by query string false "Comma-separated list of sort fields"
// @Param highlight query bool false "Enable result highlighting"
// @Success 200 {object} responses.Response{data=services.SearchResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	// Parse query parameters
	query := c.Query("query")
	if query == "" {
		responses.SendError(c, 400, "Query parameter is required")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Parse indexes
	indexesStr := c.Query("indexes")
	var indexes []string
	if indexesStr != "" {
		indexes = strings.Split(indexesStr, ",")
	}

	// Parse filters
	filtersStr := c.Query("filters")
	var filters map[string]interface{}
	if filtersStr != "" {
		if err := json.Unmarshal([]byte(filtersStr), &filters); err != nil {
			responses.SendError(c, 400, "Invalid filters format")
			return
		}
	}

	// Parse sort by
	sortByStr := c.Query("sort_by")
	var sortBy []string
	if sortByStr != "" {
		sortBy = strings.Split(sortByStr, ",")
	}

	// Parse highlight
	highlight := c.Query("highlight") == "true"

	// Create search request
	searchReq := &services.SearchRequest{
		Query:     query,
		Page:      page,
		Limit:     limit,
		Indexes:   indexes,
		Filters:   filters,
		SortBy:    sortBy,
		Highlight: highlight,
	}

	// Perform search
	results, err := h.searchService.Search(searchReq)
	if err != nil {
		responses.SendError(c, 500, "Failed to perform search: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Search completed successfully", results)
}

// SearchPosts searches only in posts
// @Summary Search posts
// @Description Search posts with filters and sorting
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Results per page (default: 20)"
// @Param status query string false "Filter by post status"
// @Param public query bool false "Filter by public status"
// @Param category_id query string false "Filter by category ID"
// @Param created_by query string false "Filter by author ID"
// @Param sort_by query string false "Comma-separated list of sort fields"
// @Param highlight query bool false "Enable result highlighting"
// @Success 200 {object} responses.Response{data=services.SearchResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/posts [get]
func (h *SearchHandler) SearchPosts(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		responses.SendError(c, 400, "Query parameter is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	highlight := c.Query("highlight") == "true"

	// Build filters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if public := c.Query("public"); public != "" {
		filters["public"] = public == "true"
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		if _, err := uuid.Parse(categoryID); err == nil {
			filters["category_id"] = categoryID
		}
	}
	if createdBy := c.Query("created_by"); createdBy != "" {
		if _, err := uuid.Parse(createdBy); err == nil {
			filters["created_by"] = createdBy
		}
	}

	// Parse sort by
	sortByStr := c.Query("sort_by")
	var sortBy []string
	if sortByStr != "" {
		sortBy = strings.Split(sortByStr, ",")
	}

	searchReq := &services.SearchRequest{
		Query:     query,
		Page:      page,
		Limit:     limit,
		Indexes:   []string{"posts"},
		Filters:   filters,
		SortBy:    sortBy,
		Highlight: highlight,
	}

	results, err := h.searchService.Search(searchReq)
	if err != nil {
		responses.SendError(c, 500, "Failed to search posts: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Posts search completed successfully", results)
}

// SearchUsers searches only in users
// @Summary Search users
// @Description Search users with filters and sorting
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Results per page (default: 20)"
// @Param email_verified query bool false "Filter by email verification status"
// @Param phone_verified query bool false "Filter by phone verification status"
// @Param sort_by query string false "Comma-separated list of sort fields"
// @Param highlight query bool false "Enable result highlighting"
// @Success 200 {object} responses.Response{data=services.SearchResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/users [get]
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		responses.SendError(c, 400, "Query parameter is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	highlight := c.Query("highlight") == "true"

	// Build filters
	filters := make(map[string]interface{})
	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		filters["email_verified_at"] = emailVerified == "true"
	}
	if phoneVerified := c.Query("phone_verified"); phoneVerified != "" {
		filters["phone_verified_at"] = phoneVerified == "true"
	}

	// Parse sort by
	sortByStr := c.Query("sort_by")
	var sortBy []string
	if sortByStr != "" {
		sortBy = strings.Split(sortByStr, ",")
	}

	searchReq := &services.SearchRequest{
		Query:     query,
		Page:      page,
		Limit:     limit,
		Indexes:   []string{"users"},
		Filters:   filters,
		SortBy:    sortBy,
		Highlight: highlight,
	}

	results, err := h.searchService.Search(searchReq)
	if err != nil {
		responses.SendError(c, 500, "Failed to search users: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Users search completed successfully", results)
}

// SearchCategories searches only in categories
// @Summary Search categories
// @Description Search categories with filters and sorting
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Results per page (default: 20)"
// @Param is_active query bool false "Filter by active status"
// @Param parent_id query string false "Filter by parent category ID"
// @Param sort_by query string false "Comma-separated list of sort fields"
// @Param highlight query bool false "Enable result highlighting"
// @Success 200 {object} responses.Response{data=services.SearchResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/categories [get]
func (h *SearchHandler) SearchCategories(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		responses.SendError(c, 400, "Query parameter is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	highlight := c.Query("highlight") == "true"

	// Build filters
	filters := make(map[string]interface{})
	if isActive := c.Query("is_active"); isActive != "" {
		filters["is_active"] = isActive == "true"
	}
	if parentID := c.Query("parent_id"); parentID != "" {
		if _, err := uuid.Parse(parentID); err == nil {
			filters["parent_id"] = parentID
		}
	}

	// Parse sort by
	sortByStr := c.Query("sort_by")
	var sortBy []string
	if sortByStr != "" {
		sortBy = strings.Split(sortByStr, ",")
	}

	searchReq := &services.SearchRequest{
		Query:     query,
		Page:      page,
		Limit:     limit,
		Indexes:   []string{"categories"},
		Filters:   filters,
		SortBy:    sortBy,
		Highlight: highlight,
	}

	results, err := h.searchService.Search(searchReq)
	if err != nil {
		responses.SendError(c, 500, "Failed to search categories: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Categories search completed successfully", results)
}

// SearchMedia searches only in media
// @Summary Search media
// @Description Search media files with filters and sorting
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Results per page (default: 20)"
// @Param is_public query bool false "Filter by public status"
// @Param disk query string false "Filter by storage disk"
// @Param mime_type query string false "Filter by MIME type"
// @Param sort_by query string false "Comma-separated list of sort fields"
// @Param highlight query bool false "Enable result highlighting"
// @Success 200 {object} responses.Response{data=services.SearchResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/media [get]
func (h *SearchHandler) SearchMedia(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		responses.SendError(c, 400, "Query parameter is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	highlight := c.Query("highlight") == "true"

	// Build filters
	filters := make(map[string]interface{})
	if isPublic := c.Query("is_public"); isPublic != "" {
		filters["is_public"] = isPublic == "true"
	}
	if disk := c.Query("disk"); disk != "" {
		filters["disk"] = disk
	}
	if mimeType := c.Query("mime_type"); mimeType != "" {
		filters["mime_type"] = mimeType
	}

	// Parse sort by
	sortByStr := c.Query("sort_by")
	var sortBy []string
	if sortByStr != "" {
		sortBy = strings.Split(sortByStr, ",")
	}

	searchReq := &services.SearchRequest{
		Query:     query,
		Page:      page,
		Limit:     limit,
		Indexes:   []string{"media"},
		Filters:   filters,
		SortBy:    sortBy,
		Highlight: highlight,
	}

	results, err := h.searchService.Search(searchReq)
	if err != nil {
		responses.SendError(c, 500, "Failed to search media: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Media search completed successfully", results)
}

// GetSuggestions gets search suggestions for autocomplete
// @Summary Get search suggestions
// @Description Get search suggestions for autocomplete
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query (minimum 2 characters)"
// @Param index query string false "Index to search in (posts,users,categories,media)"
// @Success 200 {object} responses.Response{data=[]string}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/suggestions [get]
func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	query := c.Query("query")
	if len(query) < 2 {
		responses.SendError(c, 400, "Query must be at least 2 characters long")
		return
	}

	index := c.DefaultQuery("index", "posts")
	if !h.isValidIndex(index) {
		responses.SendError(c, 400, "Invalid index. Must be one of: posts, users, categories, media")
		return
	}

	suggestions, err := h.searchService.GetSuggestions(query, index)
	if err != nil {
		responses.SendError(c, 500, "Failed to get suggestions: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Suggestions retrieved successfully", suggestions)
}

// GetSearchStats gets statistics for search indexes
// @Summary Get search statistics
// @Description Get statistics for search indexes
// @Tags Search
// @Accept json
// @Produce json
// @Param index query string false "Index to get stats for (posts,users,categories,media)"
// @Success 200 {object} responses.Response{data=map[string]interface{}}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/stats [get]
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
	index := c.Query("index")
	if index != "" && !h.isValidIndex(index) {
		responses.SendError(c, 400, "Invalid index. Must be one of: posts, users, categories, media")
		return
	}

	if index == "" {
		// Get stats for all indexes
		allStats := make(map[string]interface{})
		indexes := []string{"posts", "users", "categories", "media"}

		for _, idx := range indexes {
			stats, err := h.searchService.GetIndexStats(idx)
			if err != nil {
				log.Printf("Failed to get stats for index %s: %v", idx, err)
				continue
			}
			allStats[idx] = stats
		}

		responses.SendSuccess(c, 200, "Search statistics retrieved successfully", allStats)
		return
	}

	// Get stats for specific index
	stats, err := h.searchService.GetIndexStats(index)
	if err != nil {
		responses.SendError(c, 500, "Failed to get search stats: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Search statistics retrieved successfully", stats)
}

// HealthCheck checks if Meilisearch is healthy
// @Summary Check search service health
// @Description Check if Meilisearch is healthy and responding
// @Tags Search
// @Accept json
// @Produce json
// @Success 200 {object} responses.Response{data=map[string]interface{}}
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/health [get]
func (h *SearchHandler) HealthCheck(c *gin.Context) {
	err := h.searchService.HealthCheck()
	if err != nil {
		responses.SendError(c, 500, "Search service is not healthy: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Search service is healthy", map[string]interface{}{
		"status":  "healthy",
		"service": "meilisearch",
	})
}

// ReindexAll reindexes all data from the database
// @Summary Reindex all data
// @Description Reindex all posts, users, categories, and media from the database
// @Tags Search
// @Accept json
// @Produce json
// @Success 200 {object} responses.Response{data=map[string]interface{}}
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/reindex [post]
func (h *SearchHandler) ReindexAll(c *gin.Context) {
	err := h.searchService.ReindexAll()
	if err != nil {
		responses.SendError(c, 500, "Failed to reindex data: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Reindexing completed successfully", map[string]interface{}{
		"message": "All data has been reindexed",
		"status":  "completed",
	})
}

// InitializeIndexes initializes all search indexes
// @Summary Initialize search indexes
// @Description Initialize and configure all search indexes
// @Tags Search
// @Accept json
// @Produce json
// @Success 200 {object} responses.Response{data=map[string]interface{}}
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/search/init [post]
func (h *SearchHandler) InitializeIndexes(c *gin.Context) {
	err := h.searchService.InitializeIndexes()
	if err != nil {
		responses.SendError(c, 500, "Failed to initialize indexes: "+err.Error())
		return
	}

	responses.SendSuccess(c, 200, "Search indexes initialized successfully", map[string]interface{}{
		"message": "All search indexes have been initialized and configured",
		"status":  "initialized",
	})
}

// isValidIndex checks if the provided index is valid
func (h *SearchHandler) isValidIndex(index string) bool {
	validIndexes := []string{"posts", "users", "categories", "media"}
	for _, validIndex := range validIndexes {
		if index == validIndex {
			return true
		}
	}
	return false
}
