package vpc

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
)

type CloudRegion struct {
	Id             string
	Enabled        bool
	Region         string
	SubnetIp       string
	GKECluster     *container.Cluster
	GKEClusterName string
}
