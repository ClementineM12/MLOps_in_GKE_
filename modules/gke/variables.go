package gke

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var gkeNodePoolSpecificConfig = map[string]NodePoolConfig{
	"management": {
		MinMasterVersion: pulumi.String("1.28.15-gke.1612000"),
		OauthScopes: pulumi.StringArray{
			pulumi.String("https://www.googleapis.com/auth/devstorage.read_only"),
			pulumi.String("https://www.googleapis.com/auth/logging.write"),
			pulumi.String("https://www.googleapis.com/auth/monitoring"),
			pulumi.String("https://www.googleapis.com/auth/trace.append"),
		},
		WorkloadMetadataConfig: &container.NodePoolNodeConfigWorkloadMetadataConfigArgs{
			Mode: pulumi.String("GKE_METADATA"),
		},
		Metadata: pulumi.StringMap{
			"disable-legacy-endpoints": pulumi.String("true"),
		},
		ResourceLabels: pulumi.StringMap{},
	},
	"development": {
		MinMasterVersion:       pulumi.String("1.31.4-gke.1183000"),
		OauthScopes:            pulumi.StringArray{},
		WorkloadMetadataConfig: &container.NodePoolNodeConfigWorkloadMetadataConfigArgs{},
		Metadata:               pulumi.StringMap{},
	},
}
