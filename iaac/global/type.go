package global

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
	EnabledRegion      CloudRegion
	Target             string
	CloudSQL           *CloudSQLConfig
	Email              string
	WhitelistedIPs     string
	ArtifactRegistry   ArtifactRegistryConfig
}

type CloudRegion struct {
	Id                  string
	Country             string
	Region              string
	SubnetIp            string
	GKECluster          *container.Cluster
	GKEClusterName      string
	MasterIpv4CidrBlock string
}

type CloudSQLConfig struct {
	Create             bool   `json:"create"`
	User               string `json:"user"`
	Database           string `json:"database"`
	InstancePrefixName string
	InstanceName       pulumi.StringOutput
	Connection         pulumi.StringOutput
	Password           pulumi.StringOutput
	DatabaseName       pulumi.StringOutput
}

type ArtifactRegistryConfig struct {
	GithubRepo                                string
	RegistryName                              string
	GithubServiceAccountCreate                bool
	ContinuousDevelopmentServiceAccountCreate bool
}
