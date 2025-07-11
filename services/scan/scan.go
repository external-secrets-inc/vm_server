package scan

import (
	"vm-server/models"
	"vm-server/store/schema"
)

// Service is the business logic layer for scans.
type Service struct {
	Store schema.Store
}

// NewService creates a new scan service.
func NewService(s schema.Store) *Service {
	return &Service{Store: s}
}

// CreateScanJob creates a new scan job.
func (s *Service) CreateScanJob(scanJob *models.ScanJob) error {
	// Business logic for creating a job can go here.
	return s.Store.CreateScanJob(scanJob)
}

// GetScanJob retrieves a scan job by its ID.
func (s *Service) GetScanJob(id string) (*models.ScanJob, error) {
	return s.Store.GetScanJob(id)
}

// CreateScanEntry creates a new scan entry.
func (s *Service) CreateScanEntry(scanEntry *models.ScanEntry) error {
	// Business logic for creating an entry can go here.
	return s.Store.CreateScanEntry(scanEntry)
}

// GetScanEntry retrieves a scan entry by its EntryID.
func (s *Service) GetScanEntry(entryID string) (*models.ScanEntry, error) {
	return s.Store.GetScanEntry(entryID)
}

// GetScanEntry retrieves a scan entry by its EntryID.
func (s *Service) GetScanEntryByFingerprint(fingerprint string) (*models.ScanEntry, error) {
	return s.Store.GetScanEntryByFingerprint(fingerprint)
}

// UpdateScanEntry updates an existing scan entry.
func (s *Service) UpdateScanEntry(scanEntry *models.ScanEntry) error {
	return s.Store.UpdateScanEntry(scanEntry)
}

// UpdateScanJob updates an existing scan job.
func (s *Service) UpdateScanJob(scanJob *models.ScanJob) error {
	// Business logic for updating a job can go here.
	return s.Store.UpdateScanJob(scanJob)
}

func (s *Service) DeleteScanEntry(scanEntry *models.ScanEntry) error {
	return s.Store.DeleteScanEntry(scanEntry)
}

func (s *Service) ListScanEntries() ([]models.ScanEntry, error) {
	return s.Store.ListScanEntries()
}
