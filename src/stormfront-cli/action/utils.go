package action

import (
	"encoding/json"
	"fmt"
	"stormfront-cli/config"
	"stormfront-cli/utils"
	"strings"
)

func ParseJSON(jsonBody string) ([]map[string]interface{}, error) {
	var output []map[string]interface{}

	if strings.HasPrefix(jsonBody, "[") {
		if err := json.Unmarshal([]byte(jsonBody), &output); err != nil {
			return []map[string]interface{}{}, err
		}
	} else {
		if err := json.Unmarshal([]byte(fmt.Sprintf("[%s]", jsonBody)), &output); err != nil {
			return []map[string]interface{}{}, err
		}
	}

	return output, nil
}

func FilterNamespace(data []map[string]interface{}, namespace string) ([]map[string]interface{}, error) {
	var err error
	if namespace != "all" {
		if namespace == "" {
			namespace, err = config.GetNamespace()
			if err != nil {
				return []map[string]interface{}{}, err
			}
		}
		data = utils.Filter(data, "namespace", namespace)
	}
	return data, nil
}
