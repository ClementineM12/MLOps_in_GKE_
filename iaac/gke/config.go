package gke

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var (
	GKEDefaultName = "default"
	GKEDefaultCIDR = "10.0.0.0/16"
)

// Configuration reads and applies configuration values for the GKE cluster
func Configuration(
	ctx *pulumi.Context,
) *ClusterConfig {

	nodePoolConfigs := configureNodePools(ctx)
	management := ManagementConfig{
		AutoRepair:  config.GetBool(ctx, "gke:managementAutoRepair"),
		AutoUpgrade: config.GetBool(ctx, "gke:managementAutoUpgrade"),
	}

	clusterConfig := &ClusterConfig{
		Name:       config.Get(ctx, "gke:name"),
		Cidr:       config.Get(ctx, "gke:cidr"),
		NodePools:  nodePoolConfigs,
		Management: management,
	}

	// Apply defaults for missing cluster-level values
	if clusterConfig.Name == "" {
		clusterConfig.Name = GKEDefaultName
	}
	if clusterConfig.Cidr == "" {
		clusterConfig.Cidr = GKEDefaultCIDR
	}

	return clusterConfig
}

// configureNodePools reads the base configuration from Pulumi, then merges it with the specific overrides.
func configureNodePools(ctx *pulumi.Context) NodePoolConfigs {
	// Initialize NodePoolConfig with defaults.
	defaultNodePool := NodePoolConfig{
		MachineType:  "e2-standard-4",
		DiskSizeGb:   100,
		MaxNodeCount: 5,
		DiskType:     "pd-standard",
	}

	// Read the base configuration from Pulumi.
	base := NodePoolConfig{
		MachineType:      config.Get(ctx, "gke:nodePoolMachineType"),
		DiskSizeGb:       config.GetInt(ctx, "gke:nodePoolDiskSizeGb"),
		DiskType:         config.Get(ctx, "gke:nodePoolDiskType"),
		InitialNodeCount: 1,
		MinNodeCount:     3,
		MaxNodeCount:     config.GetInt(ctx, "gke:nodePoolMaxNodeCount"),
		Preemptible:      config.GetBool(ctx, "gke:nodePoolPreemptible"),
		// Assuming LocationPolicy is a field of NodePoolConfig.
		LocationPolicy: "BALANCED",
	}

	// Merge the base config with the overrides.
	allNodePools := mergeNodePoolConfigs(
		NodePoolConfigs{"base": base},
		nodePoolsConfig,
	)

	// Create a new map to hold the final merged configurations.
	mergedConfigs := make(NodePoolConfigs)

	// Iterate over all merged node pool configurations.
	for key, np := range allNodePools {
		// Apply defaults for any missing values.
		if np.MachineType == "" {
			np.MachineType = defaultNodePool.MachineType
		}
		if np.DiskSizeGb == 0 {
			np.DiskSizeGb = defaultNodePool.DiskSizeGb
		}
		if np.MaxNodeCount == 0 {
			np.MaxNodeCount = defaultNodePool.MaxNodeCount
		}
		if np.DiskType == "" {
			np.DiskType = defaultNodePool.DiskType
		}
		if np.Metadata == nil {
			np.Metadata = pulumi.StringMap{
				"disable-legacy-endpoints": pulumi.String("true"),
			}
		}
		if np.WorkloadMetadataConfig == nil {
			np.WorkloadMetadataConfig = &container.NodePoolNodeConfigWorkloadMetadataConfigArgs{
				Mode: pulumi.String("GKE_METADATA"),
			}
		}

		// Set the KeyName based on the map key.
		if key == "base" {
			np.KeyName = ""
		} else {
			np.KeyName = key
		}

		mergedConfigs[key] = np
	}

	return mergedConfigs
}
