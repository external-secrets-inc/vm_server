package schema

import "vm-server/models"

type StoreProvider interface {
	NewStore() (Store, error)
}

type Store interface {
	// Scan Jobs
	CreateScanJob(scanJob *models.ScanJob) error
	GetScanJob(id string) (*models.ScanJob, error)
	UpdateScanJob(scanJob *models.ScanJob) error
	// Scan Entry
	CreateScanEntry(scanEntry *models.ScanEntry) error
	GetScanEntry(entryID string) (*models.ScanEntry, error)
	GetScanEntryByFingerprint(fingerprint string) (*models.ScanEntry, error)
	UpdateScanEntry(scanEntry *models.ScanEntry) error
	DeleteScanEntry(scanEntry *models.ScanEntry) error
	ListScanEntries() ([]models.ScanEntry, error)
	// Consumer Jobs
	GetConsumerJob(id string) (*models.ConsumerJob, error)
	CreateConsumerJob(consumerJob *models.ConsumerJob) error
	UpdateConsumerJob(scanJob *models.ConsumerJob) error
	// Consumer Entry
	GetConsumerEntry(entryId string) (*models.ConsumerEntry, error)
	CreateConsumerEntry(consumerEntry *models.ConsumerEntry) error
	UpdateConsumerEntry(consumerEntry *models.ConsumerEntry) error
	DeleteConsumerEntry(consumerEntry *models.ConsumerEntry) error
	ListConsumerEntries() ([]models.ConsumerEntry, error)
}
