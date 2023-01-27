package destroy

import (
	"fmt"
	"os"
	"stormfront-cli/destroy/application"
	"stormfront-cli/destroy/client"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var DestroyHelpText = fmt.Sprintf(`usage: stormfront application <command> [-l|--log-level <log level>] [-h|--help]
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

func ParseDestroyArgs(args []string) {
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
			fmt.Println(DestroyHelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(DestroyHelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "application", "app":
		host, port, id, err := application.ParseApplicationArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DestroyHelpText)
			os.Exit(1)
		}
		err = application.ExecuteApplication(host, port, id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "client", "cl":
		host, port, err := client.ParseClientArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DestroyHelpText)
			os.Exit(1)
		}
		err = client.ExecuteClient(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid argument: %s\n", args[1])
		fmt.Println(DestroyHelpText)
		os.Exit(1)
	}

}
