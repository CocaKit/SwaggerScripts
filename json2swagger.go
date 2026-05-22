package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func json2swagger(sortExample bool, sortSchema bool) {
	var inputData []byte
	var err error

	inputData, err = os.ReadFile("input/input.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var jsonData interface{}
	if err := json.Unmarshal(inputData, &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid JSON input: %v\n", err)
		os.Exit(1)
	}

	var exampleText strings.Builder

	exampleText.WriteString("ExampleName:\n")
	exampleText.WriteString("  value:\n")
	exampleText.WriteString(formatExampleNodes(jsonData, 1, true, sortExample))

	err = os.WriteFile("output/exampleSwag.yaml", []byte(exampleText.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing example file: %v\n", err)
		os.Exit(1)
	}

	inputData, err = os.ReadFile("output/parsedProps.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading parsed propertions file: %v\n", err)
		os.Exit(1)
	}

	var entities map[string]Entity
	err = json.Unmarshal(inputData, &entities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON for propertions: %v\n", err)
		os.Exit(1)
	}

	var schemaText strings.Builder

	schemaText.WriteString("SchemaName:\n")
	schemaText.WriteString(formatSchemaNodes(jsonData, 1, entities, []string{}, sortSchema))

	err = os.WriteFile("output/schemaSwag.yaml", []byte(schemaText.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema file: %v\n", err)
		os.Exit(1)
	}
}

func inputPropValueForDescription(mainStr string, keyName string, props map[string]Entity) string {
	fmt.Println("### SET DESCRIPTION FOR:")
	fmt.Println(mainStr)

	prop, exist := props[keyName]	

	return fmt.Sprintf("%s", inputWithList(prop.Descriptions, exist))
}

func inputPropValueForExample(mainStr string, keyName string, props map[string]Entity) any {
	fmt.Println("### SET EXAMPLE FOR:")
	fmt.Println(mainStr)

	prop, exist := props[keyName]	

	return inputWithList(prop.Examples, exist)
}

func inputWithList[T any](propList []T, propIsExist bool) any {
	if propIsExist {
		for propIndex, propVar := range propList {
			fmt.Printf("%d: %v\n", propIndex + 1, propVar)	
		}
	}

	fmt.Printf("0: == Input own ==\n")	

	for {
		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')

		if err != nil {
			return ""
		}

		choice = strings.Trim(choice, " \n")

		if choice == "0" {
			fmt.Printf("New value: ")	

			reader = bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return ""
			}

			return strings.Trim(input, " \n")
		}

		idx, err := strconv.Atoi(choice)
		if err != nil {
			return ""
		}

		idx = idx - 1
		if !propIsExist || idx < 0 || idx >= len(propList) {
			fmt.Println("Input valid number:")
			continue
		}

		return propList[idx]
	}
}

func formatExampleNodes(v interface{}, indent int, isRoot bool, sortKeys bool) string {
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

		if sortKeys {
			sort.Strings(keys)
		}

		var exampleNodes strings.Builder
		if !isRoot {
			exampleNodes.WriteString("{\n")
		}

		for i, k := range keys {
			exampleNodes.WriteString(innerIndent)
			exampleNodes.WriteString(k)
			exampleNodes.WriteString(": ")

			exampleNodes.WriteString(formatExampleNodes(val[k], indent + 1, false, sortKeys))

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
		exampleNodes.WriteString(formatExampleNodes(val[0], indent + 1, false, sortKeys))
		exampleNodes.WriteString("\n")
		exampleNodes.WriteString(baseIndent)
		exampleNodes.WriteString("]")

		return exampleNodes.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

//TODO automatic select arrays and objects from json if empty
//TODO automatic fill null values (need type parsing)
func formatSchemaNodes(v interface{}, indent int, props map[string]Entity, keyNames []string, sortKeys bool) string {
	baseIndent := strings.Repeat("  ", indent)
	innerIndent := strings.Repeat("  ", indent + 1)

	keyNamesStr := strings.Join(keyNames, " -> ")

	switch val := v.(type) {
	case map[string]interface{}:
		var schemaNodes strings.Builder
		
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: object\n")

		if len(val) == 0 {
			fmt.Println("!!! EMPTY OBJECT !!! " + keyNamesStr)
			return schemaNodes.String()
		}

		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}

		if sortKeys {
			sort.Strings(keys)
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("properties:\n")

		for _, k := range keys {
			schemaNodes.WriteString(innerIndent)
			schemaNodes.WriteString(k)
			schemaNodes.WriteString(":\n")
			schemaNodes.WriteString(formatSchemaNodes(val[k], indent + 2, props, append(keyNames, k), sortKeys))
		}

		return schemaNodes.String()
	case string:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: string\n")

		descriptionStr := ""
		if len(keyNames) > 0 {
			descriptionStr = inputPropValueForDescription("string: " + keyNamesStr, keyNames[len(keyNames) - 1], props)
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("description: %s\n", descriptionStr))

		schemaNodes.WriteString(baseIndent)
		if (val == "") {
			schemaNodes.WriteString(fmt.Sprintf("example: %q\n", inputPropValueForExample("string: " + keyNamesStr, keyNames[len(keyNames) - 1], props)))
		} else {
			schemaNodes.WriteString(fmt.Sprintf("example: %q\n", val))
		}

		return schemaNodes.String()
	case float64:
		var exampleStr string
		var typeStr string
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)

		if val == float64(int64(val)) {
			exampleStr = fmt.Sprintf("%d", int64(val))
			typeStr = "integer"
		} else {
			exampleStr = fmt.Sprintf("%g", val)
			typeStr = "float"
		}

		schemaNodes.WriteString(fmt.Sprintf("type: %s\n", typeStr))

		descriptionStr := ""
		if len(keyNames) > 0 {
			descriptionStr = inputPropValueForDescription(fmt.Sprintf("%s: %s", typeStr, keyNamesStr), keyNames[len(keyNames) - 1], props)
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("description: %s\n", descriptionStr))

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %s\n", exampleStr))

		return schemaNodes.String()
	case bool:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: bool\n")

		descriptionStr := ""
		if len(keyNames) > 0 {
			descriptionStr = inputPropValueForDescription("bool: " + keyNamesStr, keyNames[len(keyNames) - 1], props)
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("description: %s\n", descriptionStr))

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %t\n", val))

		return schemaNodes.String()
	case []interface{}:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: array\n")
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("items:\n")

		if len(val) == 0 {
			fmt.Println("!!! EMPTY ARRAY !!! " + keyNamesStr)
			return schemaNodes.String() 
		}

		schemaNodes.WriteString(formatSchemaNodes(val[0], indent+1, props, keyNames, sortKeys))

		return schemaNodes.String()
	default:
		var schemaNodes strings.Builder

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %v\n"))

		fmt.Println("!!! WRONG TYPE !!! " + keyNamesStr)

		return schemaNodes.String()
	}
}
