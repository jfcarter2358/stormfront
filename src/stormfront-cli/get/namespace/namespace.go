package namespace

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

var NamespaceHelpText = fmt.Sprintf(`usage: stormfront get namespace [-o|--output <output>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-o|--output       Output format to print to console, valid options are "table", "yaml", and "json"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseNamespaceArgs(args []string) (string, error) {
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
			fmt.Println(NamespaceHelpText)
			os.Exit(1)
		}
	}

	return output, nil
}

func ExecuteNamespace(output string) error {
	logging.Info("Getting namespaces...")

	namespaces, err := config.GetNamespaces()
	if err != nil {
		return err
	}
	currentNamespace, err := config.GetNamespace()
	if err != nil {
		return err
	}

	namespaceData := []map[string]interface{}{}

	for _, namespace := range namespaces {
		datum := map[string]interface{}{}
		datum["name"] = namespace
		if namespace == currentNamespace {
			datum["selected"] = "*"
		} else {
			datum["selected"] = ""
		}
		namespaceData = append(namespaceData, datum)
	}
	headers := []string{
		"name",
		"selected",
	}
	types := []string{
		"string",
		"string",
	}

	switch output {
	case "table":
		utils.PrintTable(namespaceData, headers, types)
	case "yaml":
		contents, _ := yaml.Marshal(&namespaceData)
		fmt.Println(string(contents))
	case "json":
		contents, _ := json.Marshal(&namespaceData)
		fmt.Println(string(contents))
	}
	logging.Success("Done!")

	return nil
}
