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

// ClusterConfig holds the overall GKE cluster configuration
type ClusterConfig struct {
	Name     string
	NodePool NodePoolConfig
	Create   bool
	Cidr     string
}
