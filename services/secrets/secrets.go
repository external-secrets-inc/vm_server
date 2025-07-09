package secrets

import (
	"fmt"
	"os"
	"vm-server/models"
	"vm-server/services/scan"
)

// Service is a service for secrets
type Service struct {
	scanService *scan.Service
}

// NewService creates a new secret service
func NewService(scanService *scan.Service) *Service {
	return &Service{
		scanService: scanService,
	}
}

func (s *Service) UpdateVersion(newValue []byte, entry *models.ScanEntry) error {
	file, err := os.ReadFile(entry.FilePath)
	if err != nil {
		return err
	}
	newFileLength := len(file) - (entry.EndLine - entry.StartLine + 1) + len(newValue)
	modFile := make([]byte, newFileLength)
	fmt.Printf("modifying %v with %v", string(file[entry.StartLine:entry.EndLine]), string(newValue))
	copy(modFile[:entry.StartLine], file[:entry.StartLine])
	copy(modFile[entry.StartLine:entry.StartLine+len(newValue)], newValue)
	copy(modFile[entry.StartLine+len(newValue):], file[entry.EndLine:])
	return os.WriteFile(entry.FilePath, modFile, 0644)
}
