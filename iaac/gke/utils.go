package gke

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// import (
// 	"math/rand"
// 	"time"
// )

// // generateRandomString is a helper function to generate a random string of lowercase letters and numbers
// func generateRandomString(
// 	length int,
// ) string {
// 	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
// 	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	b := make([]byte, length)
// 	for i := range b {
// 		b[i] = charset[seededRand.Intn(len(charset))]
// 	}
// 	return string(b)
// }

func mergeStringMaps(a, b pulumi.StringMap) pulumi.StringMap {
	result := pulumi.StringMap{}
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

// mergeNodePoolConfigs merges two maps of NodePoolConfig.
func mergeNodePoolConfigs(a, b NodePoolConfigs) NodePoolConfigs {
	merged := make(NodePoolConfigs)
	// Add all entries from a.
	for k, v := range a {
		merged[k] = v
	}
	// Add all entries from b (overriding duplicates).
	for k, v := range b {
		merged[k] = v
	}
	return merged
}
