package entity

import "time"

type Lesson struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Content     string    `gorm:"type:text" json:"content"`
	Order       int       `gorm:"not null;default:0" json:"order"`
	ChapterID   uint      `gorm:"not null;index" json:"chapter_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Chapter *Chapter `gorm:"foreignKey:ChapterID" json:"chapter,omitempty"`
}

func (Lesson) TableName() string { return "lessons" }