package project

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ProjectOutputs defines the outputs of the module
type ProjectOutputs struct {
	ProjectID   string
	ProjectName pulumi.StringOutput
}

type Projects struct {
	Id   string
	Name string
}

type ProjectConfig struct {
	ResourceNamePrefix string
	ProjectId          string
	Domain             string
	SSL                bool
	EnabledRegions     []CloudRegion
	Target             string
}

type CloudRegion struct {
	Id             string
	Country        string
	Region         string
	SubnetIp       string
	GKECluster     *container.Cluster
	GKEClusterName string
}
