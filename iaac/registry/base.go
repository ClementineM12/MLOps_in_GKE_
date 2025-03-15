package registry

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateArtifactRegistry sets up an Artifact Registry and Workload Identity Federation for GitHub Actions
func CreateArtifactRegistry(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
	opts ...pulumi.ResourceOption,
) (*artifactregistry.Repository, error) {

	artifactRegistry = global.ConfigureArtifactRegistry(ctx, artifactRegistry)
	registry, err := createRegistry(ctx, projectConfig, artifactRegistry, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Artifact Registry: %w", err)
	}

	registry.ID().ApplyT(func(_ string) error {
		if artifactRegistry.GithubServiceAccountCreate {
			// Create a Workload Identity Pool
			randomString := global.GenerateRandomString(4)
			wifPool, err := createWorkloadIdentityPool(ctx, projectConfig, artifactRegistry, randomString)
			if err != nil {
				return fmt.Errorf("failed to create Workload Identity Pool: %w", err)
			}
			wifProvider, err := createWorkloadIdentityPoolProvider(ctx, projectConfig, artifactRegistry, wifPool)
			if err != nil {
				return fmt.Errorf("failed to create Workload Identity Provider: %w", err)
			}

			// Create a Service Account
			githubServiceAccount, serviceAccountMember, err := createGithubServiceAccount(ctx, projectConfig, artifactRegistry)
			if err != nil {
				return fmt.Errorf("failed to create Service Account: %w", err)
			}
			err = createGithubServiceAccountIAMBinding(ctx, projectConfig, artifactRegistry, githubServiceAccount.ID(), wifPool)
			if err != nil {
				return fmt.Errorf("failed to bind IAM role to Service Account: %w", err)
			}
			err = createRegistryIAMMember(ctx, projectConfig, artifactRegistry, githubServiceAccount, serviceAccountMember)
			if err != nil {
				return fmt.Errorf("failed to assign Artifact Registry writer role: %w", err)
			}

			ctx.Export("workloadIdentityProvider", wifProvider.Name)
			ctx.Export("githubServiceAccountEmail", githubServiceAccount.Email)
		}

		if artifactRegistry.ContinuousDevelopmentServiceAccountCreate {
			cdServiveAccount, err := createRegistryServiceAccount(ctx, projectConfig, artifactRegistry)
			if err != nil {
				return err
			}
			ctx.Export("cdServiveAccountEmail", cdServiveAccount.Email)
		}
		ctx.Export("artifactRegistryURL", registry.RepositoryId)
		return nil
	})
	return registry, nil
}
