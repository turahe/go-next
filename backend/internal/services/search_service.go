package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"

	"go-next/internal/models"
)

// SearchService handles all search operations using Meilisearch
type SearchService struct {
	client meilisearch.ServiceManager
	db     *gorm.DB
}

// SearchResult represents a search result
type SearchResult struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Content string                 `json:"content"`
	URL     string                 `json:"url,omitempty"`
	Data    map[string]interface{} `json:"data"`
	Score   float64                `json:"score"`
	Index   string                 `json:"index"`
	Created time.Time              `json:"created"`
	Updated time.Time              `json:"updated"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query       string                 `json:"query"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	SortBy      []string               `json:"sort_by,omitempty"`
	Page        int                    `json:"page"`
	Limit       int                    `json:"limit"`
	Indexes     []string               `json:"indexes,omitempty"`
	Highlight   bool                   `json:"highlight"`
	FacetQuery  string                 `json:"facet_query,omitempty"`
	FacetFilter string                 `json:"facet_filter,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results     []SearchResult         `json:"results"`
	Total       int64                  `json:"total"`
	Page        int                    `json:"page"`
	Limit       int                    `json:"limit"`
	TotalPages  int                    `json:"total_pages"`
	Query       string                 `json:"query"`
	Processing  bool                   `json:"processing"`
	Facets      map[string]interface{} `json:"facets,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// NewSearchService creates a new search service
func NewSearchService(client meilisearch.ServiceManager, db *gorm.DB) *SearchService {
	return &SearchService{
		client: client,
		db:     db,
	}
}

// InitializeIndexes creates and configures all search indexes
func (s *SearchService) InitializeIndexes() error {
	indexes := []string{"posts", "users", "categories", "media"}

	for _, indexName := range indexes {
		if err := s.createIndex(indexName); err != nil {
			return fmt.Errorf("failed to create index %s: %w", indexName, err)
		}
	}

	return nil
}

// createIndex creates and configures a search index
func (s *SearchService) createIndex(indexName string) error {
	// Create index if it doesn't exist
	_, err := s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexName,
		PrimaryKey: "id",
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	// Configure searchable attributes based on index type
	var searchableAttributes []string
	var filterableAttributes []string
	var sortableAttributes []string

	switch indexName {
	case "posts":
		searchableAttributes = []string{"title", "excerpt", "description", "slug"}
		filterableAttributes = []string{"status", "public", "category_id", "created_by", "created_at", "published_at"}
		sortableAttributes = []string{"created_at", "updated_at", "published_at", "view_count"}
	case "users":
		searchableAttributes = []string{"username", "email", "phone"}
		filterableAttributes = []string{"email_verified_at", "phone_verified_at", "created_at"}
		sortableAttributes = []string{"created_at", "updated_at"}
	case "categories":
		searchableAttributes = []string{"name", "slug", "description"}
		filterableAttributes = []string{"is_active", "parent_id", "created_at"}
		sortableAttributes = []string{"sort_order", "created_at", "updated_at"}
	case "media":
		searchableAttributes = []string{"file_name", "original_name", "mime_type"}
		filterableAttributes = []string{"is_public", "disk", "mime_type", "created_at"}
		sortableAttributes = []string{"created_at", "updated_at", "size"}
	}

	// Update searchable attributes
	if len(searchableAttributes) > 0 {
		_, err = s.client.Index(indexName).UpdateSearchableAttributes(&searchableAttributes)
		if err != nil {
			return fmt.Errorf("failed to update searchable attributes for %s: %w", indexName, err)
		}
	}

	// Update filterable attributes
	if len(filterableAttributes) > 0 {
		_, err = s.client.Index(indexName).UpdateFilterableAttributes(&filterableAttributes)
		if err != nil {
			return fmt.Errorf("failed to update filterable attributes for %s: %w", indexName, err)
		}
	}

	// Update sortable attributes
	if len(sortableAttributes) > 0 {
		_, err = s.client.Index(indexName).UpdateSortableAttributes(&sortableAttributes)
		if err != nil {
			return fmt.Errorf("failed to update sortable attributes for %s: %w", indexName, err)
		}
	}

	log.Printf("Index %s configured successfully", indexName)
	return nil
}

