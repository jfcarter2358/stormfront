package client

import (
	"fmt"
	"time"
)

func HealthCheckFollower() {
	for {
		reconcileApplications()
		updateApplicationStatus()
		err := updateSystemInfo()
		if err != nil {
			fmt.Printf("Encountered error getting system info: %v\n", err)
		}
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}

func HealthCheckLeader() {
	for {
		err := updateSuccession()
		if err != nil {
			fmt.Printf("Encountered error updating succession: %v\n", err)
		}
		updateApplicationStatus()
		err = updateSystemInfo()
		if err != nil {
			fmt.Printf("Encountered error getting system info: %v\n", err)
		}
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}
