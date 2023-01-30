package namespace

import (
	"errors"
	"fmt"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"strings"
)

var NamespaceHelpText = fmt.Sprintf(`usage: stormfront edit namespace <new namespace name> [-l|--log-level <log level>] [-h|--help]
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseNamespaceArgs(args []string) (string, error) {
	name := ""
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
					return "", err
				}
				args = args[2:]
			} else {
				return "", errors.New("no value passed after log-level flag")
			}
		default:
			if strings.HasPrefix(args[0], "-") || name != "" {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(NamespaceHelpText)
				os.Exit(1)
			} else {
				name = args[0]
				args = args[1:]
			}
		}
	}

	if name == "" {
		return "", errors.New("new name argument is required")
	}

	return name, nil
}

func ExecuteNamespace(name string) error {
	logging.Info("Editing cluster...")

	err := config.ChangeNamespace(name)

	logging.Info("Done!")

	return err
}
