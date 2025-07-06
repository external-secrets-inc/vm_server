package models

type ScanRequest struct {
	Paths     []string `json:"paths"`
	Regexes   []string `json:"regexes"`
	Threshold int      `json:"threshold"`
}
