package gke

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

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
