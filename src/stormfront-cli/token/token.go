package token

import (
	"fmt"
	"os"
	"stormfront-cli/logging"
	"stormfront-cli/token/api"
	"stormfront-cli/token/join"
	"stormfront-cli/utils"
)

var TokenHelpText = fmt.Sprintf(`usage: stormfront token <command> [-l|--log-level <log level>] [-h|--help]
commands:
	api               Manage api tokens for this Stormfront cluster
	join              Manage join tokens for this Stormfront cluster
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseTokenArgs(args []string) {
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping\n", envLogLevel)
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
			fmt.Println(TokenHelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(TokenHelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "api":
		api.ParseAPIArgs(args[1:])
	case "join":
		join.ParseJoinArgs(args[1:])
	default:
		fmt.Printf("Invalid argument: %s\n", args[1])
		fmt.Println(TokenHelpText)
		os.Exit(1)
	}

}
