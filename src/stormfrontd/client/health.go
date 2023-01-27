package client

import (
	"fmt"
	"time"
)

func HealthCheckFollower() {
	for {
		reconcileApplications()
		updateApplicationStatus()
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}

func HealthCheckLeader() {
	for {
		err := updateSuccession()
		if err != nil {
			fmt.Printf("Encountered error updating succession: %v", err)
		}
		updateApplicationStatus()
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}
