package lightning

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/google/uuid"
)

// valid bolt statuses:
// Pending (waiting to be run)
// Running
// Success
// Failure

var Bolts []Bolt

const BOLT_PENDING_STATUS = "Pending"
const BOLT_RUNNING_STATUS = "Running"
const BOLT_SUCCESS_STATUS = "Success"
const BOLT_FAILURE_STATUS = "Failure"

type Bolt struct {
	ID      string `json:"id"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Error   string `json:"error"`
	Command string `json:"command"`
	Status  string `json:"status"`
}

type BoltConstructor struct {
	Command string `json:"command"`
}

func CreateBolt(command string) (Bolt, int) {
	bolt := Bolt{
		ID:      uuid.New().String(),
		Stdout:  "",
		Stderr:  "",
		Error:   "",
		Command: command,
		Status:  BOLT_PENDING_STATUS,
	}

	idx := len(Bolts)
	Bolts = append(Bolts, bolt)

	return bolt, idx
}

func RunBolt(bolt *Bolt) {
	bolt.Status = BOLT_RUNNING_STATUS

	cmd := exec.Command("/bin/sh", "-c", bolt.Command)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		bolt.Error = err.Error()
		bolt.Status = BOLT_FAILURE_STATUS
	} else {
		bolt.Status = BOLT_SUCCESS_STATUS
	}

	bolt.Stdout = outb.String()
	bolt.Stderr = errb.String()
}

func GetBolt(boltId string) (Bolt, error) {
	for _, bolt := range Bolts {
		if boltId == bolt.ID {
			return bolt, nil
		}
	}
	return Bolt{}, errors.New("no bolt found with ID")
}
