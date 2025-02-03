package gke

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var (
	GKEDefaultCIDR = "10.0.0.0/16"
)

// Configuration reads and applies configuration values for the GKE cluster
func Configuration(
	ctx *pulumi.Context,
) *ClusterConfig {

	target := config.Get(ctx, "gke:target")

	// Initialize NodePoolConfig with defaults
	defaultNodePool := NodePoolConfig{
		MachineType:  "e2-medium",
		DiskSizeGb:   100,
		DiskType:     "pd-standard",
		MinNodeCount: 1,
		MaxNodeCount: 3,
	}

	nodePool := NodePoolConfig{
		MachineType:            config.Get(ctx, "gke:nodePoolMachineType"),
		DiskSizeGb:             config.GetInt(ctx, "gke:nodePoolDiskSizeGb"),
		DiskType:               config.Get(ctx, "gke:nodePoolDiskType"),
		MinMasterVersion:       GKENodePoolSpecificConfig[target].MinMasterVersion,
		MinNodeCount:           config.GetInt(ctx, "gke:nodePoolMinNodeCount"),
		MaxNodeCount:           config.GetInt(ctx, "gke:nodePoolMaxNodeCount"),
		Preemptible:            config.GetBool(ctx, "gke:nodePoolPreemptible"),
		OauthScopes:            GKENodePoolSpecificConfig[target].OauthScopes,
		WorkloadMetadataConfig: GKENodePoolSpecificConfig[target].WorkloadMetadataConfig,
		Metadata:               GKENodePoolSpecificConfig[target].Metadata,
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
		clusterConfig.Cidr = GKEDefaultCIDR
	}

	return clusterConfig
}
