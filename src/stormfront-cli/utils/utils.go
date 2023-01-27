// utils.go

package utils

import (
	"fmt"
	"strings"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func typeToString(value interface{}, valueType string) string {
	switch valueType {
	case "string":
		return value.(string)
	case "int":
		return fmt.Sprintf("%v", value.(int))
	case "float":
		return fmt.Sprintf("%v", value.(float32))
	case "bool":
		return fmt.Sprintf("%v", value.(bool))
	}

	return ""
}

func PrintTable(data []map[string]interface{}, headers, types []string) {
	widths := []int{}
	output := ""
	border := "="
	spacer := "  |  "

	// Figure out spacing
	for _, header := range headers {
		widths = append(widths, len(header))
	}
	for _, datum := range data {
		for idx, header := range headers {
			if len(typeToString(datum[header], types[idx])) > widths[idx] {
				widths[idx] = len(datum[header].(string))
			}
		}
	}

	// Print out headers
	for idx, header := range headers {
		delta := widths[idx] - len(header)
		if idx < len(headers)-1 {
			output += header + strings.Repeat(" ", delta) + spacer
		} else {
			output += header + strings.Repeat(" ", delta)
		}
	}
	output += "\n"

	// Print out border line
	for idx, width := range widths {
		if idx < len(widths)-1 {
			output += strings.Repeat(border, width+len(spacer))
		} else {
			output += strings.Repeat(border, width)
		}
	}
	output += "\n"

	// Print out data
	for _, datum := range data {
		for idx, header := range headers {
			delta := widths[idx] - len(datum[header].(string))
			if idx < len(headers)-1 {
				output += typeToString(datum[header], types[idx]) + strings.Repeat(" ", delta) + spacer
			} else {
				output += typeToString(datum[header], types[idx]) + strings.Repeat(" ", delta)
			}
		}
		output += "\n"
	}

	fmt.Println(output)
}
