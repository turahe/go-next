package services

import (
	"context"
	"fmt"
	"time"
)

// ExampleTagUsage demonstrates how to use the TagService
func ExampleTagUsage() {
	ctx := context.Background()

	// Initialize logger
	logger := GetLogger("TagExample")

	// Example 1: Create a new tag
	logger.Info(ctx, "ExampleTagUsage", "Creating new tag", map[string]interface{}{
		"tag_name": "Golang",
		"tag_type": "general",
	})

	tag := &Tag{
		Name:        "Golang",
		Slug:        "golang",
		Description: "Go programming language content",
		Color:       "#00ADD8",
		Type:        "general",
		IsActive:    true,
	}

	start := time.Now()
	err := TagSvc.CreateTag(ctx, tag)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "ExampleTagUsage", "Failed to create tag", err, map[string]interface{}{
			"tag_name": tag.Name,
		})
		return
	}

	logger.Performance(ctx, "CreateTag", duration, false, 1, map[string]interface{}{
		"tag_id":   tag.ID,
		"tag_name": tag.Name,
	})

	// Example 2: Add tag to a post
	logger.Info(ctx, "ExampleTagUsage", "Adding tag to post", map[string]interface{}{
		"tag_id":  tag.ID,
		"post_id": 123,
	})

	err = TagSvc.AddTagToEntity(ctx, tag.ID, 123, "post")
	if err != nil {
		logger.Error(ctx, "ExampleTagUsage", "Failed to add tag to post", err, map[string]interface{}{
			"tag_id":  tag.ID,
			"post_id": 123,
		})
		return
	}

	// Example 3: Get tags for a post
	tags, err := TagSvc.GetTagsByEntity(ctx, 123, "post")
	if err != nil {
		logger.Error(ctx, "ExampleTagUsage", "Failed to get post tags", err, map[string]interface{}{
			"post_id": 123,
		})
		return
	}

	logger.Info(ctx, "ExampleTagUsage", "Retrieved post tags", map[string]interface{}{
		"post_id":   123,
		"tag_count": len(tags),
		"tags":      tags,
	})

	// Example 4: Search tags
	searchResults, total, err := TagSvc.SearchTags(ctx, "golang", 10, 0)
	if err != nil {
		logger.Error(ctx, "ExampleTagUsage", "Failed to search tags", err, map[string]interface{}{
			"query": "golang",
		})
		return
	}

	logger.Info(ctx, "ExampleTagUsage", "Tag search completed", map[string]interface{}{
		"query":          "golang",
		"results":        len(searchResults),
		"total":          total,
		"search_results": searchResults,
	})
}

// ExampleLogging demonstrates different logging capabilities
func ExampleLogging() {
	ctx := context.Background()

	// Add context values for tracing
	ctx = context.WithValue(ctx, "user_id", uint(123))
	ctx = context.WithValue(ctx, "request_id", "req-456")
	ctx = context.WithValue(ctx, "trace_id", "trace-789")

	logger := GetLogger("LoggingExample")

	// Basic logging
	logger.Debug(ctx, "ExampleLogging", "Debug message", map[string]interface{}{
		"debug_info": "This is debug information",
	})

	logger.Info(ctx, "ExampleLogging", "Info message", map[string]interface{}{
		"info_data": "This is info data",
	})

	logger.Warning(ctx, "ExampleLogging", "Warning message", map[string]interface{}{
		"warning_reason": "Something to be aware of",
	})

	// Performance logging
	start := time.Now()
	time.Sleep(100 * time.Millisecond) // Simulate work
	duration := time.Since(start)

	logger.Performance(ctx, "SimulatedOperation", duration, true, 2, map[string]interface{}{
		"operation_type": "database_query",
		"cache_hit":      true,
		"database_ops":   2,
	})

	// Cache operation logging
	logger.Cache(ctx, "GetUserByID", "get", "user:123", true, map[string]interface{}{
		"user_id": 123,
		"ttl":     "30m",
	})

	// Database operation logging
	logger.Database(ctx, "CreateUser", "insert", "users", 50*time.Millisecond, 1, nil, map[string]interface{}{
		"user_id": 123,
		"email":   "user@example.com",
	})

	// Security event logging
	logger.Security(ctx, "Login", "user_login", 123, true, map[string]interface{}{
		"ip_address": "192.168.1.1",
		"user_agent": "Mozilla/5.0...",
		"method":     "password",
	})

	// Audit trail logging
	logger.Audit(ctx, "UpdatePost", "post_updated", 123, 456, "post", map[string]interface{}{
		"changes": map[string]interface{}{
			"title":   "Old Title → New Title",
			"content": "Content modified",
		},
		"timestamp": time.Now(),
	})
}

