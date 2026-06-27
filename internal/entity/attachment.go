package entity

import "time"

// Attachment is a file stored in MinIO and linked to a lesson. The bytes live
// in object storage under ObjectKey; this row keeps the metadata.
type Attachment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	LessonID    uint      `gorm:"not null;index" json:"lesson_id"`
	FileName    string    `gorm:"type:varchar(255);not null" json:"file_name"`
	ObjectKey   string    `gorm:"type:varchar(512);not null;uniqueIndex" json:"object_key"`
	ContentType string    `gorm:"type:varchar(255)" json:"content_type"`
	Size        int64     `gorm:"not null" json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Lesson *Lesson `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`
}

func (Attachment) TableName() string { return "attachments" }