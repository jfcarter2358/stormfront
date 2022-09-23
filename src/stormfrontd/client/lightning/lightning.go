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
	ID      string   `json:"id"`
	Stdout  string   `json:"stdout"`
	Stderr  string   `json:"stderr"`
	Error   string   `json:"error"`
	Command []string `json:"command"`
	Status  string   `json:"status"`
}

type BoltConstructor struct {
	Command []string `json:"command"`
}

func CreateBolt(command []string) (Bolt, int) {
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

	var cmd *exec.Cmd

	if len(bolt.Command) == 0 {
		bolt.Error = "Invalid number of elements in the command string: 0"
		bolt.Status = BOLT_FAILURE_STATUS
		return
	} else if len(bolt.Command) == 1 {
		cmd = exec.Command(bolt.Command[0])
	} else {
		cmd = exec.Command(bolt.Command[0], bolt.Command[1:]...)
	}

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