// ExampleTagManagement demonstrates tag management operations
func ExampleTagManagement() {
	ctx := context.Background()
	logger := GetLogger("TagManagement")

	// Create different types of tags
	tags := []*Tag{
		{
			Name:        "Featured",
			Slug:        "featured",
			Description: "Featured content",
			Color:       "#FFD700",
			Type:        "feature",
			IsActive:    true,
		},
		{
			Name:        "Technology",
			Slug:        "technology",
			Description: "Technology related content",
			Color:       "#007BFF",
			Type:        "category",
			IsActive:    true,
		},
		{
			Name:        "Beta Feature",
			Slug:        "beta-feature",
			Description: "Beta feature flag",
			Color:       "#FF6B6B",
			Type:        "system",
			IsActive:    true,
		},
	}

	// Create tags
	for _, tag := range tags {
		logger.Info(ctx, "ExampleTagManagement", "Creating tag", map[string]interface{}{
			"tag_name": tag.Name,
			"tag_type": tag.Type,
		})

		err := TagSvc.CreateTag(ctx, tag)
		if err != nil {
			logger.Error(ctx, "ExampleTagManagement", "Failed to create tag", err, map[string]interface{}{
				"tag_name": tag.Name,
			})
			continue
		}

		logger.Info(ctx, "ExampleTagManagement", "Tag created successfully", map[string]interface{}{
			"tag_id":   tag.ID,
			"tag_name": tag.Name,
		})
	}

	// Get tags by type
	featureTags, err := TagSvc.GetAllTags(ctx, "feature")
	if err != nil {
		logger.Error(ctx, "ExampleTagManagement", "Failed to get feature tags", err)
	} else {
		logger.Info(ctx, "ExampleTagManagement", "Retrieved feature tags", map[string]interface{}{
			"count": len(featureTags),
			"tags":  featureTags,
		})
	}

	// Get active tags
	activeTags, err := TagSvc.GetActiveTags(ctx)
	if err != nil {
		logger.Error(ctx, "ExampleTagManagement", "Failed to get active tags", err)
	} else {
		logger.Info(ctx, "ExampleTagManagement", "Retrieved active tags", map[string]interface{}{
			"count": len(activeTags),
			"tags":  activeTags,
		})
	}

	// Get tag count
	count, err := TagSvc.GetTagCount(ctx)
	if err != nil {
		logger.Error(ctx, "ExampleTagManagement", "Failed to get tag count", err)
	} else {
		logger.Info(ctx, "ExampleTagManagement", "Tag count retrieved", map[string]interface{}{
			"total_count": count,
		})
	}
}

// ExampleEntityTagging demonstrates tagging different entity types
func ExampleEntityTagging() {
	ctx := context.Background()
	logger := GetLogger("EntityTagging")

	// Create a tag for demonstration
	tag := &Tag{
		Name:        "Popular",
		Slug:        "popular",
		Description: "Popular content",
		Color:       "#28A745",
		Type:        "general",
		IsActive:    true,
	}

	err := TagSvc.CreateTag(ctx, tag)
	if err != nil {
		logger.Error(ctx, "ExampleEntityTagging", "Failed to create tag", err)
		return
	}

	// Tag different entity types
	entities := []struct {
		ID   uint
		Type string
	}{
		{123, "post"},
		{456, "user"},
		{789, "media"},
		{101, "category"},
	}

	for _, entity := range entities {
		logger.Info(ctx, "ExampleEntityTagging", "Adding tag to entity", map[string]interface{}{
			"tag_id":      tag.ID,
			"entity_id":   entity.ID,
			"entity_type": entity.Type,
		})

		err := TagSvc.AddTagToEntity(ctx, tag.ID, entity.ID, entity.Type)
		if err != nil {
			logger.Error(ctx, "ExampleEntityTagging", "Failed to add tag to entity", err, map[string]interface{}{
				"tag_id":      tag.ID,
				"entity_id":   entity.ID,
				"entity_type": entity.Type,
			})
			continue
		}

		// Get tags for this entity
		entityTags, err := TagSvc.GetTagsByEntity(ctx, entity.ID, entity.Type)
		if err != nil {
			logger.Error(ctx, "ExampleEntityTagging", "Failed to get entity tags", err, map[string]interface{}{
				"entity_id":   entity.ID,
				"entity_type": entity.Type,
			})
			continue
		}

		logger.Info(ctx, "ExampleEntityTagging", "Entity tags retrieved", map[string]interface{}{
			"entity_id":   entity.ID,
			"entity_type": entity.Type,
			"tag_count":   len(entityTags),
			"tags":        entityTags,
		})
	}

	// Get all entities with this tag
	for _, entityType := range []string{"post", "user", "media", "category"} {
		entities, total, err := TagSvc.GetEntitiesByTag(ctx, tag.ID, entityType, 10, 0)
		if err != nil {
			logger.Error(ctx, "ExampleEntityTagging", "Failed to get entities by tag", err, map[string]interface{}{
				"tag_id":      tag.ID,
				"entity_type": entityType,
			})
			continue
		}

		logger.Info(ctx, "ExampleEntityTagging", "Entities by tag retrieved", map[string]interface{}{
			"tag_id":      tag.ID,
			"entity_type": entityType,
			"count":       len(entities),
			"total":       total,
			"entities":    entities,
		})
	}
}

// RunAllExamples runs all the example functions
func RunAllExamples() {
	fmt.Println("Running TagService and Logging Examples...")

	ExampleTagUsage()
	ExampleLogging()
	ExampleTagManagement()
	ExampleEntityTagging()

	fmt.Println("All examples completed!")
}
