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
	fmt.Printf("old file length: %v\nold Entry Len: %v\n, new Entry Len: %v\n, new file len: %v\n", len(file), entry.EndLine-entry.StartLine+1, len(newValue), newFileLength)
	modFile := make([]byte, newFileLength)
	fmt.Printf("modifying %v with %v", string(file[entry.StartLine:entry.EndLine]), string(newValue))
	copy(modFile[:entry.StartLine], file[:entry.StartLine])
	fmt.Printf("New modFile before start: %v\n", string(modFile[:entry.StartLine]))
	copy(modFile[entry.StartLine:entry.StartLine+len(newValue)], newValue)
	fmt.Printf("New modFile up until beginning: %v\n", string(modFile[:entry.StartLine+len(newValue)]))
	copy(modFile[entry.StartLine+len(newValue):], file[entry.EndLine+1:])
	fmt.Printf("New modFile after end: %v\n", string(modFile))

	err = os.WriteFile(entry.FilePath, modFile, 0644)
	if err != nil {
		return err
	}
	// Update Scan Entry
	entry.EndLine = entry.StartLine + len(newValue)
	err = s.scanService.UpdateScanEntry(entry)
	if err != nil {
		return err
	}
	return nil
}
