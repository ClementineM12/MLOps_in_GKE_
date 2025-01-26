package gke

// NodePoolConfig holds the configuration for a GKE node pool
type NodePoolConfig struct {
	MachineType  string
	DiskSizeGb   int
	DiskType     string
	Preemptible  bool
	MinNodeCount int
	MaxNodeCount int
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
