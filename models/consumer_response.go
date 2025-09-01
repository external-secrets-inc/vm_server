package models

import "time"

type ConsumerResponse struct {
	ID string `json:"id"`
}

type ConsumerIdResponse struct {
	ID         string     `json:"id"`
	Status     ScanStatus `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	FinishedAt time.Time  `json:"finishedAt"`
	Consumers  []Consumer `json:"consumers"`
}

type Consumer struct {
	Hostname       string `json:"hostname"`
	Executable     string `json:"executable"`
	User           string `json:"user"`
	PID            string `json:"pid"`
	StartTimestamp string `json:"startTimestamp"`
}
