package cluster

import (
	"errors"
	"fmt"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"strings"
)

var ClusterHelpText = fmt.Sprintf(`usage: stormfront edit cluster <old cluster name|'current' to change targeted cluster> <new cluster name> [-l|--log-level <log level>] [-h|--help]
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseClusterArgs(args []string) (string, string, error) {
	oldName := ""
	newName := ""
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
		default:
			if strings.HasPrefix(args[0], "-") || (oldName != "" && newName != "") {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(ClusterHelpText)
				os.Exit(1)
			} else {
				if oldName == "" {
					oldName = args[0]
				} else {
					newName = args[0]
				}
				args = args[1:]
			}
		}
	}

	if oldName == "" {
		return "", "", errors.New("old name argument is required")
	}
	if newName == "" {
		return "", "", errors.New("new name argument is required")
	}

	return oldName, newName, nil
}

func ExecuteCluster(oldName, newName string) error {

	if oldName == "current" {
		err := config.ChangeCluster(newName)
		logging.Info("Done!")
		return err
	}

	logging.Info("Editing cluster...")

	err := config.RenameCluster(oldName, newName)

	logging.Info("Done!")

	return err
}
