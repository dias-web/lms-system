package entity

import "time"

type Chapter struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Order       int       `gorm:"not null;default:0" json:"order"`
	CourseID    uint      `gorm:"not null;index" json:"course_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Course  *Course  `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Lessons []Lesson `gorm:"foreignKey:ChapterID;constraint:OnDelete:CASCADE" json:"lessons,omitempty"`
}

func (Chapter) TableName() string { return "chapters" }