package models

type Setting struct {
	BaseModelWithUser
	EntityType string `json:"entity_type" gorm:"null;size:255"`
	EntityID   string `json:"entity_id" gorm:"null;size:255"`
	Key        string `json:"key" gorm:"primaryKey;not null;size:255"`
	Value      string `json:"value" gorm:"type:text;null"`
}
