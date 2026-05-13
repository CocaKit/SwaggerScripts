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

	var exampleText strings.Builder

	exampleText.WriteString("ExampleName:\n")
	exampleText.WriteString("  value:")
	exampleText.WriteString(formatExampleNodes(jsonData, 1, true))

	err = os.WriteFile("output/exampleSwag.yaml", []byte(exampleText.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing example file: %v\n", err)
		os.Exit(1)
	}

	var schemaText strings.Builder

	schemaText.WriteString("SchemaName:\n")
	schemaText.WriteString(formatSchemaNodes(jsonData, 1))

	err = os.WriteFile("output/schemaSwag.yaml", []byte(schemaText.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema file: %v\n", err)
		os.Exit(1)
	}
}

func formatExampleNodes(v interface{}, indent int, isRoot bool) string {
	baseIndent := strings.Repeat("  ", indent)
	innerIndent := strings.Repeat("  ", indent + 1)

	switch val := v.(type) {
	case map[string]interface{}:
		if len(val) == 0 {
			if isRoot {
				return ""
			}
			return "{}"
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

		for i, k := range keys {
			exampleNodes.WriteString(innerIndent)
			exampleNodes.WriteString(k)
			exampleNodes.WriteString(": ")

			exampleNodes.WriteString(formatExampleNodes(val[k], indent + 1, false))

			if i < len(keys) - 1 && !isRoot {
				exampleNodes.WriteString(",")
			}
			exampleNodes.WriteString("\n")
		}

		if !isRoot {
			exampleNodes.WriteString(baseIndent)
			exampleNodes.WriteString("}")
		}

		return exampleNodes.String()
	case string:
		return fmt.Sprintf("%q", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}

		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case nil:
		return "null"
	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}

		var exampleNodes strings.Builder

		exampleNodes.WriteString("[\n")
		exampleNodes.WriteString(innerIndent)
		exampleNodes.WriteString(formatExampleNodes(val[0], indent + 1, false))
		exampleNodes.WriteString("\n")
		exampleNodes.WriteString(baseIndent)
		exampleNodes.WriteString("]")

		return exampleNodes.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatSchemaNodes(v interface{}, indent int) string {
	baseIndent := strings.Repeat("  ", indent)
	innerIndent := strings.Repeat("  ", indent + 1)

	switch val := v.(type) {
	case map[string]interface{}:
		var schemaNodes strings.Builder
		
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: object\n")

		if len(val) == 0 {
			return schemaNodes.String()
		}

		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("properties:\n")

		for _, k := range keys {
			schemaNodes.WriteString(innerIndent)
			schemaNodes.WriteString(k)
			schemaNodes.WriteString(":\n")
			schemaNodes.WriteString(formatSchemaNodes(val[k], indent + 2))
		}

		return schemaNodes.String()
	case string:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: string\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %q\n", val))

		return schemaNodes.String()
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
		schemaNodes.WriteString(fmt.Sprintf("example: %s\n", exampleStr))

		return schemaNodes.String()
	case bool:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: bool\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %t\n", val))

		return schemaNodes.String()
	case nil:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("example: null\n")

		return schemaNodes.String()
	case []interface{}:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: array\n")
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("items:\n")

		if len(val) == 0 {
			return schemaNodes.String() 
		}

		schemaNodes.WriteString(formatSchemaNodes(val[0], indent+1))

		return schemaNodes.String()
	default:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %v\n", val))

		return schemaNodes.String()
	}
}
