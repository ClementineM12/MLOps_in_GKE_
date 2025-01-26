package gke

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Configuration reads and applies configuration values for the GKE cluster
func Configuration(ctx *pulumi.Context) *ClusterConfig {
	// Initialize NodePoolConfig with defaults
	defaultNodePool := NodePoolConfig{
		MachineType:  "e2-medium",
		DiskSizeGb:   100,
		DiskType:     "pd-standard",
		MinNodeCount: 1,
		MaxNodeCount: 3,
	}

	nodePool := NodePoolConfig{
		MachineType:  config.Get(ctx, "gke:nodePoolMachineType"),
		DiskSizeGb:   config.GetInt(ctx, "gke:nodePoolDiskSizeGb"),
		DiskType:     config.Get(ctx, "gke:nodePoolDiskType"),
		MinNodeCount: config.GetInt(ctx, "gke:nodePoolMinNodeCount"),
		MaxNodeCount: config.GetInt(ctx, "gke:nodePoolMaxNodeCount"),
		Preemptible:  config.GetBool(ctx, "gke:nodePoolPreemptible"),
	}

	management := ManagementConfig{
		AutoRepair:  config.GetBool(ctx, "gke:managementAutoRepair"),
		AutoUpgrade: config.GetBool(ctx, "gke:managementAutoUpgrade"),
	}

	// Apply defaults for any missing values
	if nodePool.MachineType == "" {
		nodePool.MachineType = defaultNodePool.MachineType
	}
	if nodePool.DiskSizeGb == 0 {
		nodePool.DiskSizeGb = defaultNodePool.DiskSizeGb
	}
	if nodePool.DiskType == "" {
		nodePool.DiskType = defaultNodePool.DiskType
	}
	if nodePool.MinNodeCount == 0 {
		nodePool.MinNodeCount = defaultNodePool.MinNodeCount
	}
	if nodePool.MaxNodeCount == 0 {
		nodePool.MaxNodeCount = defaultNodePool.MaxNodeCount
	}

	clusterConfig := &ClusterConfig{
		Name:       config.Get(ctx, "gke:name"),
		Cidr:       config.Get(ctx, "gke:cidr"),
		NodePool:   nodePool,
		Management: management,
	}

	// Apply defaults for missing cluster-level values
	if clusterConfig.Name == "" {
		clusterConfig.Name = "default"
	}
	if clusterConfig.Cidr == "" {
		clusterConfig.Cidr = "10.0.0.0/16"
	}

	return clusterConfig
}
