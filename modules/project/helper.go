package project

import (
	"math/rand"
	"strings"
	"time"
)

// generateRandomString is a helper function to generate a random string of lowercase letters and numbers
func generateRandomString(
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

// formatListIntoString is a helper function to format the regions for printing
func formatRegions(regions []CloudRegion) string {
	var regionNames []string
	for _, region := range regions {
		regionNames = append(regionNames, region.Region)
	}
	return strings.Join(regionNames, ", ")
}
