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

func formatExampleString(s string) string {
	var exampleString strings.Builder

	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	linesLength := len(lines)

	if linesLength < 3 || 
		strings.TrimSpace(lines[0]) != "{" || 
		strings.TrimSpace(lines[linesLength - 1]) != "}" {
		return s
	}

	depth := 1

	for _, l := range lines[1 : linesLength - 1] {
		currentLine := strings.TrimSpace(l)

		if strings.ContainsRune(currentLine, ':') { 
			splittedLine := strings.Split(currentLine, ":")
			splittedLine[0] = strings.ReplaceAll(splittedLine[0], "\"", "")
			currentLine = strings.Join(splittedLine, ":")
		}

		if currentLine == "}" || currentLine == "}," || currentLine == "]" || currentLine == "],"{
			depth -= 1
		}

		if depth == 1 {
			currentLine = strings.TrimSuffix(currentLine, ",")
		}

		exampleString.WriteString(strings.Repeat("  ", depth + 1))
		exampleString.WriteString(currentLine)
		exampleString.WriteString("\n")

		if strings.HasSuffix(currentLine, "{") || strings.HasSuffix(currentLine, "[") {
			depth += 1
		}
	}

	return exampleString.String()
}
