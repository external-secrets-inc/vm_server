package jobs

import (
	"vm-server/models"
	consumer "vm-server/services/consumer"
)

// ConsumerScanner defines the structure for our scanning job runner.
type ConsumerScanner struct {
	Service *consumer.Service
}

func (s *ConsumerScanner) PerformScan(job *models.ConsumerJob, req *models.ConsumerRequest) {
	// TODO: This is where the trigger to add new config for the ebpf probe
	// / start the ebpf probe
	//  / get info from the already running ebpf probe
	// should happen
}
