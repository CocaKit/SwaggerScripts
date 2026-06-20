package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-faster/jx"
)

//TODO make web service
func json2swagger() {
	var inputData []byte
	var err error

	inputData, err = os.ReadFile("input/input.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var exampleText strings.Builder

	exampleText.WriteString("ExampleName:\n")
	exampleText.WriteString("  value:\n")
	exampleText.WriteString(formatExampleNodes(jx.DecodeBytes(inputData), 1, true))

	err = os.WriteFile("output/exampleSwag.yaml", []byte(exampleText.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing example file: %v\n", err)
		os.Exit(1)
	}

	parsedPropsData, err := os.ReadFile("output/parsedProps.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading parsed propertions file: %v\n", err)
		os.Exit(1)
	}

	var entities map[string]Entity
	err = json.Unmarshal(parsedPropsData, &entities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON for propertions: %v\n", err)
		os.Exit(1)
	}

	var schemaText strings.Builder

	schemaText.WriteString("SchemaName:\n")
	schemaText.WriteString(formatSchemaNodes(jx.DecodeBytes(inputData), 1, entities, []string{}))

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

//TODO automatic select arrays and objects from json if empty
//TODO automatic fill null values (need type parsing)
//TODO place on top more suitable variants (checking element key nesting)
func formatSchemaNodes(d *jx.Decoder, indent int, props map[string]Entity, keyNames []string) string {
	baseIndent := strings.Repeat("  ", indent)
	innerIndent := strings.Repeat("  ", indent + 1)

	keyNamesStr := strings.Join(keyNames, " -> ")

	nextType := d.Next()

	var schemaNodes strings.Builder

	switch nextType {
	case jx.Object:
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: object\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("properties:\n")

		isEmpty := true

		d.Obj(func(d *jx.Decoder, key string) error {
			isEmpty = false
			schemaNodes.WriteString(innerIndent)
			schemaNodes.WriteString(key)
			schemaNodes.WriteString(":\n")
			schemaNodes.WriteString(formatSchemaNodes(d, indent + 2, props, append(keyNames, key)))
			
			return nil
		})
		
		if isEmpty {
			fmt.Println("!!! EMPTY OBJECT !!! " + keyNamesStr)
		}

	case jx.Array:
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: array\n")
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("items:\n")

		arrayIsEmpty := true
		isFirst := true

		d.Arr(func(d *jx.Decoder) error {
			arrayIsEmpty = false

			if isFirst {
				isFirst = false
				schemaNodes.WriteString(formatSchemaNodes(d, indent+1, props, keyNames))
				return nil
			}

			return d.Skip()
		})

		if arrayIsEmpty {
			fmt.Println("!!! EMPTY ARRAY !!! " + keyNamesStr)
		}

	case jx.Number:
		var exampleStr string
		var typeStr string

		num, _ := d.Num()

		if num.IsInt() {
			val, _ := num.Int64()
			exampleStr = fmt.Sprintf("%d", int64(val))
			typeStr = "integer"
		} else {
			val, _ := num.Float64()
			exampleStr = fmt.Sprintf("%g", val)
			typeStr = "float"
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("type: %s\n", typeStr))

		descriptionStr := ""
		if len(keyNames) > 0 {
			descriptionStr = inputPropValueForDescription(fmt.Sprintf("%s: %s", typeStr, keyNamesStr), keyNames[len(keyNames) - 1], props)
		}

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("description: %s\n", descriptionStr))

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: %s\n", exampleStr))

	case jx.String:
		val, _ := d.Str()

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

	case jx.Bool:
		val, _ := d.Bool()

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
	case jx.Null:
		_ = d.Null()

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type: null\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example: null\n"))

		fmt.Println("!!! NULL TYPE !!! " + keyNamesStr)

	default:
		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("type:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString("description:\n")

		schemaNodes.WriteString(baseIndent)
		schemaNodes.WriteString(fmt.Sprintf("example:\n"))

		fmt.Println("!!! WRONG TYPE !!! " + keyNamesStr)
	}

	return schemaNodes.String()
}

func formatExampleNodes(d *jx.Decoder, indent int, isRoot bool) string {
	baseIndent := strings.Repeat("  ", indent)
	innerIndent := strings.Repeat("  ", indent + 1)

	nextType := d.Next()

	var exampleNodes strings.Builder

	switch nextType {
	case jx.Object:
		if !isRoot {
			exampleNodes.WriteString("{")
		}

		var innerNodes strings.Builder

		d.Obj(func(d *jx.Decoder, key string) error {
			innerNodes.WriteString(fmt.Sprintf("%s%s: ", innerIndent, key))
			innerNodes.WriteString(formatExampleNodes(d, indent + 1, false))

			if !isRoot {
				innerNodes.WriteString(",")
			}
			innerNodes.WriteString("\n")
			
			return nil
		})

		if innerNodes.Len() > 0 {
			if !isRoot {
				exampleNodes.WriteString("\n")
			}

			innerString := innerNodes.String()
			innerStringLen := len(innerString)

			if (innerStringLen > 2 && innerString[innerStringLen - 2] == ',' && innerString[innerStringLen - 1] == '\n') {
				innerString = innerString[:innerStringLen - 2]
			}

			exampleNodes.WriteString(innerString)

			if !isRoot {
				exampleNodes.WriteString("\n")
				exampleNodes.WriteString(baseIndent)
			}
		}

		if !isRoot {
			exampleNodes.WriteString("}")
		}
	case jx.Array:
		isFirst := true

		exampleNodes.WriteString("[")

		var innerNodes strings.Builder

		d.Arr(func(d *jx.Decoder) error {

			if isFirst {
				isFirst = false
				innerNodes.WriteString(formatExampleNodes(d, indent + 1, false))
				innerNodes.WriteString("\n")
				return nil
			}

			return d.Skip()
		})

		if innerNodes.Len() > 0 {
			exampleNodes.WriteString(fmt.Sprintf("\n%s%s", innerIndent, innerNodes.String()))
			exampleNodes.WriteString(baseIndent)
		}

		exampleNodes.WriteString("]")
	case jx.String:
		val, _ := d.Str()
		exampleNodes.WriteString(fmt.Sprintf("%q", val))
	case jx.Number:
		num, _ := d.Num()

		if num.IsInt() {
			val, _ := num.Int64()
			exampleNodes.WriteString(fmt.Sprintf("%d", int64(val)))
		} else {
			val, _ := num.Float64()
			exampleNodes.WriteString(fmt.Sprintf("%g", val))
		}
	case jx.Bool:
		val, _ := d.Bool()
		exampleNodes.WriteString(fmt.Sprintf("%t", val))
	case jx.Null:
		_ = d.Null()
		exampleNodes.WriteString("null")
	default:
		val, _ := d.Raw()
		exampleNodes.WriteString(fmt.Sprintf("%v", val))
	}

	return exampleNodes.String()
}
