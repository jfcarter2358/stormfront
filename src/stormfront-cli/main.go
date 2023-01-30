package main

import (
	"fmt"
	"os"
	"stormfront-cli/apply"
	"stormfront-cli/create"
	"stormfront-cli/delete"
	"stormfront-cli/edit"
	"stormfront-cli/get"
	"stormfront-cli/join"
	"stormfront-cli/logging"
	"stormfront-cli/logs"
	"stormfront-cli/token"
	"stormfront-cli/utils"
)

var HelpText = fmt.Sprintf(`usage: stormfront <command> [-l|--log-level <log level>] [-h|--help]
commands:
	apply            Apply an object definition file
	create           Create a Stormfront client
	delete           Delete Stormfront objects
	edit             Change cluster or namespace in your ~/.stormfrontconfig file
	get              Get Stormfront cluster objects
	join             Join an existing Stormfront cluster
	logs             Get logs for a running application
	restart          Restart a running client or application
	token            Manage cluster access, API, and join tokens
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func main() {

	args := os.Args

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
			fmt.Println(HelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(HelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "apply":
		definition, namespace, err := apply.ParseApplyArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = apply.ExecuteApply(definition, namespace)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "create":
		create.ParseCreateArgs(args[1:])
	case "delete":
		delete.ParseDeleteArgs(args[1:])
	case "get":
		get.ParseGetArgs(args[1:])
	case "edit":
		edit.ParseEditArgs(args[1:])
	case "join":
		host, port, leader, joinToken, err := join.ParseJoinArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = join.ExecuteJoin(host, port, leader, joinToken)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "logs":
		id, err := logs.ParseLogsArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = logs.ExecuteLogs(id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "restart":
		get.ParseGetArgs(args[1:])
	case "token":
		token.ParseTokenArgs(args[1:])
	// case "update":
	// 	update.ParseUpdateArgs(args[1:])
	default:
		logging.Error(fmt.Sprintf("Invalid argument: %s\n", args[1]))
		fmt.Println(HelpText)
		os.Exit(1)
	}
}
