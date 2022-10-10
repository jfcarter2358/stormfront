package daemon

import (
	"fmt"
	"os"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var DaemonHelpText = fmt.Sprintf(`usage: stormfront daemon <command> [-l|--log-level <log level>] [-h|--help]
commands:
	health            Get the health of the stormfront daemon
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func ParseDaemonArgs(args []string) {
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	if len(args) > 1 {
		if args[1] == "-l" || args[1] == "--log-level" {
			if len(args) == 2 {
				logging.Fatal("No value passed after log-level flag")
			}
			err := logging.SetLevel(args[2])
			if err != nil {
				logging.Fatal(err.Error())
			}
			args = append(args[:0], args[2:]...)
		}
	}

	if len(args) == 2 {
		if utils.Contains(args, "-h") || utils.Contains(args, "--help") {
			fmt.Println(DaemonHelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(DaemonHelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "health":
		host, port, err := ParseHealthArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DaemonHelpText)
			os.Exit(1)
		}
		err = ExecuteHealth(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid argument: %s\n", args[1])
		fmt.Println(DaemonHelpText)
		os.Exit(1)
	}

}
