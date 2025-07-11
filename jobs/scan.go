package jobs

import (
	"crypto/sha3"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"vm-server/models"
	scanService "vm-server/services/scan"

	"github.com/google/uuid"
)

const startingChars = " \t'\"="
const endingChars = " \t\n'\""

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
	filesToScan := make(chan string)
	var wg sync.WaitGroup
	results := make(chan *models.ScanEntry, 100) // Buffer the channel

	// Start worker pool
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.scanWorker(&wg, filesToScan, results, req.Regexes, req.Threshold, job.ID)
	}

	// Collect file paths
	go func() {
		for _, path := range req.Paths {
			filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("Failed to walk directory %s: %v", path, err)
					return nil // Continue walking
				}
				if !d.IsDir() && d.Type().IsRegular() {
					filesToScan <- path
				}
				return nil
			})
		}
		close(filesToScan)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	for entry := range results {
		existing, err := s.Service.GetScanEntryByFingerprint(entry.Fingerprint)
		if err != nil {
			if errors.Is(err, &models.NotFoundErr{}) {
				err = s.Service.CreateScanEntry(entry)
				if err != nil {
					log.Printf("failed to create scan entry %s: %v", entry.EntryID, err)
				} else {
					job.Match = append(job.Match, models.ScanMatchResumed{EntryID: entry.EntryID})
				}
			} else {
				log.Printf("failed to get scan entry %s: %v", entry.EntryID, err)
			}
		} else {
			// Update existing entry if it already exists
			err = s.Service.UpdateScanEntry(existing)
			if err != nil {
				log.Printf("failed to update scan entry %s: %v", entry.EntryID, err)
			} else {
				job.Match = append(job.Match, models.ScanMatchResumed{EntryID: existing.EntryID})
			}
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

func isText(s []byte) bool {
	n := 1024
	if len(s) < n {
		n = len(s)
	}
	return utf8.Valid(s[:n])
}

func (s *Scanner) scanWorker(wg *sync.WaitGroup, files <-chan string, results chan<- *models.ScanEntry, regexes []string, threshold int, jobID uint) {
	defer wg.Done()
	for path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("failed to read file %s: %v", path, err)
			continue
		}

		if !isText(content) {
			continue
		}

		entries := make(map[string]*models.ScanEntry)
		matchCount := make(map[string]int)

		for _, regex := range regexes {
			re, err := regexp.Compile(regex)
			if err != nil {
				log.Printf("failed to compile regex %s: %v", regex, err)
				continue
			}

			matches := re.FindAllIndex(content, -1)
			for _, match := range matches {
				if (match[0] != 0 && !strings.ContainsRune(startingChars, rune(content[match[0]-1]))) ||
					(match[1] != len(content) && !strings.ContainsRune(endingChars, rune(content[match[1]]))) {
					continue
				}

				idx := fmt.Sprintf("%v@%v:%v", path, match[0], match[1])
				fingerprint := sha3.New224().Sum([]byte(idx))
				id := uuid.New().String()

				if existing, ok := entries[idx]; ok {
					existing.MatchedRegexes = append(existing.MatchedRegexes, regex)
				} else {
					entries[idx] = &models.ScanEntry{
						EntryID:        id,
						Fingerprint:    fmt.Sprintf("%x", fingerprint),
						FilePath:       path,
						MatchedRegexes: models.StringSlice{regex},
						StartLine:      match[0],
						EndLine:        match[1],
						ScanJobID:      jobID,
					}
				}
				matchCount[idx]++
			}
		}

		for idx, count := range matchCount {
			if count >= threshold {
				results <- entries[idx]
			}
		}
	}
}
