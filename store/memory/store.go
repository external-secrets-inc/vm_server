package memory

import (
	"sync"
	"time"
	"vm-server/models"
	"vm-server/store/schema"

	uuid "github.com/google/uuid"
)

const MAX_SCAN_JOB_SIZE = 1000
const MAX_SCAN_ENTRY_SIZE = 10000

const CLEANUP_TO = 0.8

// Store manages the database connection and operations.
type StoreMemory struct {
	scanJobAge             map[uuid.UUID]time.Time
	scanEntryAge           map[uuid.UUID]time.Time
	scanJobs               map[uuid.UUID]*models.ScanJob
	scanEntries            map[uuid.UUID]*models.ScanEntry
	scanEntriesFingerprint map[string]uuid.UUID
	mu                     sync.RWMutex
}

// NewStore initializes the database connection and migrates the schema.
func NewStore() (schema.Store, error) {
	return &StoreMemory{
		scanJobs:               make(map[uuid.UUID]*models.ScanJob),
		scanEntries:            make(map[uuid.UUID]*models.ScanEntry),
		scanEntriesFingerprint: make(map[string]uuid.UUID),
		scanJobAge:             make(map[uuid.UUID]time.Time),
		scanEntryAge:           make(map[uuid.UUID]time.Time),
	}, nil
}

// cleanupScanJobs removes the oldest scan jobs until the count is down to the target size.
func (s *StoreMemory) cleanupScanJobs() {
	targetSize := int(MAX_SCAN_JOB_SIZE * CLEANUP_TO)
	for len(s.scanJobs) > targetSize {
		var oldestID uuid.UUID
		var oldestTime time.Time
		first := true
		for id, t := range s.scanJobAge {
			if first {
				oldestTime = t
				oldestID = id
				first = false
				continue
			}
			if t.Before(oldestTime) {
				oldestTime = t
				oldestID = id
			}
		}
		delete(s.scanJobs, oldestID)
		delete(s.scanJobAge, oldestID)
	}
}

// cleanupScanEntries removes the oldest scan entries until the count is down to the target size.
func (s *StoreMemory) cleanupScanEntries() {
	targetSize := int(MAX_SCAN_ENTRY_SIZE * CLEANUP_TO)
	for len(s.scanEntries) > targetSize {
		var oldestID uuid.UUID
		var oldestTime time.Time
		first := true
		for id, t := range s.scanEntryAge {
			if first {
				oldestTime = t
				oldestID = id
				first = false
				continue
			}
			if t.Before(oldestTime) {
				oldestTime = t
				oldestID = id
			}
		}
		delete(s.scanEntriesFingerprint, s.scanEntries[oldestID].Fingerprint)
		delete(s.scanEntries, oldestID)
		delete(s.scanEntryAge, oldestID)
	}
}

// CreateScanJob creates a new scan job in the database.
func (s *StoreMemory) CreateScanJob(scanJob *models.ScanJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.scanJobs) >= MAX_SCAN_JOB_SIZE {
		s.cleanupScanJobs()
	}
	uuid, err := uuid.Parse(scanJob.JobID)
	if err != nil {
		return err
	}
	if _, ok := s.scanJobs[uuid]; ok {
		return &models.AlreadyExistErr{}
	}
	s.scanJobs[uuid] = scanJob
	s.scanJobAge[uuid] = time.Now()
	return nil
}

// GetScanJob retrieves a scan job by its ID, preloading associated matches.
func (s *StoreMemory) GetScanJob(id string) (*models.ScanJob, error) {
	s.mu.RLock()
	uuid, err := uuid.Parse(id)
	if err != nil {
		s.mu.RUnlock()
		return nil, err
	}
	scanJob, ok := s.scanJobs[uuid]
	if !ok {
		s.mu.RUnlock()
		return nil, &models.NotFoundErr{}
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanJobAge[uuid] = time.Now()
	return scanJob, nil
}

// CreateScanEntry creates a new scan entry in the database.
func (s *StoreMemory) CreateScanEntry(scanEntry *models.ScanEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.scanEntries) >= MAX_SCAN_ENTRY_SIZE {
		s.cleanupScanEntries()
	}
	uuid, err := uuid.Parse(scanEntry.EntryID)
	if err != nil {
		return err
	}
	if _, ok := s.scanEntries[uuid]; ok {
		return &models.AlreadyExistErr{}
	}
	s.scanEntries[uuid] = scanEntry
	s.scanEntriesFingerprint[scanEntry.Fingerprint] = uuid
	s.scanEntryAge[uuid] = time.Now()
	return nil
}

// GetScanEntry retrieves a scan entry by its EntryID.
func (s *StoreMemory) GetScanEntry(entryID string) (*models.ScanEntry, error) {
	s.mu.RLock()
	uuid, err := uuid.Parse(entryID)
	if err != nil {
		s.mu.RUnlock()
		return nil, err
	}
	scanEntry, ok := s.scanEntries[uuid]
	if !ok {
		s.mu.RUnlock()
		return nil, &models.NotFoundErr{}
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanEntryAge[uuid] = time.Now()
	return scanEntry, nil
}

// GetScanEntry retrieves a scan entry by its EntryID.
func (s *StoreMemory) GetScanEntryByFingerprint(fingerprint string) (*models.ScanEntry, error) {
	s.mu.RLock()
	entryId, ok := s.scanEntriesFingerprint[fingerprint]
	if !ok {
		s.mu.RUnlock()
		return nil, &models.NotFoundErr{}
	}
	scanEntry, ok := s.scanEntries[entryId]
	if !ok {
		s.mu.RUnlock()
		return nil, &models.NotFoundErr{}
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanEntryAge[entryId] = time.Now()
	return scanEntry, nil
}

// UpdateScanJob updates an existing scan job in the database.
func (s *StoreMemory) UpdateScanJob(scanJob *models.ScanJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	uuid, err := uuid.Parse(scanJob.JobID)
	if err != nil {
		return err
	}
	if _, ok := s.scanJobs[uuid]; !ok {
		return &models.NotFoundErr{}
	}
	s.scanJobs[uuid] = scanJob
	s.scanJobAge[uuid] = time.Now()
	return nil
}

// UpdateScanEntry updates an existing scan entry in the database.
func (s *StoreMemory) UpdateScanEntry(scanEntry *models.ScanEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	uuid, err := uuid.Parse(scanEntry.EntryID)
	if err != nil {
		return err
	}
	if _, ok := s.scanEntries[uuid]; !ok {
		return &models.NotFoundErr{}
	}
	s.scanEntries[uuid] = scanEntry
	s.scanEntryAge[uuid] = time.Now()
	return nil
}

func (s *StoreMemory) DeleteScanEntry(scanEntry *models.ScanEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	uuid, err := uuid.Parse(scanEntry.EntryID)
	if err != nil {
		return err
	}
	if _, ok := s.scanEntries[uuid]; !ok {
		// Already deleted, operation ok
		return nil
	}
	delete(s.scanEntries, uuid)
	delete(s.scanEntryAge, uuid)
	return nil
}

func (s *StoreMemory) ListScanEntries() ([]models.ScanEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var scanEntries []models.ScanEntry
	for _, scanEntry := range s.scanEntries {
		scanEntries = append(scanEntries, *scanEntry)
	}
	return scanEntries, nil
}
