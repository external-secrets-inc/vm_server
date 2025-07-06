package models

import "time"

type ScanResponse struct {
	ID string `json:"id"`
}

type ScanIdResponse struct {
	ID         string      `json:"id"`
	Status     ScanStatus  `json:"status"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
	FinishedAt time.Time   `json:"finishedAt"`
	Match      []ScanMatch `json:"match"`
}

type ScanStatus string

const (
	ScanStatusPending ScanStatus = "pending"
	ScanStatusRunning ScanStatus = "running"
	ScanStatusFailed  ScanStatus = "failed"
	ScanStatusSuccess ScanStatus = "success"
)

type ScanMatch struct {
	ID              string              `json:"id"`
	FilePath        string              `json:"filePath"`
	ProviderMapping ScanProviderMapping `json:"providerMapping"`
}

type ScanProviderMapping struct {
	SecretStore string    `json:"secretStore"`
	RemoteRef   RemoteRef `json:"remoteRef"`
}

type RemoteRef struct {
	Key string `json:"key"`
}
