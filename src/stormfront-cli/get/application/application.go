package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"stormfront-cli/action"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
	"strings"

	"gopkg.in/yaml.v2"
)

var ApplicationHelpText = fmt.Sprintf(`usage: stormfront application get [<application id>] [-o|--output <output>] [-n|--namespace] [-a|--all-namespaces] [-l|--log-level <log level>] [-h|--help]
arguments:
	-o|--output            Output format to print to console, valid options are "table", "yaml", and "json"
	-n|--namespace         Namespace to grab applications from
	-a|--all-namespaces    Show applications from all namespaces
	-l|--log-level         Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help              Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseApplicationArgs(args []string) (string, string, string, error) {
	id := ""
	output := "table"
	namespace := ""
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
					return "", "", "", fmt.Errorf("invalid output value %s, allowed values are 'table', 'yaml', and 'json", args[1])
				}
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after output flag")
			}
		case "-a", "--all-namespaces":
			namespace = "all"
			args = args[1:]
		case "-n", "--namespace":
			if len(args) > 1 {
				namespace = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after namespace flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", "", err
				}
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after log-level flag")
			}
		default:
			if strings.HasPrefix(args[0], "-") || id != "" {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(ApplicationHelpText)
				os.Exit(1)
			} else {
				id = args[0]
				args = args[1:]
			}
		}
	}

	return id, output, namespace, nil
}

func ExecuteApplication(id, output, namespace string) error {
	var applications []map[string]interface{}
	var err error
	if namespace == "" {
		namespace, err = config.GetNamespace()
		if err != nil {
			return err
		}
	}

	if id == "" {
		applications, err = action.GetAllApplications(namespace)
		if err != nil {
			return err
		}
	} else {
		applications, err = action.GetApplicationByNameNamespace(id, namespace)
		if err != nil {
			applications, err = action.GetApplicationById(id)
			if err != nil {
				return err
			}
		}
	}

	for idx, app := range applications {
		applications[idx]["state"] = app["status"].(map[string]interface{})["status"].(string)
	}

	headers := []string{
		"id",
		"name",
		"image",
		"node",
		"hostname",
		"namespace",
		"state",
	}
	types := []string{
		"string",
		"string",
		"string",
		"string",
		"string",
		"string",
		"string",
	}

	switch output {
	case "table":
		utils.PrintTable(applications, headers, types)
	case "yaml":
		contents, _ := yaml.Marshal(&applications)
		fmt.Println(string(contents))
	case "json":
		contents, _ := json.Marshal(&applications)
		fmt.Println(string(contents))
	}
	logging.Success("Done!")

	return nil
}
