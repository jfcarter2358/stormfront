package main

import (
	"fmt"
	"os"
	"stormfront-cli/deploy"
	"stormfront-cli/destroy"
	"stormfront-cli/get"
	"stormfront-cli/join"
	"stormfront-cli/logging"
	"stormfront-cli/logs"
	"stormfront-cli/token"
	"stormfront-cli/utils"
)

var HelpText = fmt.Sprintf(`usage: stormfront <command> [-l|--log-level <log level>] [-h|--help]
commands:
	api-token           Manage cluster API tokens
	app                 Manage applications deployed to Stormfront
	client              Interact with a running client
	daemon              Interact with the stormfront daemon
	debug               Execute debugging actions against a running client
	down                Destroy a running stormfront client
	get-join-command    Generate a join command to add a client as a follower
	join                Deploy a follower client that joins a leader at a specified location
	restart             Restart a running stormfront client
	up                  Start up a stormfront client
arguments:
	-l|--log-level      Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help           Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

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
	case "deploy":
		deploy.ParseDeployArgs(args[1:])
	case "destroy":
		destroy.ParseDestroyArgs(args[1:])
	case "get":
		get.ParseGetArgs(args[1:])
	case "join":
		host, port, leader, joinToken, err := join.ParseJoinArgs(args[1:])
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
		host, port, id, err := logs.ParseLogsArgs(args[1:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = logs.ExecuteLogs(host, port, id)
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
