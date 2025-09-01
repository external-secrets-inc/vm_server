package models

type ConsumerRequest struct {
	Location ConsumerLocation `json:"location"`
	Paths    []string         `json:"paths"`
}

type ConsumerLocation struct {
	Kind       string    `json:"kind"`
	Name       string    `json:"name"`
	APIVersion string    `json:"apiVersion"`
	RemoteRef  RemoteRef `json:"remoteRef"`
}

type ConsumerRemoteRef struct {
	Key      string `json:"key"`
	Property string `json:"property"`
}
