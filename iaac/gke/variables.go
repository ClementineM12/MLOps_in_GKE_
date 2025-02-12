package gke

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	GKEDefaultVersion = "1.31.4-gke.1183000"

	GKENodePoolSpecificConfig = map[string]NodePoolConfig{
		"management": {
			MinMasterVersion: pulumi.String(GKEDefaultVersion),
			OauthScopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/devstorage.read_only"),
				pulumi.String("https://www.googleapis.com/auth/logging.write"),
				pulumi.String("https://www.googleapis.com/auth/monitoring"),
				pulumi.String("https://www.googleapis.com/auth/trace.append"),
				pulumi.String("https://www.googleapis.com/auth/cloud-platform"),
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
			MinMasterVersion: pulumi.String(GKEDefaultVersion),
			OauthScopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/devstorage.read_only"),
				pulumi.String("https://www.googleapis.com/auth/logging.write"),
				pulumi.String("https://www.googleapis.com/auth/monitoring"),
			},
			WorkloadMetadataConfig: &container.NodePoolNodeConfigWorkloadMetadataConfigArgs{},
			Metadata:               pulumi.StringMap{},
		},
	}
)
