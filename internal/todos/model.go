package todos

import (
	"time"
)

type ToDoItem struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
	Text      string     `gorm:"not null" validate:"required"`
	Done      bool       `gorm:"default:false"`
	UserID    uint
}

type ToDoItemUpdateInput struct {
	Text *string `json:"text"`
	Done *bool   `json:"done"`
}

type PaginationDetails struct {
	Page  int
	Limit int
	Order string
}

type PaginationMetadata struct {
	ResultCount int
	TotalCount  int
}

type PaginatedResponse struct {
	Data []ToDoItem         `json:"data"`
	Meta PaginationMetadata `json:"metadata, omitempty"`
}
