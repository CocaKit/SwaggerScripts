package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sort"
)

func json2swagger() {
	var inputData []byte
	var err error

	inputData, err = io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var jsonData interface{}
	if err := json.Unmarshal(inputData, &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid JSON input: %v\n", err)
		os.Exit(1)
	}

	exampleOutput, schemaOutput := formatNodes(jsonData, 2, true)

	fmt.Printf("ExampleName:\n")
	fmt.Println("  value:")
	fmt.Print(exampleOutput)

	fmt.Println()

	fmt.Printf("SchemaName:\n")
	fmt.Print(schemaOutput)
}

func formatNodes(v interface{}, indent int, isRoot bool) (string, string) {
	baseIndent := strings.Repeat("  ", indent)
	innerIndent := strings.Repeat("  ", indent+1)

	switch val := v.(type) {
	case map[string]interface{}:
		if len(val) == 0 {
			if isRoot {
				return "", ""
			}
			return "{}", ""
		}

		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var exampleNodes strings.Builder
		if !isRoot {
			exampleNodes.WriteString("{\n")
		}

		var schemaNodes strings.Builder
		
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: object\n")
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("properties:\n")

		for i, k := range keys {
			exampleNodes.WriteString(innerIndent)
			exampleNodes.WriteString(k)
			exampleNodes.WriteString(": ")

			schemaNodes.WriteString(innerIndent)
			schemaNodes.WriteString(k)
			schemaNodes.WriteString(":\n")

			exampleNode, schemaNode := formatNodes(val[k], indent+2, false)

			exampleNodes.WriteString(exampleNode)
			schemaNodes.WriteString(schemaNode)

			if i < len(keys)-1 {
				exampleNodes.WriteString(",")
			}
			exampleNodes.WriteString("\n")
		}

		if !isRoot {
			exampleNodes.WriteString(baseIndent)
			exampleNodes.WriteString("}")
		}

		return exampleNodes.String(), schemaNodes.String()
	case string:
		exampleStr := fmt.Sprintf("%q", val)

		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: string\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("example: ")
		schemaNodes.WriteString(exampleStr)
		schemaNodes.WriteString("\n")

		return exampleStr, schemaNodes.String()
	case float64:
		var exampleStr string
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)

		if val == float64(int64(val)) {
			exampleStr = fmt.Sprintf("%d", int64(val))
			schemaNodes.WriteString("type: integer\n")
		} else {
			exampleStr = fmt.Sprintf("%g", val)
			schemaNodes.WriteString("type: float\n")
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("example: ")
		schemaNodes.WriteString(exampleStr)
		schemaNodes.WriteString("\n")

		return exampleStr, schemaNodes.String()
	case bool:
		exampleStr := fmt.Sprintf("%t", val)

		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: bool\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("example: ")
		schemaNodes.WriteString(exampleStr)
		schemaNodes.WriteString("\n")

		return exampleStr, schemaNodes.String()
	case nil:
		exampleStr := "null"

		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("example: ")
		schemaNodes.WriteString(exampleStr)
		schemaNodes.WriteString("\n")

		return exampleStr, schemaNodes.String()
	case []interface{}:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: array\n")
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("items:\n")

		if len(val) == 0 {
			return "[]", schemaNodes.String() 
		}

		exampleNode, schemaNode := formatNodes(val[0], indent+1, false)

		schemaNodes.WriteString(schemaNode)

		var exampleNodes strings.Builder

		exampleNodes.WriteString("[\n")
		exampleNodes.WriteString(innerIndent)
		exampleNodes.WriteString(exampleNode)
		exampleNodes.WriteString("\n")
		exampleNodes.WriteString(baseIndent)
		exampleNodes.WriteString("]")

		return exampleNodes.String(), schemaNodes.String()
	default:
		exampleStr := fmt.Sprintf("%v", val)

		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("example: ")
		schemaNodes.WriteString(exampleStr)
		schemaNodes.WriteString("\n")

		return exampleStr, schemaNodes.String()
	}
}
