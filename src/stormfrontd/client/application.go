package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"stormfrontd/config"
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
	connection.Host = Client.Leader.Host
	data, err := connection.Query("get record stormfront.application")
	if err != nil {
		return err
	}
	for _, appMap := range data {
		var app StormfrontApplication
		appBytes, _ := json.Marshal(appMap)
		json.Unmarshal(appBytes, &app)
		if app.Node != Client.ID {
			continue
		}
		status, cpu, memory := getApplicationStatus(app)
		_, err := connection.Query(fmt.Sprintf(`patch record stormfront.application '%s' {"status": {"status":"%s","cpu":"%s","memory":"%s"}}`, appMap[".id"].(string), status, cpu, memory))
		if err != nil {
			fmt.Printf("Unable to update database with status for application %s", app.ID)
		}
	}
	connection.Host = config.Config.CeresDBHost
	return nil
}

func getApplicationStatus(app StormfrontApplication) (string, string, string) { // status, cpu, memory
	status := ""
	cpu := "-1"
	memory := "-1"

	var cmd *exec.Cmd
	if config.Config.ContainerEngine == "docker" {
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s stats %s --no-stream --no-trunc --format \"{{.CPUPerc}}||{{.MemPerc}}\"", config.Config.ContainerEngine, app.Name))
	} else {
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s stats %s --no-stream --format \"{{.CPUPerc}}||{{.MemPerc}}\"", config.Config.ContainerEngine, app.Name))
	}
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

	if config.Config.ContainerEngine == "docker" {
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s ps --no-trunc --all --format \"{{.Names}}||{{.Status}}\"", config.Config.ContainerEngine))
	} else {
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s ps --all --format \"{{.Names}}||{{.Status}}\"", config.Config.ContainerEngine))
	}
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
					if parts[0] == app.Name {
						status = parts[1]
					}
				} else {
					break
				}
			}
		}
	}

	return status, cpu, memory
}

func deployApplication(app StormfrontApplication, shouldAppend bool) {
	fmt.Printf("Deploying application %s\n", app.Name)

	// Clean up any possible artifacts
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s kill %s", config.Config.ContainerEngine, app.Name)).Run(); err != nil {
		fmt.Printf("No running container with name %s exists, skipping kill\n", app.Name)
	}
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s rm %s", config.Config.ContainerEngine, app.Name)).Run(); err != nil {
		fmt.Printf("No running container with name %s exists, skipping removal\n", app.Name)
	}
	if err := os.Remove(fmt.Sprintf("/var/stormfront/%s.hosts", app.Name)); err != nil {
		fmt.Printf("Could not find /var/stormfront/%s.hosts file, skipping removal\n", app.Name)
	}

	// dockerCommand := fmt.Sprintf("%s run --net host -d --rm ", config.Config.ContainerEngine)
	dockerCommand := fmt.Sprintf("%s run --net host -d ", config.Config.ContainerEngine)
	dockerCommand += fmt.Sprintf("--name %s ", app.Name)
	dockerCommand += fmt.Sprintf("--cpus=\"%f\" ", app.CPU)
	dockerCommand += fmt.Sprintf("--memory=\"%db\" ", int(app.Memory))
	for key, val := range app.Env {
		dockerCommand += fmt.Sprintf("-e %s=%s ", key, val)
	}
	for to, from := range app.Ports {
		dockerCommand += fmt.Sprintf("-p %s:%s ", to, from)
	}
	dockerCommand += app.Image
	fmt.Printf("Docker command: %s\n", dockerCommand)
	cmd := exec.Command("/bin/sh", "-c", dockerCommand)
	var outb1, errb1 bytes.Buffer
	cmd.Stdout = &outb1
	cmd.Stderr = &errb1
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error deploying Docker container: %v\n", err.Error())
		fmt.Printf("STDOUT: %s\n", outb1.String())
		fmt.Printf("STDERR: %s\n", errb1.String())
		return
	}

	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s exec -u 0 %s sh -c \"cat /etc/hosts\"", config.Config.ContainerEngine, app.Name))
	var outb2, errb2 bytes.Buffer
	cmd.Stdout = &outb2
	cmd.Stderr = &errb2
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error getting container /etc/hosts: %v\n", err.Error())
		fmt.Printf("STDOUT: %s\n", outb2.String())
		fmt.Printf("STDERR: %s\n", errb2.String())
		return
	}

	hostData := outb2.String()
	err = os.WriteFile(fmt.Sprintf("/var/stormfront/%s.hosts", app.Name), []byte(hostData), 0644)
	if err != nil {
		fmt.Printf("Encountered error writing hosts file: %v\n", err.Error())
		return
	}

	if shouldAppend {
		Client.Applications = append(Client.Applications, app)
	}
}

