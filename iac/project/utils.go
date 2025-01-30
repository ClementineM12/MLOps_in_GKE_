package project

import (
	"os"
	"strings"
)

// formatListIntoString is a helper function to format the regions for printing
func formatRegions(regions []CloudRegion) string {
	var regionNames []string
	for _, region := range regions {
		regionNames = append(regionNames, region.Region)
	}
	return strings.Join(regionNames, ", ")
}

// CheckFileExists checks if a file exists at the given path
func CheckFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
