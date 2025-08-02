package dto

import (
	"time"
	"wordpress-go-next/backend/internal/models"

	"github.com/google/uuid"
)

type MediaDTO struct {
	ID        uint64    `json:"id"`
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	FileName  string    `json:"fileName"`
	Disk      string    `json:"disk"`
	MimeType  string    `json:"mimeType"`
	Size      int64     `json:"size"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	Duration  *int      `json:"duration,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToMediaDTO(m *models.Media) *MediaDTO {
	return &MediaDTO{
		ID:        m.ID,
		UUID:      m.UUID,
		Name:      m.Name,
		FileName:  m.FileName,
		Disk:      m.Disk,
		MimeType:  m.MimeType,
		Size:      m.Size,
		Width:     m.Width,
		Height:    m.Height,
		Duration:  m.Duration,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
