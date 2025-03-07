package gke

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NodePoolConfig holds the configuration for a GKE node pool
type NodePoolConfig struct {
	KeyName                string
	MachineType            string
	DiskSizeGb             int
	DiskType               string
	Preemptible            bool
	InitialNodeCount       int
	MinMasterVersion       pulumi.String
	MinNodeCount           int
	MaxNodeCount           int
	OauthScopes            pulumi.StringArray
	WorkloadMetadataConfig *container.NodePoolNodeConfigWorkloadMetadataConfigArgs
	Metadata               pulumi.StringMap
	ResourceLabels         pulumi.StringMap
	Labels                 pulumi.StringMap
	LocationPolicy         string
}

type NodePoolConfigs = map[string]NodePoolConfig
type ManagementConfig struct {
	AutoRepair  bool
	AutoUpgrade bool
}

// ClusterConfig holds the overall GKE cluster configuration
type ClusterConfig struct {
	Name       string
	NodePools  NodePoolConfigs
	Management ManagementConfig
	Cidr       string
}