// IndexPost indexes a post for search
func (s *SearchService) IndexPost(post *models.Post) error {
	// Get related data
	var user models.User
	var category models.Category

	if err := s.db.First(&user, post.CreatedBy).Error; err != nil {
		log.Printf("Failed to get user for post indexing: %v", err)
	}

	if post.CategoryID != nil {
		if err := s.db.First(&category, post.CategoryID).Error; err != nil {
			log.Printf("Failed to get category for post indexing: %v", err)
		}
	}

	// Prepare search document
	doc := map[string]interface{}{
		"id":           post.ID.String(),
		"type":         "post",
		"title":        post.Title,
		"slug":         post.Slug,
		"excerpt":      post.Excerpt,
		"description":  post.Description,
		"status":       string(post.Status),
		"public":       post.Public,
		"view_count":   post.ViewCount,
		"created_by":   post.CreatedBy.String(),
		"category_id":  post.CategoryID,
		"created_at":   post.CreatedAt.Unix(),
		"updated_at":   post.UpdatedAt.Unix(),
		"published_at": nil,
		"user": map[string]interface{}{
			"id":       user.ID.String(),
			"username": user.Username,
			"email":    user.Email,
		},
		"category": nil,
	}

	if post.PublishedAt != nil {
		doc["published_at"] = post.PublishedAt.Unix()
	}

	if post.CategoryID != nil {
		doc["category"] = map[string]interface{}{
			"id":   category.ID.String(),
			"name": category.Name,
			"slug": category.Slug,
		}
	}

	// Add to search index
	_, err := s.client.Index("posts").AddDocuments([]map[string]interface{}{doc})
	return err
}

// IndexUser indexes a user for search
func (s *SearchService) IndexUser(user *models.User) error {
	doc := map[string]interface{}{
		"id":                user.ID.String(),
		"type":              "user",
		"username":          user.Username,
		"email":             user.Email,
		"phone":             user.Phone,
		"email_verified_at": nil,
		"phone_verified_at": nil,
		"created_at":        user.CreatedAt.Unix(),
		"updated_at":        user.UpdatedAt.Unix(),
	}

	if user.EmailVerified != nil {
		doc["email_verified_at"] = user.EmailVerified.Unix()
	}

	if user.PhoneVerified != nil {
		doc["phone_verified_at"] = user.PhoneVerified.Unix()
	}

	_, err := s.client.Index("users").AddDocuments([]map[string]interface{}{doc})
	return err
}

// IndexCategory indexes a category for search
func (s *SearchService) IndexCategory(category *models.Category) error {
	doc := map[string]interface{}{
		"id":          category.ID.String(),
		"type":        "category",
		"name":        category.Name,
		"slug":        category.Slug,
		"description": category.Description,
		"is_active":   category.IsActive,
		"parent_id":   category.ParentID,
		"sort_order":  category.SortOrder,
		"created_at":  category.CreatedAt.Unix(),
		"updated_at":  category.UpdatedAt.Unix(),
	}

	_, err := s.client.Index("categories").AddDocuments([]map[string]interface{}{doc})
	return err
}

// IndexMedia indexes a media file for search
func (s *SearchService) IndexMedia(media *models.Media) error {
	doc := map[string]interface{}{
		"id":            media.ID.String(),
		"type":          "media",
		"uuid":          media.UUID,
		"file_name":     media.FileName,
		"original_name": media.OriginalName,
		"mime_type":     media.MimeType,
		"size":          media.Size,
		"disk":          media.Disk,
		"path":          media.Path,
		"url":           media.URL,
		"width":         media.Width,
		"height":        media.Height,
		"duration":      media.Duration,
		"is_public":     media.IsPublic,
		"created_at":    media.CreatedAt.Unix(),
		"updated_at":    media.UpdatedAt.Unix(),
	}

	_, err := s.client.Index("media").AddDocuments([]map[string]interface{}{doc})
	return err
}

