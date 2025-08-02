package dto

import (
	"go-next/internal/models"
	"time"

	"github.com/google/uuid"
)

type MediaDTO struct {
	ID           uuid.UUID `json:"id"`
	UUID         string    `json:"uuid"`
	FileName     string    `json:"fileName"`
	OriginalName string    `json:"originalName"`
	Disk         string    `json:"disk"`
	MimeType     string    `json:"mimeType"`
	Size         int64     `json:"size"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	Duration     *float64  `json:"duration,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func ToMediaDTO(m *models.Media) *MediaDTO {
	return &MediaDTO{
		ID:           m.ID,
		UUID:         m.UUID,
		FileName:     m.FileName,
		OriginalName: m.OriginalName,
		Disk:         m.Disk,
		MimeType:     m.MimeType,
		Size:         m.Size,
		Width:        m.Width,
		Height:       m.Height,
		Duration:     m.Duration,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
