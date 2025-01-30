package gke

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NodePoolConfig holds the configuration for a GKE node pool
type NodePoolConfig struct {
	MachineType            string
	DiskSizeGb             int
	DiskType               string
	Preemptible            bool
	MinMasterVersion       pulumi.String
	MinNodeCount           int
	MaxNodeCount           int
	OauthScopes            pulumi.StringArray
	WorkloadMetadataConfig *container.NodePoolNodeConfigWorkloadMetadataConfigArgs
	Metadata               pulumi.StringMap
	ResourceLabels         pulumi.StringMap
}

type ManagementConfig struct {
	AutoRepair  bool
	AutoUpgrade bool
}

// ClusterConfig holds the overall GKE cluster configuration
type ClusterConfig struct {
	Name       string
	NodePool   NodePoolConfig
	Management ManagementConfig
	Cidr       string
}