func destroyApplication(app StormfrontApplication) {
	fmt.Printf("Destroying application %s\n", app.Name)
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s kill %s", config.Config.ContainerEngine, app.Name))
	var outb1, errb1 bytes.Buffer
	cmd.Stdout = &outb1
	cmd.Stderr = &errb1
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error killing container: %v\n", err.Error())
		fmt.Printf("STDOUT: %s\n", outb1.String())
		fmt.Printf("STDERR: %s\n", errb1.String())
	}
	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s rm %s", config.Config.ContainerEngine, app.Name))
	var outb2, errb2 bytes.Buffer
	cmd.Stdout = &outb1
	cmd.Stderr = &errb1
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error removing container: %v\n", err.Error())
		fmt.Printf("STDOUT: %s\n", outb2.String())
		fmt.Printf("STDERR: %s\n", errb2.String())
	}
	err = os.Remove(fmt.Sprintf("/var/stormfront/%s.hosts", app.Name))
	if err != nil {
		fmt.Printf("Encountered error removing hosts file: %v\n", err.Error())
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
				deployApplication(definedApp, true)
			}
		}
	}

	toRemove := []int{}

	// Check for applications that should be torn down
	for idx, runningApp := range Client.Applications {
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
				toRemove = append(toRemove, idx)
			}
		}
	}

	// Remove applications that should have been destroyed
	sort.Sort(sort.Reverse(sort.IntSlice(toRemove)))
	for idx := range toRemove {
		Client.Applications = append(Client.Applications[:idx], Client.Applications[idx+1:]...)
	}

	for _, runningApp := range Client.Applications {
		for idx, definedApp := range definedApplications {
			if runningApp.ID == definedApp.ID {
				if !reflect.DeepEqual(runningApp.Env, definedApp.Env) || runningApp.Image != definedApp.Image || !reflect.DeepEqual(runningApp.Ports, definedApp.Ports) || runningApp.CPU != definedApp.CPU || runningApp.Memory != definedApp.Memory {
					fmt.Printf("Performing application update for %s\n", runningApp.Name)
					destroyApplication(runningApp)
					deployApplication(definedApp, false)
					Client.Applications[idx].Env = definedApp.Env
					Client.Applications[idx].Image = definedApp.Image
					Client.Applications[idx].Ports = definedApp.Ports
					Client.Applications[idx].CPU = definedApp.CPU
					Client.Applications[idx].Memory = definedApp.Memory
				}
			}
		}
	}

	// Update /etc/hosts for running applications
	for _, definedApp := range definedApplications {
		if Client.ID == definedApp.ID {
			err = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s exec -u 0 %s sh -c \"echo \\\"$(cat /var/stormfront/%s.hosts)\\n$(cat /var/stormfront/hosts)\\\" > /etc/hosts\"", config.Config.ContainerEngine, definedApp.Name, definedApp.Name)).Run()
			if err != nil {
				fmt.Printf("Encountered error copying to hosts file: %v\n", err.Error())
				continue
			}
		}
	}
}
