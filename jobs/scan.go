package jobs

import (
	"bytes"
	"crypto/sha3"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"unicode/utf8"
	"vm-server/models"
	scanService "vm-server/services/scan"

	"gorm.io/gorm"
)

// Scanner defines the structure for our scanning job runner.
type Scanner struct {
	Service *scanService.Service
}

// NewScanner creates a new scanner job runner.
func NewScanner(s *scanService.Service) *Scanner {
	return &Scanner{Service: s}
}

func (s *Scanner) Cleanup() error {
	// Get All existing entries
	entries, err := s.Service.ListScanEntries()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		content, err := os.ReadFile(entry.FilePath)
		if err != nil {
			return err
		}
		for _, reg := range entry.MatchedRegexes {
			re, err := regexp.Compile(reg)
			if err != nil {
				return err
			}
			// If they no longer match any regex, delete.
			if match := re.Match(content[entry.StartLine:entry.EndLine]); !match {
				if err := s.Service.DeleteScanEntry(&entry); err != nil {
					return err
				}
				// stop looking for regexes and go to the next entry
				break
			}
		}
	}
	return nil
}

// PerformScan runs the actual scan logic in the background.
func (s *Scanner) PerformScan(job *models.ScanJob, req *models.ScanRequest) {
	log.Printf("Starting scan for job ID: %s", job.JobID)
	job.Status = "in_progress"
	job.UpdatedAt = time.Now()
	if err := s.Service.UpdateScanJob(job); err != nil {
		log.Printf("Failed to update job status to in_progress for job %s: %v", job.JobID, err)
		return
	}
	start_time := time.Now()
	for _, path := range req.Paths {
		log.Printf("Scanning path: %s", path)
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("Failed to walk directory %s: %v", path, err) // Continue for other paths
				err = nil
				return err
			}
			if !d.Type().IsRegular() {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			// Check if file content is binary
			if !isText(content) {
				return nil
			}
			entries := map[string]*models.ScanEntry{}
			matchCount := map[string]int{}
			for _, regex := range req.Regexes {
				re, err := regexp.Compile(regex)
				if err != nil {
					return fmt.Errorf("failed to compile regex %s: %w", regex, err)
				}
				matches := re.FindAllIndex(content, -1)
				for _, match := range matches {
					idx := fmt.Sprintf("%v@%v:%v", path, match[0], match[1])
					entryId := sha3.New224().Sum([]byte(idx))
					entry := models.ScanEntry{
						EntryID:        fmt.Sprintf("%x", entryId),
						FilePath:       path,
						MatchedRegexes: models.StringSlice{regex},
						StartLine:      match[0],
						EndLine:        match[1],
						ScanJobID:      job.ID,
					}
					matchCount[idx]++
					if existing, ok := entries[idx]; ok {
						// Update entry to include matched Regex
						existing.MatchedRegexes = append(existing.MatchedRegexes, regex)
						entries[idx] = existing
					} else {
						// Create a new  entry
						entries[idx] = &entry
					}
				}

			}
			// Now here, entries contain a list of all scan Entries that might have enough matches
			// According to the threshold criteria. We just need to filter them
			// To properly create/update scan entries.
			for idx, count := range matchCount {
				if count >= req.Threshold {
					entry := entries[idx]
					existing, err := s.Service.GetScanEntry(entry.EntryID)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							err = s.Service.CreateScanEntry(entry)
							if err != nil {
								return fmt.Errorf("failed to create scan entry %s: %w", entry.EntryID, err)
							}
						} else {
							return fmt.Errorf("failed to get scan entry %s: %w", entry.EntryID, err)
						}
					} else {
						// Update existing entry if already exists
						err = s.Service.UpdateScanEntry(existing)
						if err != nil {
							return fmt.Errorf("failed to update scan entry %s: %w", entry.EntryID, err)
						}
					}
					job.Match = append(job.Match, *entry)
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("Failed to scan directory %s: %v", path, err)
		}
	}
	duration := time.Since(start_time)
	log.Printf("Scan completed for job %s in %s", job.JobID, duration)
	job.FinishedAt = time.Now()
	job.UpdatedAt = time.Now()
	job.Status = "completed"
	if err := s.Service.UpdateScanJob(job); err != nil {
		log.Printf("Failed to update job status to completed/failed for job %s: %v", job.JobID, err)
	}
}

// isText checks if content is text by looking for null bytes and validating UTF-8.
func isText(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	// A single null byte is a strong indicator of a binary file.
	if bytes.Contains(content, []byte{0}) {
		return false
	}
	// Check if the file is valid UTF-8.
	return utf8.Valid(content)
}
