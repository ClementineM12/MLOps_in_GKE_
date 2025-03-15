package global

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v2"
)

var logLevel = "INFO" //  TO DO: set a log Level field

// formatListIntoString is a helper function to format the regions for printing
func formatListIntoString(values []string) string {
	var valuesNames []string
	for _, val := range values {
		valuesNames = append(valuesNames, val)
	}
	return strings.Join(valuesNames, ", ")
}

func listContains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// CheckFileExists checks if a file exists at the given path
func CheckFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

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

// GetValues reads a YAML file from filePath, verifies its existence, substitutes dynamic placeholders,
// and returns the resulting data as a pulumi.MapInput.
func GetValues(
	filePath string,
	replacements map[string]interface{},
) (pulumi.MapInput, error) {

	// Check if the file exists.
	if !CheckFileExists(filePath) {
		return nil, fmt.Errorf("file %s does not exist", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values.yaml: %w", err)
	}

	// DEBUG: Print raw YAML content if needed.
	if logLevel == "DEBUG" {
		fmt.Println("🔹 Raw YAML Content:")
		fmt.Println(string(data))
	}

	// Unmarshal YAML into a generic map.
	var values interface{}
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	// Normalize the YAML in case keys are not strings.
	normalized := normalizeYAML(values).(map[string]interface{})

	// Substitute placeholders using the provided replacements.
	substituted := substitutePlaceholders(normalized, replacements).(map[string]interface{})

	// DEBUG: Print parsed and substituted YAML map if needed.
	if logLevel == "DEBUG" {
		fmt.Println("🔹 Substituted YAML Map:")
		printPrettyJSON(substituted)
	}

	// Convert to Pulumi MapInput.
	pulumiValues := convertToPulumiMap(substituted)
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

// substitutePlaceholders recursively replaces placeholders of the form ${key} in strings.
// If a placeholder is found and exists in the replacements map, it is replaced with its value.
func substitutePlaceholders(input interface{}, replacements map[string]interface{}) interface{} {
	switch v := input.(type) {
	case string:
		// Use a regex to find all placeholders in the string.
		re := regexp.MustCompile(`\$\{([^}]+)\}`)
		return re.ReplaceAllStringFunc(v, func(match string) string {
			// Extract the key name without the ${ and }.
			key := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
			if replacement, ok := replacements[key]; ok {
				// Convert the replacement value to a string.
				return fmt.Sprintf("%v", replacement)
			}
			// If no replacement is found, return the original match.
			return match
		})
	case map[string]interface{}:
		for key, value := range v {
			v[key] = substitutePlaceholders(value, replacements)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = substitutePlaceholders(item, replacements)
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

// printPrettyJSON is a helper function to pretty-print JSON (for debugging).
func printPrettyJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
	} else {
		fmt.Println(string(jsonData))
	}
}

// GenerateRandomString is a helper function to generate a random string of lowercase letters and numbers
func GenerateRandomString(
	length int,
) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
