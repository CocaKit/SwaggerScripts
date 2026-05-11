package main

import (
	"fmt"
	"io"
	"os"
	"gopkg.in/yaml.v3"
)

type SchemaProperty struct {
	Description string 						`yaml:"description,omitempty"`
	Example 	interface{} 				`yaml:"example,omitempty"`
	Type 		string						`yaml:"example,omitempty"`
	Properties map[string]SchemaProperty 	`yaml:"properties,omitempty"`
}

type OpenAPISchema struct {
	Type string 							`yaml:"type,omitempty"`
	Properties map[string]SchemaProperty 	`yaml:"properties,omitempty"`
}

type OpenAPI struct {
	Components struct {
		Schemas map[string]OpenAPISchema `yaml:"schemas,omitempty"`
	} `yaml:"components,omitempty"`
}

type Entity struct {
	Examples []interface{}
	Descriptions []string
}

func extractProps(props map[string]SchemaProperty, entities *[]Entity) {
	for propName, prop := range props {
		ent, exists := *entities[propName]
		if !exists {
			ent = &Entity{}
			*entities[propName] = ent
		}

		fmt.Printf("Property: %s\n", propName)

		switch prop.Example.(type) {
		case string, int, float64, bool:
			ent.Examples = append(ent.Examples, prop.Example)
			fmt.Printf("Appending example: %v\n", prop.Example)
		default:
			fmt.Printf("Ignoring example of type: %T\n", prop.Example)
		}

		if prop.Description != "" {
			ent.Descriptions = append(ent.Descriptions, prop.Description)
			fmt.Printf("Appending description: %s\n", prop.Description)
		} else {
			fmt.Printf("Ignoring empty string\n")
		}

		if prop.Type == "object" && prop.Properties != nil {
			extractProps(prop.Properties, entities)
		}
	}
}

func schemaParser() {
	var inputData []byte
	var err error

	inputData, err = io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var api OpenAPI
	err = yaml.Unmarshal(inputData, &api)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	var entities = make(map[string]*Entity)

	for schemaName, schema := range api.Components.Schemas {
		fmt.Printf("=== Schema %s ===\n", schemaName)	

		if schema.Type != "object" || schema.Properties == nil {
			fmt.Printf("Is't object or don't have properties\n")	
			continue
		}

		extractProps(*schema.Properties, entities)
	}

	fmt.Println()

	fmt.Printf("=== All entities ===\n")
	for entityName, entity := range entities {
		fmt.Printf("=== %s ===\n", entityName)

		
		fmt.Printf("Examples:\n")
		for _, example := range entity.Examples {
			fmt.Printf("%v\n", example)	
		}

		fmt.Printf("Descriptions:\n")
		for _, description := range entity.Descriptions {
			fmt.Printf("%s\n", description)	
		}
	}
}
