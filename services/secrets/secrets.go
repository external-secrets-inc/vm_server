package secrets

import (
	"vm-server/models"
	"vm-server/services/scan"
)

// SecretService is a service for secrets
type SecretService struct {
	scanService *scan.Service
}

// NewSecretService creates a new secret service
func NewSecretService(scanService *scan.Service) *SecretService {
	return &SecretService{
		scanService: scanService,
	}
}

func (s *SecretService) UpdateVersion(newValue []byte, entry *models.ScanEntry) error {

	return nil
}
