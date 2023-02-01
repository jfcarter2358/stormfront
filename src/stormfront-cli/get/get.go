package get

import (
	"fmt"
	"os"
	"stormfront-cli/get/application"
	"stormfront-cli/get/client"
	"stormfront-cli/get/cluster"
	"stormfront-cli/get/namespace"
	"stormfront-cli/get/node"
	"stormfront-cli/get/route"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
)

var GetHelpText = fmt.Sprintf(`usage: stormfront get <command> [-l|--log-level <log level>] [-h|--help]
commands:
	application       Get information about running applications
	client            Get information about running clients
	cluster           Get information about available clusters
	namespace         Get information about namespaces in current cluster
	node              Get information about running nodes
	route             Get information about defined routes
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseGetArgs(args []string) {
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
			fmt.Println(GetHelpText)
			os.Exit(0)
		}
	}

	if len(args) == 1 {
		fmt.Println(GetHelpText)
		os.Exit(1)
	}

	switch args[1] {
	case "application", "app":
		id, output, namespace, err := application.ParseApplicationArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(GetHelpText)
			os.Exit(1)
		}
		err = application.ExecuteApplication(id, output, namespace)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "cluster":
		output, err := cluster.ParseClusterArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(GetHelpText)
			os.Exit(1)
		}
		err = cluster.ExecuteCluster(output)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "route", "rt":
		id, output, namespace, err := route.ParseRouteArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(GetHelpText)
			os.Exit(1)
		}
		err = route.ExecuteRoute(id, output, namespace)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "namespace", "ns":
		output, err := namespace.ParseNamespaceArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(GetHelpText)
			os.Exit(1)
		}
		err = namespace.ExecuteNamespace(output)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "client", "cl":
		id, output, err := client.ParseClientArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(GetHelpText)
			os.Exit(1)
		}
		err = client.ExecuteClient(id, output)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	case "node", "no":
		id, output, err := node.ParseNodeArgs(args[2:])
		if err != nil {
			logging.Error(err.Error())
			fmt.Println(GetHelpText)
			os.Exit(1)
		}
		err = node.ExecuteNode(id, output)
		if err != nil {
			logging.Error(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid argument: %s\n", args[1])
		fmt.Println(GetHelpText)
		os.Exit(1)
	}

}
