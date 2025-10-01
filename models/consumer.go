package models

import (
	"time"

	"gorm.io/gorm"
)

// Consumer represents a process that has accessed a file.
// Uniqueness is defined by the tuple (FilePath, Comm, Exe, RUID, EUID).
type Consumer struct {
	gorm.Model
	FilePath string `json:"filePath" gorm:"index;uniqueIndex:uniq_consumer,priority:1"`
	Comm     string `json:"comm" gorm:"uniqueIndex:uniq_consumer,priority:2"`
	Exe      string `json:"exe" gorm:"uniqueIndex:uniq_consumer,priority:3"`
	RUID     int    `json:"ruid" gorm:"uniqueIndex:uniq_consumer,priority:4"`
	EUID     int    `json:"euid" gorm:"uniqueIndex:uniq_consumer,priority:5"`
}

// FirstSeen returns CreatedAt for convenience.
func (c *Consumer) FirstSeen() time.Time { return c.CreatedAt }

// LastSeen returns UpdatedAt for convenience.
func (c *Consumer) LastSeen() time.Time { return c.UpdatedAt }

// ConsumerFilter filters consumers when listing
type ConsumerFilter struct {
	FilePath string
}
