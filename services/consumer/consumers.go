package consumer

import (
	"vm-server/models"
	"vm-server/store/schema"
)

// Service is the business logic layer for consumers.
type Service struct {
	Store schema.Store
}

// NewService creates a new consumer service.
func NewService(s schema.Store) *Service {
	return &Service{Store: s}
}

// CreateConsumerJob creates a new consumer job.
func (s *Service) CreateConsumerJob(consumerJob *models.ConsumerJob) error {
	// Business logic for creating a job can go here.
	return s.Store.CreateConsumerJob(consumerJob)
}

// GetConsumerJob retrieves a consumer job by its ID.
func (s *Service) GetConsumerJob(id string) (*models.ConsumerJob, error) {
	return s.Store.GetConsumerJob(id)
}

// CreateConsumerEntry creates a new consumer entry.
func (s *Service) CreateConsumerEntry(consumerEntry *models.ConsumerEntry) error {
	// Business logic for creating an entry can go here.
	return s.Store.CreateConsumerEntry(consumerEntry)
}

// GetConsumerEntry retrieves a consumer entry by its EntryID.
func (s *Service) GetConsumerEntry(entryID string) (*models.ConsumerEntry, error) {
	return s.Store.GetConsumerEntry(entryID)
}

// UpdateConsumerEntry updates an existing consumer entry.
func (s *Service) UpdateConsumerEntry(consumerEntry *models.ConsumerEntry) error {
	return s.Store.UpdateConsumerEntry(consumerEntry)
}

// UpdateConsumerJob updates an existing consumer job.
func (s *Service) UpdateConsumerJob(consumerJob *models.ConsumerJob) error {
	// Business logic for updating a job can go here.
	return s.Store.UpdateConsumerJob(consumerJob)
}

func (s *Service) DeleteConsumerEntry(consumerEntry *models.ConsumerEntry) error {
	return s.Store.DeleteConsumerEntry(consumerEntry)
}

func (s *Service) ListConsumerEntries() ([]models.ConsumerEntry, error) {
	return s.Store.ListConsumerEntries()
}
