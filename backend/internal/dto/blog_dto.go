package dto

// BlogStats represents comprehensive blog statistics
type BlogStats struct {
	TotalPosts      int64 `json:"total_posts"`
	PublishedPosts  int64 `json:"published_posts"`
	TotalViews      int64 `json:"total_views"`
	TotalComments   int64 `json:"total_comments"`
	TotalCategories int64 `json:"total_categories"`
	TotalTags       int64 `json:"total_tags"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	PostCount    int64  `json:"post_count"`
	ViewCount    int64  `json:"view_count"`
}

// MonthlyArchive represents monthly post counts
type MonthlyArchive struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}
