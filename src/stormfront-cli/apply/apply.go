package apply

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
	"strings"

	"gopkg.in/yaml.v3"
)

var ApplyCreateHelpText = fmt.Sprintf(`usage: stormfront apply -f|--file <object definition file> [-n|--namespace <namespace>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-f|--file         The path to the JSON file defining the apply
	-n|--namespace    Namespace to deploy the apply to
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseApplyArgs(args []string) (string, string, error) {
	definition := ""
	namespace := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping\n", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-f", "--file":
			if len(args) > 1 {
				definition = args[1]
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after file flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", err
				}
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after log-level flag")
			}
		case "-n", "--namespace":
			if len(args) > 1 {
				namespace = args[1]
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after port flag")
			}
		default:
			fmt.Printf("Invalid argument: %s\n", args[0])
			fmt.Println(ApplyCreateHelpText)
			os.Exit(1)
		}
	}

	if definition == "" {
		return "", "", errors.New("missing object definition file")
	}

	return definition, namespace, nil
}

func ExecuteApply(definition, namespace string) error {
	host, err := config.GetHost()
	if err != nil {
		return err
	}
	port, err := config.GetPort()
	if err != nil {
		return err
	}

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return err
	}

	var data []map[string]interface{}
	extension := filepath.Ext(definition)
	switch extension {
	case ".yaml", ".yml":
		data = parseYAML(definition)
	case ".json":
		data = parseJSON(definition)
	default:
		logging.Error("Unsupported apply file format, must either be .json, .yaml, or .yml")
	}

	fmt.Printf("%v\n", data)

	for _, datum := range data {
		name := datum["name"].(string)
		if _, ok := datum["kind"]; !ok {
			return fmt.Errorf("object %s is missing 'kind' field", name)
		}

		kind := datum["kind"].(string)

		switch kind {
		case "namespace":
			if err := createNamespace(name); err != nil {
				return err
			}
		case "application":
			if err := createApplication(host, port, namespace, apiToken, datum); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid object type of '%s', allowed types are 'namespace' and 'application'", kind)
		}
	}

	logging.Success("All objects created")

	return nil
}

func createNamespace(name string) error {
	logging.Info("Creating namespace...")
	err := config.AddNamespace(name)

	logging.Success("Done!")

	return err
}

func createApplication(host, port, namespace, apiToken string, datum map[string]interface{}) error {
	logging.Info("Creating application...")
	requestURL := fmt.Sprintf("http://%s:%s/api/application", host, port)

	logging.Debug("Sending POST request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	if _, ok := datum["namespace"]; ok {
		if namespace != "" {
			namespaces, err := config.GetNamespaces()
			if err != nil {
				return err
			}
			if !utils.Contains(namespaces, namespace) {
				return fmt.Errorf("application trying to be deployed to non-existent namespace: '%s'", namespace)
			}
			datum["namespace"] = namespace
		}
	} else {
		if namespace != "" {
			namespaces, err := config.GetNamespaces()
			if err != nil {
				return err
			}
			if !utils.Contains(namespaces, namespace) {
				return fmt.Errorf("application trying to be deployed to non-existent namespace: '%s'", namespace)
			}
			datum["namespace"] = namespace
		} else {
			configNamespace, err := config.GetNamespace()
			if err != nil {
				return err
			}
			datum["namespace"] = configNamespace
		}
	}

	postBody, _ := json.Marshal(datum)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		logging.Success("Done!")
	} else {
		var response_data map[string]string
		if err := json.Unmarshal([]byte(responseBody), &response_data); err == nil {
			if errMessage, ok := response_data["error"]; ok {
				logging.Error(errMessage)
			}
		}
		logging.Fatal(fmt.Sprintf("Client has returned error with status code %v", resp.StatusCode))
	}
	logging.Info("Done!")

	return nil
}

func parseJSON(definition string) []map[string]interface{} {
	file, _ := ioutil.ReadFile(definition)
	contents := string(file)

	if strings.HasPrefix(contents, "{") {
		contents = fmt.Sprintf("[%s]", contents)
	}
	data := []map[string]interface{}{}

	err := json.Unmarshal([]byte(contents), &data)
	if err != nil {
		panic(err)
	}

	return data
}

func parseYAML(definition string) []map[string]interface{} {
	file, err := ioutil.ReadFile(definition)
	if err != nil {
		panic(err)
	}

	r := bytes.NewReader(file)
	dec := yaml.NewDecoder(r)

	// var data []map[string]interface{}
	// var document map[string]interface{}
	// for dec.Decode(&document) == nil {
	// 	data = append(data, document)
	// }

	var data []map[string]interface{}
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		content, err := yaml.Marshal(&node)
		if err != nil {
			panic(err)
		}

		document := map[string]interface{}{}
		yaml.Unmarshal(content, &document)
		data = append(data, document)
	}

	return data
}