// DeleteFromIndex removes a document from the search index
func (s *SearchService) DeleteFromIndex(indexName, documentID string) error {
	_, err := s.client.Index(indexName).DeleteDocument(documentID)
	return err
}

// Search performs a search across all indexes or specific indexes
func (s *SearchService) Search(req *SearchRequest) (*SearchResponse, error) {
	var allResults []SearchResult
	var total int64

	// Determine which indexes to search
	indexes := req.Indexes
	if len(indexes) == 0 {
		indexes = []string{"posts", "users", "categories", "media"}
	}

	// Search each index
	for _, indexName := range indexes {
		results, err := s.searchIndex(indexName, req)
		if err != nil {
			log.Printf("Failed to search index %s: %v", indexName, err)
			continue
		}

		allResults = append(allResults, results...)
	}

	// Calculate total
	total = int64(len(allResults))

	// Apply pagination
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if start >= len(allResults) {
		start = len(allResults)
	}
	if end > len(allResults) {
		end = len(allResults)
	}

	paginatedResults := allResults[start:end]

	// Calculate total pages
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &SearchResponse{
		Results:    paginatedResults,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
		Query:      req.Query,
		Processing: false,
	}, nil
}

// searchIndex searches a specific index
func (s *SearchService) searchIndex(indexName string, req *SearchRequest) ([]SearchResult, error) {
	// Build search request
	searchRequest := &meilisearch.SearchRequest{
		Query:               req.Query,
		Limit:               int64(req.Limit),
		Offset:              int64((req.Page - 1) * req.Limit),
		ShowMatchesPosition: req.Highlight,
	}

	// Add filters if provided
	if len(req.Filters) > 0 {
		filterStr := s.buildFilterString(req.Filters)
		searchRequest.Filter = &filterStr
	}

	// Add sorting if provided
	if len(req.SortBy) > 0 {
		searchRequest.Sort = req.SortBy
	}

	// Perform search
	resp, err := s.client.Index(indexName).Search(req.Query, searchRequest)
	if err != nil {
		return nil, err
	}

	// Convert results
	var results []SearchResult
	for _, hit := range resp.Hits {
		result := s.convertHitToResult(hit, indexName)
		results = append(results, result)
	}

	return results, nil
}

// buildFilterString converts filters map to Meilisearch filter string
func (s *SearchService) buildFilterString(filters map[string]interface{}) string {
	var conditions []string

	for key, value := range filters {
		switch v := value.(type) {
		case string:
			conditions = append(conditions, fmt.Sprintf("%s = %s", key, v))
		case int, int64:
			conditions = append(conditions, fmt.Sprintf("%s = %v", key, v))
		case bool:
			conditions = append(conditions, fmt.Sprintf("%s = %t", key, v))
		case []interface{}:
			// Handle array values (IN operator)
			var values []string
			for _, val := range v {
				values = append(values, fmt.Sprintf("%v", val))
			}
			conditions = append(conditions, fmt.Sprintf("%s IN [%s]", key, strings.Join(values, ", ")))
		}
	}

	return strings.Join(conditions, " AND ")
}

