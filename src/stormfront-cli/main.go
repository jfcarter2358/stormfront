package main

import (
	"fmt"
	"os"
	"stormfront-cli/application"
	"stormfront-cli/client"
	"stormfront-cli/daemon"
	"stormfront-cli/debug"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var HelpText = fmt.Sprintf(`usage: stormfront <command> [-l|--log-level <log level>] [-h|--help]
commands:
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
	-h|--help           Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

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
	case "app":
		application.ParseApplicationArgs(args[1:])
	case "daemon":
		daemon.ParseDaemonArgs(args[1:])
	case "client":
		client.ParseClientArgs(args[1:])
	case "debug", "db":
		debug.ParseDebugArgs(args[1:])
	case "up":
		host, port, err := ParseUpArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = ExecuteUp(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "down":
		host, port, err := ParseDownArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = ExecuteDown(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "get-join-command":
		host, port, err := ParseGetJoinCommandArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = ExecuteGetJoinCommand(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "join":
		host, port, leader, joinToken, err := ParseJoinArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = ExecuteJoin(host, port, leader, joinToken)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "restart":
		host, port, err := ParseRestartArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(HelpText)
			os.Exit(1)
		}
		err = ExecuteRestart(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	default:
		logging.Error(fmt.Sprintf("Invalid argument: %s\n", args[1]))
		fmt.Println(HelpText)
		os.Exit(1)
	}
}
