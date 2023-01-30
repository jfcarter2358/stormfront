package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"stormfront-cli/utils"

	"gopkg.in/yaml.v2"
)

var ClusterHelpText = fmt.Sprintf(`usage: stormfront get cluster [-o|--output <output>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-o|--output            Output format to print to console, valid options are "table", "yaml", and "json"
	-l|--log-level         Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help              Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseClusterArgs(args []string) (string, error) {
	output := "table"
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-o", "--output":
			if len(args) > 1 {
				switch args[1] {
				case "table", "yaml", "json":
					output = args[1]
				default:
					return "", fmt.Errorf("invalid output value %s, allowed values are 'table', 'yaml', and 'json", args[1])
				}
				args = args[2:]
			} else {
				return "", errors.New("no value passed after output flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", err
				}
				args = args[2:]
			} else {
				return "", errors.New("no value passed after log-level flag")
			}
		default:
			fmt.Printf("Invalid argument: %s\n", args[0])
			fmt.Println(ClusterHelpText)
			os.Exit(1)
		}
	}

	return output, nil
}

func ExecuteCluster(output string) error {
	logging.Info("Getting clusters...")

	clusters, err := config.GetClusters()
	if err != nil {
		return err
	}
	currentCluster, err := config.GetCluster()
	if err != nil {
		return err
	}

	clusterData := []map[string]interface{}{}

	for _, cluster := range clusters {
		datum := map[string]interface{}{}
		datum["name"] = cluster.Name
		datum["namespace"] = cluster.CurrentNamespace
		datum["host"] = cluster.Host
		datum["port"] = cluster.Port
		if cluster.Name == currentCluster {
			datum["selected"] = "*"
		} else {
			datum["selected"] = ""
		}
		clusterData = append(clusterData, datum)
	}

	headers := []string{
		"name",
		"namespace",
		"host",
		"port",
		"selected",
	}
	types := []string{
		"string",
		"string",
		"string",
		"string",
		"string",
	}

	switch output {
	case "table":
		utils.PrintTable(clusterData, headers, types)
	case "yaml":
		contents, _ := yaml.Marshal(&clusterData)
		fmt.Println(string(contents))
	case "json":
		contents, _ := json.Marshal(&clusterData)
		fmt.Println(string(contents))
	}
	logging.Success("Done!")

	return nil
}
