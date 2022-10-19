package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/jfcarter2358/ceresdb-go/connection"
)

type StormfrontApplication struct {
	ID       string                      `json:"id"`
	Node     string                      `json:"node"`
	Name     string                      `json:"name"`
	Image    string                      `json:"image"`
	Hostname string                      `json:"hostname"`
	Env      map[string]string           `json:"env"`
	Ports    map[string]string           `json:"ports"`
	Memory   int                         `json:"memory"`
	CPU      float64                     `json:"cpu"`
	Status   StormfrontApplicationStatus `json:"status"`
}

type StormfrontApplicationStatus struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Status string `json:"status"`
}

func updateApplicationStatus() error {
	data, err := connection.Query("get record stormfront.application")
	if err != nil {
		return err
	}
	connection.Host = Client.Leader.Host
	for _, appMap := range data {
		var app StormfrontApplication
		appBytes, _ := json.Marshal(appMap)
		json.Unmarshal(appBytes, &app)
		status, cpu, memory := getApplicationStatus(app)
		_, err := connection.Query(fmt.Sprintf(`patch record stormfront.application '%s' {"status":"%s","cpu":"%s","memory":"%s"}`, appMap[".id"].(string), status, cpu, memory))
		if err != nil {
			fmt.Printf("Unable to update database with status for application %s", app.ID)
		}
	}
	connection.Host = Client.Host
	return nil
}

func getApplicationStatus(app StormfrontApplication) (string, string, string) { // status, cpu, memory
	status := ""
	cpu := ""
	memory := ""

	cmd := exec.Command("/bin/sh", "-c", "docker stats --no-stream --no-trunc --all --format \"{{.CPUPerc}}||{{.MemPerc}}\"")
	var outb1 bytes.Buffer
	cmd.Stdout = &outb1
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error getting container stats: %v\n", err.Error())
	} else {
		lines := strings.Split(outb1.String(), "\n")
		if len(lines) > 1 {
			for _, line := range lines {
				parts := strings.Split(line, "||")
				if len(parts) == 2 {
					cpu = parts[0]
					memory = parts[1]
				} else {
					break
				}
			}
		}
	}

	cmd = exec.Command("/bin/sh", "-c", "docker ps --no-trunc --all --format \"{{.Names}}||{{.Status}}\"")
	var outb2 bytes.Buffer
	cmd.Stdout = &outb2
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error getting container stats: %v\n", err.Error())
	} else {
		lines := strings.Split(outb2.String(), "\n")
		if len(lines) > 1 {
			for _, line := range lines {
				parts := strings.Split(line, "||")
				if len(parts) == 2 {
					status = parts[1]
				} else {
					break
				}
			}
		}
	}

	return status, cpu, memory
}

func deployApplication(app StormfrontApplication) {
	fmt.Printf("Deploying application %s\n", app.Name)

	// Clean up any possible artifacts
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker kill %s", app.Name)).Run(); err != nil {
		fmt.Printf("No running container with name %s exists, skipping removal\n", app.Name)
	}
	if err := os.Remove(fmt.Sprintf("/var/stormfront/%s.hosts", app.Name)); err != nil {
		fmt.Printf("Could not find /var/stormfront/%s.hosts file, skipping removal\n", app.Name)
	}
	if err := os.Remove(fmt.Sprintf("/var/stormfront/%s.cid", app.Name)); err != nil {
		fmt.Printf("Could not find /var/stormfront/%s.cid file, skipping removal\n", app.Name)
	}

	dockerCommand := "docker run --net host -d --rm "
	dockerCommand += fmt.Sprintf("--name %s ", app.Name)
	dockerCommand += fmt.Sprintf("--cidfile /var/stormfront/%s.cid ", app.Name)
	dockerCommand += fmt.Sprintf("--cpus=\"%f\" ", app.CPU)
	dockerCommand += fmt.Sprintf("--memory=\"%db\" ", int(app.Memory))
	for key, val := range app.Env {
		dockerCommand += fmt.Sprintf("-e %s=%s ", key, val)
	}
	for to, from := range app.Ports {
		dockerCommand += fmt.Sprintf("-p %s:%s ", to, from)
	}
	dockerCommand += app.Image
	err := exec.Command("/bin/sh", "-c", dockerCommand).Run()
	if err != nil {
		fmt.Printf("Encountered error deploying Docker container: %v\n", err.Error())
	}

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker exec -u 0 %s sh -c \"cat /etc/hosts\"", app.Name))
	var outb bytes.Buffer
	cmd.Stdout = &outb
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error getting container /etc/hosts: %v\n", err.Error())
	}

	hostData := outb.String()
	err = os.WriteFile(fmt.Sprintf("/var/stormfront/%s.hosts", app.Name), []byte(hostData), 0644)
	if err != nil {
		fmt.Printf("Encountered error writing hosts file: %v\n", err.Error())
	}
}

