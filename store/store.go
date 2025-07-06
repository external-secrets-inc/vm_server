package store

import (
	"log"
	"vm-server/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Store manages the database connection and operations.
type Store struct {
	DB *gorm.DB
}

// NewStore initializes the database connection and migrates the schema.
func NewStore() (*Store, error) {
	// Using an in-memory SQLite database for simplicity.
	// For persistence, you can replace "file::memory:?cache=shared" with a file path like "gorm.db".
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established.")

	// Auto-migrate the schema for ScanJob and ScanEntry models.
	err = db.AutoMigrate(&models.ScanJob{}, &models.ScanEntry{})
	if err != nil {
		// Attempt to close the database connection if migration fails.
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, err
	}

	log.Println("Database schema migrated.")

	return &Store{DB: db}, nil
}

// CreateScanJob creates a new scan job in the database.
func (s *Store) CreateScanJob(scanJob *models.ScanJob) error {
	result := s.DB.Create(scanJob)
	return result.Error
}

// GetScanJob retrieves a scan job by its ID, preloading associated matches.
func (s *Store) GetScanJob(id string) (*models.ScanJob, error) {
	var scanJob models.ScanJob
	result := s.DB.Preload("Match").First(&scanJob, "job_id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &scanJob, nil
}

// CreateScanEntry creates a new scan entry in the database.
func (s *Store) CreateScanEntry(scanEntry *models.ScanEntry) error {
	result := s.DB.Create(scanEntry)
	return result.Error
}

// GetScanEntry retrieves a scan entry by its EntryID.
func (s *Store) GetScanEntry(entryID string) (*models.ScanEntry, error) {
	var scanEntry models.ScanEntry
	result := s.DB.First(&scanEntry, "entry_id = ?", entryID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &scanEntry, nil
}

// UpdateScanJob updates an existing scan job in the database.
func (s *Store) UpdateScanJob(scanJob *models.ScanJob) error {
	result := s.DB.Save(scanJob)
	return result.Error
}

// UpdateScanEntry updates an existing scan entry in the database.
func (s *Store) UpdateScanEntry(scanEntry *models.ScanEntry) error {
	result := s.DB.Save(scanEntry)
	return result.Error
}

func (s *Store) DeleteScanEntry(scanEntry *models.ScanEntry) error {
	result := s.DB.Delete(scanEntry)
	return result.Error
}

func (s *Store) ListScanEntries() ([]models.ScanEntry, error) {
	var scanEntries []models.ScanEntry
	result := s.DB.Find(&scanEntries)
	return scanEntries, result.Error
}
