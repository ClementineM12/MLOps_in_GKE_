package project

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateWorkloadIdentityPool creates Google Cloud Workload Identity Pool for GKE
func CreateWorkloadIdentityPool(
	ctx *pulumi.Context,
	projectConfig ProjectConfig,
) error {
	// Generate a unique suffix for the Workload Identity Pool ID
	randomSuffix := generateRandomString(6)
	resourceName := fmt.Sprintf("%s-wip-gke-cluster", projectConfig.ResourceNamePrefix)

	// Create the Workload Identity Pool
	_, err := iam.NewWorkloadIdentityPool(ctx, resourceName, &iam.WorkloadIdentityPoolArgs{
		Project:                pulumi.String(projectConfig.ProjectId),
		Description:            pulumi.String("GKE - Workload Identity Pool for GKE Cluster"),
		Disabled:               pulumi.Bool(false),
		DisplayName:            pulumi.String(resourceName),
		WorkloadIdentityPoolId: pulumi.String(fmt.Sprintf("%s-%s", projectConfig.ResourceNamePrefix, randomSuffix)),
	})
	if err != nil {
		return fmt.Errorf("failed to create Workload Identity Pool %s: %w", resourceName, err)
	}
	return nil
}
