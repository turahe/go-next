package models

type Mediable struct {
	MediaID      uint   `gorm:"column:media_id;not null"`
	MediableID   uint   `gorm:"column:mediable_id;not null"`
	MediableType string `gorm:"column:mediable_type;not null"`
	Group        string `gorm:"column:group;not null"`
}
