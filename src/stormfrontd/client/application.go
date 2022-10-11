package client

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
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

func getApplicationStatus(apps []StormfrontApplication) []StormfrontApplication {
	cmd := exec.Command("/bin/sh", "-c", "docker stats --no-stream --no-trunc --all --format \"{{.Name}}||{{.CPUPerc}}||{{.MemPerc}}\"")
	var outb1 bytes.Buffer
	cmd.Stdout = &outb1
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error getting container stats: %v\n", err.Error())
	} else {
		lines := strings.Split(outb1.String(), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "||")
			if len(parts) == 3 {
				for idx, app := range apps {
					if app.Name == parts[0] {
						app.Status.CPU = parts[1]
						app.Status.Memory = parts[2]
						apps[idx] = app
					}
					break
				}
			} else {
				break
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
		for _, line := range lines {
			parts := strings.Split(line, "||")
			if len(parts) == 2 {
				for idx, app := range apps {
					if app.Name == parts[0] {
						app.Status.Status = strings.Split(parts[1], " ")[0]
						apps[idx] = app
					}
					break
				}
			} else {
				break
			}
		}
	}

	return apps
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

	dockerCommand := "docker run -d --rm "
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

func reconcileApplications(updatePackage StormfrontUpdatePackage) {
	hostContents := ""
	for _, definedApp := range updatePackage.Applications {
		hostContents += fmt.Sprintf("%s %s\n", Client.Host, definedApp.Hostname)
	}

	err := os.WriteFile("/var/stormfront/hosts", []byte(hostContents), 0644)
	if err != nil {
		panic(err)
	}

	// Check for applications that should be deployed
	for _, definedApp := range updatePackage.Applications {
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
		for _, definedApp := range updatePackage.Applications {
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
		for _, definedApp := range updatePackage.Applications {
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
	for _, definedApp := range updatePackage.Applications {
		err = exec.Command("/bin/sh", "-c", fmt.Sprintf("docker exec -u 0 %s sh -c \"echo \\\"$(cat /var/stormfront/%s.hosts)\\n$(cat /var/stormfront/hosts)\\\" > /etc/hosts\"", definedApp.Name, definedApp.Name)).Run()
		if err != nil {
			fmt.Printf("Encountered error copying to hosts file: %v\n", err.Error())
			continue
		}
	}
}
