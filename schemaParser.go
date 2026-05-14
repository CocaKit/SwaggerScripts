package main

import (
	"fmt"
	"os"
	"slices"
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type SchemaProperty struct {
	Description 	string 						`yaml:"description,omitempty"`
	Example 		interface{} 				`yaml:"example,omitempty"`
	Type 			string						`yaml:"type,omitempty"`
	Properties 		map[string]SchemaProperty 	`yaml:"properties,omitempty"`
	Items			*SchemaProperty				`yaml:"items,omitempty"`
}

type OpenAPISchema struct {
	Type 			string 						`yaml:"type,omitempty"`
	Properties 		map[string]SchemaProperty 	`yaml:"properties,omitempty"`
}

type OpenAPI struct {
	Components struct {
		Schemas 	map[string]OpenAPISchema 	`yaml:"schemas,omitempty"`
	} 											`yaml:"components,omitempty"`
}

type Entity struct {
	Examples 		[]interface{}				`json:"examples"`
	Descriptions 	[]string					`json:"descriptions"`
}

func appendIfNotExists[T comparable](slice []T, elem T) []T {
    if slices.Contains(slice, elem) {
        return slice
    }

    return append(slice, elem)
}

func extractProps(props map[string]SchemaProperty, entities map[string]*Entity) {
	for propName, prop := range props {
		if prop.Type == "object" {
			if prop.Properties != nil {
				extractProps(prop.Properties, entities)
			}

			continue
		}

		if prop.Type == "array" {
			if prop.Items != nil {
				extractProps(map[string]SchemaProperty{propName: *prop.Items}, entities)
			}

			continue
		}

		ent, exists := entities[propName]
		if !exists {
			ent = &Entity{}
			entities[propName] = ent
		}

		switch prop.Example.(type) {
		case string, int, float64, bool:
			ent.Examples = appendIfNotExists(ent.Examples, prop.Example)
		}

		if prop.Description != "" {
			ent.Descriptions = appendIfNotExists(ent.Descriptions, prop.Description)
		}
	}
}

func schemaParser() {
	var inputData []byte
	var err error

	inputData, err = os.ReadFile("input/schemasToParse.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var api OpenAPI
	err = yaml.Unmarshal(inputData, &api)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	var entities = make(map[string]*Entity)

	for _, schema := range api.Components.Schemas {
		if schema.Properties == nil {
			continue
		}

		extractProps(schema.Properties, entities)
	}


	data, err := json.MarshalIndent(entities, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("output/parsedProps.yaml", []byte(string(data)), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing parsed proportions file: %v\n", err)
		os.Exit(1)
	}
}
