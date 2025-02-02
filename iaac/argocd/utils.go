package argocd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v2"
)

var logLevel = "INFO" //  TO DO: config.get(ctx, "argocd:logLevel")

// Convert map[interface{}]interface{} to map[string]interface{} recursively
func normalizeYAML(input interface{}) interface{} {
	switch v := input.(type) {
	case map[interface{}]interface{}: // ❌ YAML sometimes parses maps with interface keys
		newMap := make(map[string]interface{})
		for key, value := range v {
			strKey, ok := key.(string)
			if !ok {
				continue // Skip non-string keys (should not happen in YAML)
			}
			newMap[strKey] = normalizeYAML(value) // Recursively process values
		}
		return newMap
	case []interface{}: // Handle YAML arrays correctly
		for i, item := range v {
			v[i] = normalizeYAML(item)
		}
		return v
	default:
		return v
	}
}

// Convert map[string]interface{} to pulumi.MapInput
func convertToPulumiMap(input map[string]interface{}) pulumi.MapInput {
	pulumiMap := pulumi.Map{}
	for key, value := range input {
		pulumiMap[key] = pulumiMapConvert(normalizeYAML(value))
	}
	return pulumiMap
}

// Reads YAML file and converts it to pulumi.MapInput
func getValues(filePath string) (pulumi.MapInput, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values.yaml: %w", err)
	}

	// DEBUG: Print raw YAML content
	if logLevel == "DEBUG" {
		fmt.Println("🔹 Raw YAML Content:")
		fmt.Println(string(data))
	}

	var values map[string]interface{}
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	// DEBUG: Print parsed YAML map
	if logLevel == "DEBUG" {
		fmt.Println("🔹 Parsed YAML Map:")
		printPrettyJSON(values)
	}

	// Convert to Pulumi Map
	pulumiValues := convertToPulumiMap(values)

	// DEBUG: Print converted Pulumi Map
	if logLevel == "DEBUG" {
		fmt.Println("🔹 Pulumi Converted Map:")
		printPrettyJSON(pulumiValues)
	}

	return pulumiValues, nil
}

// Recursively convert interface{} values to pulumi.Input values
func pulumiMapConvert(value interface{}) pulumi.Input {
	switch v := value.(type) {
	case map[string]interface{}:
		return convertToPulumiMap(v)
	case []interface{}:
		var arr pulumi.Array
		for _, item := range v {
			arr = append(arr, pulumiMapConvert(item))
		}
		return arr
	case string:
		return pulumi.String(v)
	case int:
		return pulumi.Int(v)
	case float64:
		return pulumi.Float64(v)
	case bool:
		return pulumi.Bool(v)
	default:
		return pulumi.Any(v)
	}
}

// DEBUG: Helper function: Pretty Print JSON for Debugging
func printPrettyJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
	} else {
		fmt.Println(string(jsonData))
	}
}