// convertHitToResult converts a Meilisearch hit to SearchResult
func (s *SearchService) convertHitToResult(hit interface{}, indexName string) SearchResult {
	hitMap, ok := hit.(map[string]interface{})
	if !ok {
		return SearchResult{}
	}

	result := SearchResult{
		Index: indexName,
	}

	// Extract common fields
	if id, ok := hitMap["id"].(string); ok {
		result.ID = id
	}

	if score, ok := hitMap["_score"].(float64); ok {
		result.Score = score
	}

	// Extract type-specific fields
	switch indexName {
	case "posts":
		result.Type = "post"
		if title, ok := hitMap["title"].(string); ok {
			result.Title = title
		}
		if excerpt, ok := hitMap["excerpt"].(string); ok {
			result.Content = excerpt
		}
		if slug, ok := hitMap["slug"].(string); ok {
			result.URL = fmt.Sprintf("/posts/%s", slug)
		}
	case "users":
		result.Type = "user"
		if username, ok := hitMap["username"].(string); ok {
			result.Title = username
		}
		if email, ok := hitMap["email"].(string); ok {
			result.Content = email
		}
		result.URL = fmt.Sprintf("/users/%s", result.ID)
	case "categories":
		result.Type = "category"
		if name, ok := hitMap["name"].(string); ok {
			result.Title = name
		}
		if description, ok := hitMap["description"].(string); ok {
			result.Content = description
		}
		if slug, ok := hitMap["slug"].(string); ok {
			result.URL = fmt.Sprintf("/categories/%s", slug)
		}
	case "media":
		result.Type = "media"
		if fileName, ok := hitMap["file_name"].(string); ok {
			result.Title = fileName
		}
		if originalName, ok := hitMap["original_name"].(string); ok {
			result.Content = originalName
		}
		if url, ok := hitMap["url"].(string); ok {
			result.URL = url
		}
	}

	// Extract timestamps
	if created, ok := hitMap["created_at"].(float64); ok {
		result.Created = time.Unix(int64(created), 0)
	}
	if updated, ok := hitMap["updated_at"].(float64); ok {
		result.Updated = time.Unix(int64(updated), 0)
	}

	// Store all data
	result.Data = hitMap

	return result
}

// GetSuggestions gets search suggestions for autocomplete
func (s *SearchService) GetSuggestions(query string, indexName string) ([]string, error) {
	if len(query) < 2 {
		return []string{}, nil
	}

	// Use Meilisearch's search with limit for suggestions
	searchRequest := &meilisearch.SearchRequest{
		Query: query,
		Limit: 10,
	}

	resp, err := s.client.Index(indexName).Search(query, searchRequest)
	if err != nil {
		return nil, err
	}

	var suggestions []string
	for _, hit := range resp.Hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract relevant fields for suggestions
		switch indexName {
		case "posts":
			if title, ok := hitMap["title"].(string); ok {
				suggestions = append(suggestions, title)
			}
		case "users":
			if username, ok := hitMap["username"].(string); ok {
				suggestions = append(suggestions, username)
			}
		case "categories":
			if name, ok := hitMap["name"].(string); ok {
				suggestions = append(suggestions, name)
			}
		case "media":
			if fileName, ok := hitMap["file_name"].(string); ok {
				suggestions = append(suggestions, fileName)
			}
		}
	}

	return suggestions, nil
}

// ReindexAll reindexes all data from the database
func (s *SearchService) ReindexAll() error {
	// Reindex posts
	var posts []models.Post
	if err := s.db.Preload("User").Preload("Category").Find(&posts).Error; err != nil {
		return fmt.Errorf("failed to load posts: %w", err)
	}

	for _, post := range posts {
		if err := s.IndexPost(&post); err != nil {
			log.Printf("Failed to index post %s: %v", post.ID, err)
		}
	}

	// Reindex users
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	for _, user := range users {
		if err := s.IndexUser(&user); err != nil {
			log.Printf("Failed to index user %s: %v", user.ID, err)
		}
	}

	// Reindex categories
	var categories []models.Category
	if err := s.db.Find(&categories).Error; err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	for _, category := range categories {
		if err := s.IndexCategory(&category); err != nil {
			log.Printf("Failed to index category %s: %v", category.ID, err)
		}
	}

	// Reindex media
	var media []models.Media
	if err := s.db.Find(&media).Error; err != nil {
		return fmt.Errorf("failed to load media: %w", err)
	}

	for _, m := range media {
		if err := s.IndexMedia(&m); err != nil {
			log.Printf("Failed to index media %s: %v", m.ID, err)
		}
	}

	log.Println("Reindexing completed successfully")
	return nil
}

// GetIndexStats gets statistics for a search index
func (s *SearchService) GetIndexStats(indexName string) (map[string]interface{}, error) {
	stats, err := s.client.Index(indexName).GetStats()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"number_of_documents": stats.NumberOfDocuments,
		"is_indexing":         stats.IsIndexing,
		"field_distribution":  stats.FieldDistribution,
	}, nil
}

// HealthCheck checks if Meilisearch is healthy
func (s *SearchService) HealthCheck() error {
	_, err := s.client.Health()
	return err
}
