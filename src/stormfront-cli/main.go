package main

import (
	"fmt"
	"os"
	"stormfront-cli/client"
	"stormfront-cli/daemon"
	"stormfront-cli/debug"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var HelpText = fmt.Sprintf(`usage: stormfront <command> [-l|--log-level <log level>] [-h|--help]
commands:
	daemon            Interact with the stormfront daemon
	client            Interact with a running client
	debug             Execute debugging actions against a running client
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func main() {

	args := os.Args

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
			fmt.Println(HelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(HelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "daemon":
		daemon.ParseDaemonArgs(args[1:])
	case "client":
		client.ParseClientArgs(args[1:])
	case "debug":
		debug.ParseDebugArgs(args[1:])
	default:
		logging.Error(fmt.Sprintf("Invalid argument: %s\n", args[1]))
		fmt.Println(HelpText)
		os.Exit(1)
	}
}