func destroyApplication(app StormfrontApplication) {
	fmt.Printf("Destroying application %s\n", app.Name)
	err := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker kill %s", app.Name)).Run()
	if err != nil {
		fmt.Printf("Encountered error killing docker container: %v\n", err.Error())
	}
	err = os.Remove(fmt.Sprintf("/var/stormfront/%s.hosts", app.Name))
	if err != nil {
		fmt.Printf("Encountered error removing hosts file: %v\n", err.Error())
	}
	err = os.Remove(fmt.Sprintf("/var/stormfront/%s.cid", app.Name))
	if err != nil {
		fmt.Printf("Encountered error removing CID file: %v\n", err.Error())
	}
}

func reconcileApplications() {
	var definedApplications []StormfrontApplication
	data, err := connection.Query("get record stormfront.application")

	if err != nil {
		fmt.Printf("Unable to contact database: %v\n", err)
		return
	}

	dataBytes, _ := json.Marshal(data)
	json.Unmarshal(dataBytes, &definedApplications)

	hostContents := ""
	for _, definedApp := range definedApplications {
		hostContents += fmt.Sprintf("%s %s\n", Client.Host, definedApp.Hostname)
	}

	err = os.WriteFile("/var/stormfront/hosts", []byte(hostContents), 0644)
	if err != nil {
		panic(err)
	}

	// Check for applications that should be deployed
	for _, definedApp := range definedApplications {
		shouldBeDeployed := true
		for _, runningApp := range Client.Applications {
			if definedApp.ID == runningApp.ID {
				shouldBeDeployed = false
				break
			}
		}
		if shouldBeDeployed {
			if definedApp.Node == Client.ID {
				deployApplication(definedApp)
			}
		}
	}

	// Check for applications that should be torn down
	for _, runningApp := range Client.Applications {
		shouldBeDestroyed := true
		for _, definedApp := range definedApplications {
			if definedApp.ID == runningApp.ID {
				shouldBeDestroyed = false
				break
			}
		}
		if shouldBeDestroyed {
			if runningApp.Node == Client.ID {
				destroyApplication(runningApp)
			}
		}
	}

	for _, runningApp := range Client.Applications {
		for _, definedApp := range definedApplications {
			if runningApp.ID == definedApp.ID {
				if !reflect.DeepEqual(runningApp.Env, definedApp.Env) || runningApp.Image != definedApp.Image || !reflect.DeepEqual(runningApp.Ports, definedApp.Ports) {
					fmt.Printf("Performing application update for %s\n", runningApp.Name)
					destroyApplication(runningApp)
					deployApplication(definedApp)
				}
			}
		}
	}

	// Update /etc/hosts for running applications
	for _, definedApp := range definedApplications {
		if Client.ID == definedApp.ID {
			err = exec.Command("/bin/sh", "-c", fmt.Sprintf("docker exec -u 0 %s sh -c \"echo \\\"$(cat /var/stormfront/%s.hosts)\\n$(cat /var/stormfront/hosts)\\\" > /etc/hosts\"", definedApp.Name, definedApp.Name)).Run()
			if err != nil {
				fmt.Printf("Encountered error copying to hosts file: %v\n", err.Error())
				continue
			}
		}
	}
}
