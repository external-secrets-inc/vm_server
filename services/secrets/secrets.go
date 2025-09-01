package secrets

import (
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
	copy(modFile[:entry.StartLine], file[:entry.StartLine])
	copy(modFile[entry.StartLine:entry.StartLine+len(newValue)], newValue)
	copy(modFile[entry.StartLine+len(newValue):], file[entry.EndLine+1:])

	err = os.WriteFile(entry.FilePath, modFile, 0644)
	if err != nil {
		return err
	}
	// Update Scan Entry
	entry.EndLine = entry.StartLine + len(newValue) - 1
	err = s.scanService.UpdateScanEntry(entry)
	if err != nil {
		return err
	}
	return nil
}
