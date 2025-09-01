package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ConsumerEntry struct {
	gorm.Model
	EntryID        string `json:"entryId" gorm:"unique;index"`
	Hostname       string `json:"hostname"`
	Executable     string `json:"executable"`
	User           string `json:"user"`
	PID            string `json:"pid"`
	StartTimestamp string `json:"startTimestamp"`
	ConsumerJobID  uint   `json:"consumerJobId"` // Foreign key for ConsumerJob
}

type ConsumerJob struct {
	gorm.Model
	JobID      string             `json:"jobId" gorm:"unique;index"`
	Status     string             `json:"status"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
	FinishedAt time.Time          `json:"finishedAt"`
	Match      ConsumerSliceMatch `json:"match" gorm:"type:text"` // One-to-many relationship
}

type ConsumerSliceMatch []ConsumerMatchResumed
type ConsumerMatchResumed struct {
	Hostname       string `json:"hostname"`
	Executable     string `json:"executable"`
	User           string `json:"user"`
	PID            string `json:"pid"`
	StartTimestamp string `json:"startTimestamp"`
}

// Value implements the driver.Valuer interface, converting the slice to a JSON string.
func (s ConsumerSliceMatch) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface, converting the JSON string from the database to a slice.
func (s *ConsumerSliceMatch) Scan(value interface{}) error {
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
