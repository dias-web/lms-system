package entity

import "time"

type Course struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Chapters []Chapter `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"chapters,omitempty"`
}

func (Course) TableName() string { return "courses" }
