package logs

import (
	"errors"
	"fmt"
	"os"
	"stormfront-cli/action"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"strings"
)

var LogsHelpText = fmt.Sprintf(`usage: stormfront logs <application id> [-l|--log-level <log level>] [-h|--help]
arguments:
	-i|--id           The ID of the application to get
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseLogsArgs(args []string) (string, string, error) {
	id := ""
	namespace := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", err
				}
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after log-level flag")
			}
		case "-n", "--namespace":
			if len(args) > 1 {
				namespace = args[1]
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after namespace flag")
			}
		default:
			if strings.HasPrefix(args[0], "-") || id != "" {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(LogsHelpText)
				os.Exit(1)
			} else {
				id = args[0]
				args = args[1:]
			}
		}
	}

	return id, namespace, nil
}

func ExecuteLogs(id, namespace string) error {
	var output string
	var err error
	if namespace == "" {
		namespace, err = config.GetNamespace()
		if err != nil {
			return err
		}
	}

	output, err = action.GetLogsByNameNamespace(id, namespace)
	if err != nil {
		output, err = action.GetLogsById(id)
		if err != nil {
			return err
		}
	}

	fmt.Println(output)

	return nil
}
