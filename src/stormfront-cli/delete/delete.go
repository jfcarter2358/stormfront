package delete

import (
	"fmt"
	"os"
	"stormfront-cli/delete/application"
	"stormfront-cli/delete/client"
	"stormfront-cli/delete/cluster"
	"stormfront-cli/delete/namespace"
	"stormfront-cli/delete/route"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var DeleteHelpText = fmt.Sprintf(`usage: stormfront delete <object> [-l|--log-level <log level>] [-h|--help]
commands:
	application       Delete an existing application
	client            Delete an existing client
	cluster           Delete a cluster from your .stormfrontconfig file
	route             Delete an existing route
	namespace         Delete a namespace from an existing cluster
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseDeleteArgs(args []string) {
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
			fmt.Println(DeleteHelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(DeleteHelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "application", "app":
		id, namespace, err := application.ParseApplicationArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DeleteHelpText)
			os.Exit(1)
		}
		err = application.ExecuteApplication(id, namespace)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "client", "cl":
		host, port, err := client.ParseClientArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DeleteHelpText)
			os.Exit(1)
		}
		err = client.ExecuteClient(host, port)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "cluster":
		name, err := cluster.ParseClusterArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DeleteHelpText)
			os.Exit(1)
		}
		err = cluster.ExecuteCluster(name)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "namespace", "ns":
		id, err := namespace.ParseNamespaceArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DeleteHelpText)
			os.Exit(1)
		}
		err = namespace.ExecuteNamespace(id)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "route", "rt":
		id, namespace, err := route.ParseRouteArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(DeleteHelpText)
			os.Exit(1)
		}
		err = route.ExecuteRoute(id, namespace)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid argument: %s\n", args[1])
		fmt.Println(DeleteHelpText)
		os.Exit(1)
	}

}
