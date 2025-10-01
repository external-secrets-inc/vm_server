package store

import (
	"errors"
	"log"
	"time"
	"vm-server/models"
	"vm-server/store/schema"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Store manages the database connection and operations.
type StoreGorm struct {
	DB *gorm.DB
}

// NewStore initializes the database connection and migrates the schema.
func NewStore() (schema.Store, error) {
	// Using an in-memory SQLite database for simplicity.
	// For persistence, you can replace "file::memory:?cache=shared" with a file path like "gorm.db".
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established.")

	// Auto-migrate the schema for ScanJob, ScanEntry and Consumer models.
	err = db.AutoMigrate(&models.ScanJob{}, &models.ScanEntry{}, &models.Consumer{})
	if err != nil {
		// Attempt to close the database connection if migration fails.
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, err
	}

	log.Println("Database schema migrated.")

	return &StoreGorm{DB: db}, nil
}

// CreateScanJob creates a new scan job in the database.
func (s *StoreGorm) CreateScanJob(scanJob *models.ScanJob) error {
	result := s.DB.Create(scanJob)
	return result.Error
}

// GetScanJob retrieves a scan job by its ID, preloading associated matches.
func (s *StoreGorm) GetScanJob(id string) (*models.ScanJob, error) {
	var scanJob models.ScanJob
	result := s.DB.First(&scanJob, "job_id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &models.NotFoundErr{}
		}
		return nil, result.Error
	}
	return &scanJob, nil
}

// CreateScanEntry creates a new scan entry in the database.
func (s *StoreGorm) CreateScanEntry(scanEntry *models.ScanEntry) error {
	result := s.DB.Create(scanEntry)
	return result.Error
}

// GetScanEntry retrieves a scan entry by its EntryID.
func (s *StoreGorm) GetScanEntry(entryID string) (*models.ScanEntry, error) {
	var scanEntry models.ScanEntry
	result := s.DB.First(&scanEntry, "entry_id = ?", entryID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &models.NotFoundErr{}
		}
		return nil, result.Error
	}
	return &scanEntry, nil
}

// GetScanEntryByFingerprint retrieves a scan entry by its Fingerprint.
func (s *StoreGorm) GetScanEntryByFingerprint(fingerprint string) (*models.ScanEntry, error) {
	var scanEntry models.ScanEntry
	result := s.DB.First(&scanEntry, "fingerprint = ?", fingerprint)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &models.NotFoundErr{}
		}
		return nil, result.Error
	}
	return &scanEntry, nil
}

// UpdateScanJob updates an existing scan job in the database.
func (s *StoreGorm) UpdateScanJob(scanJob *models.ScanJob) error {
	result := s.DB.Save(scanJob)
	return result.Error
}

// UpdateScanEntry updates an existing scan entry in the database.
func (s *StoreGorm) UpdateScanEntry(scanEntry *models.ScanEntry) error {
	result := s.DB.Save(scanEntry)
	return result.Error
}

func (s *StoreGorm) DeleteScanEntry(scanEntry *models.ScanEntry) error {
	result := s.DB.Delete(scanEntry)
	return result.Error
}

func (s *StoreGorm) ListScanEntries() ([]models.ScanEntry, error) {
	var scanEntries []models.ScanEntry
	result := s.DB.Find(&scanEntries)
	return scanEntries, result.Error
}

// UpsertConsumer creates or updates a consumer based on unique tuple.
func (s *StoreGorm) UpsertConsumer(consumer *models.Consumer) error {
	var existing models.Consumer
	tx := s.DB.Where("file_path = ? AND comm = ? AND exe = ? AND ruid = ? AND euid = ?",
		consumer.FilePath, consumer.Comm, consumer.Exe, consumer.RUID, consumer.EUID).
		First(&existing)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return s.DB.Create(consumer).Error
		}
		return tx.Error
	}
	existing.UpdatedAt = time.Now()
	// Keep latest exe/comm if they changed (rare)
	existing.Comm = consumer.Comm
	existing.Exe = consumer.Exe
	return s.DB.Save(&existing).Error
}

func (s *StoreGorm) ListConsumers(filter *models.ConsumerFilter) ([]models.Consumer, error) {
	var out []models.Consumer
	q := s.DB.Model(&models.Consumer{})
	if filter != nil && filter.FilePath != "" {
		q = q.Where("file_path = ?", filter.FilePath)
	}
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
