package application

import (
	"fmt"
	"os"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var ApplicationHelpText = fmt.Sprintf(`usage: stormfront application <command> [-l|--log-level <log level>] [-h|--help]
commands:
	create            Create a new application
	delete            Delete a running application
	get-all           Get all running applications
	get               Get a running application
	logs              Get the logs for a running application
	update            Update a running application
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseApplicationArgs(args []string) {
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
			fmt.Println(ApplicationHelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(ApplicationHelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "get-all":
		host, port, err := ParseGetAllArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(ApplicationHelpText)
			os.Exit(1)
		}
		err = ExecuteGetAll(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "get":
		host, port, id, err := ParseGetArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(ApplicationHelpText)
			os.Exit(1)
		}
		err = ExecuteGet(host, port, id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "create":
		host, port, definition, err := ParseCreateArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(ApplicationHelpText)
			os.Exit(1)
		}
		err = ExecuteCreate(host, port, definition)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "delete":
		host, port, id, err := ParseDeleteArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(ApplicationHelpText)
			os.Exit(1)
		}
		err = ExecuteDelete(host, port, id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "logs":
		host, port, id, err := ParseLogsArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(ApplicationHelpText)
			os.Exit(1)
		}
		err = ExecuteLogs(host, port, id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "update":
		host, port, definition, id, err := ParseUpdateArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(ApplicationHelpText)
			os.Exit(1)
		}
		err = ExecuteUpdate(host, port, definition, id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid argument: %s\n", args[1])
		fmt.Println(ApplicationHelpText)
		os.Exit(1)
	}

}
