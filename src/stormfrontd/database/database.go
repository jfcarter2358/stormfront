package database

import (
	"fmt"
	"os"
	"os/exec"
	"stormfrontd/config"
)

const CERESDB_IMAGE = "jfcarter2358/ceresdb:1.1.0"

func Deploy(leader string) error {
	os.MkdirAll("/var/stormfront/ceresdb/data", os.ModePerm)
	os.MkdirAll("/var/stormfront/ceresdb/indices", os.ModePerm)
	fmt.Println("Removing any existing ceresdb containers...")
	exec.Command("/bin/sh", "-c", fmt.Sprintf("%s kill ceresdb || true; %s rm ceresdb || true", config.Config.ContainerEngine, config.Config.ContainerEngine)).Run()
	fmt.Println("Deploying CeresDB...")
	dockerCommand := fmt.Sprintf("%s run --net host -d --rm ", config.Config.ContainerEngine)
	dockerCommand += "--name ceresdb "
	dockerCommand += fmt.Sprintf("-e CERESDB_DEFAULT_ADMIN_PASSWORD=%s ", config.Config.CeresDBPassword)
	if leader != "" {
		dockerCommand += fmt.Sprintf("-e CERESDB_LEADER='%s' ", leader)
		dockerCommand += fmt.Sprintf("-e CERESDB_FOLLOWER_AUTH='ceresdb:%s' ", config.Config.CeresDBPassword)
	}
	dockerCommand += "-p 7437:7437 "
	// dockerCommand += "-v /var/stormfront/ceresdb/data:/home/ceresdb/.ceresdb/data "
	// dockerCommand += "-v /var/stormfront/ceresdb/indices:/home/ceresdb/.ceresdb/indices "
	dockerCommand += CERESDB_IMAGE
	err := exec.Command("/bin/sh", "-c", dockerCommand).Run()
	if err != nil {
		fmt.Printf("Encountered error deploying CeresDB instance: %v\n", err.Error())
		return err
	}
	return nil
}

func Destroy() error {
	err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s kill ceresdb", config.Config.ContainerEngine)).Run()
	if err != nil {
		fmt.Printf("Encountered error killing container: %v\n", err.Error())
		return err
	}
	err = os.RemoveAll("/var/stormfront/ceresdb")
	if err != nil {
		fmt.Printf("Encountered error removing CeresDB data: %v\n", err.Error())
		return err
	}
	return nil
}
