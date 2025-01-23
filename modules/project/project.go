package project

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ProjectOutputs defines the outputs of the module
type ProjectOutputs struct {
	ProjectID pulumi.StringOutput
}

// SetupProject creates a new GCP project and returns its outputs
func SetupProject(
	ctx *pulumi.Context,
	orgID string,
	billingAccount string,
) (ProjectOutputs, error) {

	// Define a unique project ID
	projectID := "my-kubeflow-project-" + ctx.Stack()

	// Create a new GCP project
	project, err := organizations.NewProject(ctx, "gcp-project", &organizations.ProjectArgs{
		Name:           pulumi.String("My Kubeflow Project"),
		ProjectId:      pulumi.String(projectID),
		OrgId:          pulumi.String(orgID),
		BillingAccount: pulumi.String(billingAccount),
	})
	if err != nil {
		return ProjectOutputs{}, err
	}

	return ProjectOutputs{
		ProjectID: project.ProjectId,
	}, nil
}

// CreateWorkloadIdentityPool creates Google Cloud Workload Identity Pool for GKE
func CreateWorkloadIdentityPool(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
) error {

	// Generate a unique suffix for the Workload Identity Pool ID
	randomId, err := random.NewRandomString(ctx, fmt.Sprintf("%s-id", resourceNamePrefix), &random.RandomStringArgs{
		Length:  pulumi.Int(6),
		Special: pulumi.Bool(false),
		Upper:   pulumi.Bool(false),
		Lower:   pulumi.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to generate random suffix for Workload Identity Pool ID: %w", err)
	}

	// Construct the resource name
	resourceName := fmt.Sprintf("%s-wip-gke-cluster", resourceNamePrefix)

	// Create the Workload Identity Pool
	_, err = iam.NewWorkloadIdentityPool(ctx, resourceName, &iam.WorkloadIdentityPoolArgs{
		Project:                pulumi.String(gcpProjectId),
		Description:            pulumi.String("GKE - Workload Identity Pool for GKE Cluster"),
		Disabled:               pulumi.Bool(false),
		DisplayName:            pulumi.String(resourceName),
		WorkloadIdentityPoolId: pulumi.String(fmt.Sprintf("%s-%s", resourceNamePrefix, randomId.ID())),
	})
	if err != nil {
		return fmt.Errorf("failed to create Workload Identity Pool %s: %w", resourceName, err)
	}

	ctx.Log.Info(fmt.Sprintf("Successfully created Workload Identity Pool: %s", resourceName), nil)

	return nil
}
