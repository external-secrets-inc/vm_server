package schema

import "vm-server/models"

type StoreProvider interface {
	NewStore() (Store, error)
}

type Store interface {
	CreateScanJob(scanJob *models.ScanJob) error
	GetScanJob(id string) (*models.ScanJob, error)
	CreateScanEntry(scanEntry *models.ScanEntry) error
	GetScanEntry(entryID string) (*models.ScanEntry, error)
	GetScanEntryByFingerprint(fingerprint string) (*models.ScanEntry, error)
	UpdateScanJob(scanJob *models.ScanJob) error
	UpdateScanEntry(scanEntry *models.ScanEntry) error
	DeleteScanEntry(scanEntry *models.ScanEntry) error
	ListScanEntries() ([]models.ScanEntry, error)

	// Consumers
	UpsertConsumer(consumer *models.Consumer) error
	ListConsumers(filter *models.ConsumerFilter) ([]models.Consumer, error)
}
