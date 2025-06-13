package todos

import (
	"gorm.io/gorm"
)

type ToDoItem struct {
	gorm.Model
	Text string `gorm:"not null" validate:"required"`
	Done bool   `gorm:"default:false"`
}

type ToDoItemUpdateInput struct {
	Text *string `json:"text"`
	Done *bool   `json:"done"`
}
