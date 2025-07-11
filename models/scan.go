package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringSlice is a custom type to handle storing string slices as JSON in the database.
type StringSlice []string

// Value implements the driver.Valuer interface, converting the slice to a JSON string.
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface, converting the JSON string from the database to a slice.
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type ScanEntry struct {
	gorm.Model
	Fingerprint    string      `json:"fingerprint" gorm:"unique;index"`
	EntryID        string      `json:"entryId" gorm:"unique;index"`
	SuperseededBy  *ScanEntry  `json:"superseededBy" gorm:"foreignKey:EntryID"`
	FilePath       string      `json:"filePath"`
	MatchedRegexes StringSlice `json:"matchedRegexes" gorm:"type:text"`
	StartLine      int         `json:"startLine"`
	EndLine        int         `json:"endLine"`
	StartColumn    int         `json:"startColumn"`
	EndColumn      int         `json:"endColumn"`
	ScanJobID      uint        `json:"scanJobId"` // Foreign key for ScanJob
}

type ScanJob struct {
	gorm.Model
	JobID      string     `json:"jobId" gorm:"unique;index"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	FinishedAt time.Time  `json:"finishedAt"`
	Match      SliceMatch `json:"match" gorm:"type:text"` // One-to-many relationship
}

type SliceMatch []ScanMatchResumed
type ScanMatchResumed struct {
	Key      string `json:"key"`
	Property string `json:"property"`
}

// Value implements the driver.Valuer interface, converting the slice to a JSON string.
func (s SliceMatch) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface, converting the JSON string from the database to a slice.
func (s *SliceMatch) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}
