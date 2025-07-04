package services

import (
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/pkg/database"
)

type BaseService struct{}

func (s *BaseService) Create(value interface{}) error {
	return database.DB.Create(value).Error
}

func (s *BaseService) Save(value interface{}) error {
	return database.DB.Save(value).Error
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

// Paginate is a generic pagination method for all service models
func (s *BaseService) Paginate(model interface{}, params PaginationParams, out interface{}) (*responses.PaginationResponse, error) {
	db := database.DB.Model(model)

	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}

	offset := (params.Page - 1) * params.PerPage

	if err := db.Limit(params.PerPage).Offset(offset).Find(out).Error; err != nil {
		return nil, err
	}

	totalPage := (totalCount + int64(params.PerPage) - 1) / int64(params.PerPage)
	lastPage := totalPage
	var nextPage, prevPage int64
	if int64(params.Page) < totalPage {
		nextPage = int64(params.Page) + 1
	}
	if params.Page > 1 {
		prevPage = int64(params.Page) - 1
	}

	return &responses.PaginationResponse{
		Data:         out,
		TotalCount:   totalCount,
		TotalPage:    totalPage,
		CurrentPage:  int64(params.Page),
		LastPage:     lastPage,
		PerPage:      int64(params.PerPage),
		NextPage:     nextPage,
		PreviousPage: prevPage,
	}, nil
}

func (s *BaseService) Update(value interface{}) error {
	return database.DB.Save(value).Error
}

func (s *BaseService) Delete(value interface{}) error {
	return database.DB.Delete(value).Error
}

func (s *BaseService) FirstOrFail(out interface{}, query interface{}, args ...interface{}) error {
	db := database.DB
	if err := db.Where(query, args...).First(out).Error; err != nil {
		return err
	}
	return nil
}
