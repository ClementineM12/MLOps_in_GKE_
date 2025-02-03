package registry

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// CreateArtifactRegistry sets up an Artifact Registry and Workload Identity Federation for GitHub Actions
func CreateArtifactRegistry(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	opts ...pulumi.ResourceOption,
) error {

	githubRepo := config.Get(ctx, "ar:githubRepo")

	registry, err := createRegistry(ctx, projectConfig, opts...)
	if err != nil {
		return fmt.Errorf("failed to create Artifact Registry: %w", err)
	}

	registry.ID().ApplyT(func(_ string) error {
		// Create a Workload Identity Pool
		wifPool, err := createWorkloadIdentityPool(ctx, projectConfig)
		if err != nil {
			return fmt.Errorf("failed to create Workload Identity Pool: %w", err)
		}
		wifProvider, err := createWorkloadIdentityPoolProvider(ctx, projectConfig, wifPool, githubRepo)
		if err != nil {
			return fmt.Errorf("failed to create Workload Identity Provider: %w", err)
		}

		// Create a Service Account
		serviceAccount, serviceAccountMember, err := createGithubServiceAccount(ctx, projectConfig)
		if err != nil {
			return fmt.Errorf("failed to create Service Account: %w", err)
		}
		err = createGithubServiceAccountIAMBinding(ctx, projectConfig, serviceAccount.ID(), wifPool, githubRepo)
		if err != nil {
			return fmt.Errorf("failed to bind IAM role to Service Account: %w", err)
		}
		err = createRegistryIAMMember(ctx, projectConfig, serviceAccount, serviceAccountMember)
		if err != nil {
			return fmt.Errorf("failed to assign Artifact Registry writer role: %w", err)
		}

		// Output required values for GitHub Actions
		ctx.Export("artifactRegistryURL", registry.RepositoryId)
		ctx.Export("workloadIdentityProvider", wifProvider.Name)
		ctx.Export("serviceAccountEmail", serviceAccount.Email)

		return nil
	})
	return nil
}
