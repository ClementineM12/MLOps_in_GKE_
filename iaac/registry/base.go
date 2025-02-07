package registry

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// CreateArtifactRegistry sets up an Artifact Registry and Workload Identity Federation for GitHub Actions
func CreateArtifactRegistry(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	opts ...pulumi.ResourceOption,
) error {

	githubRepo := config.Get(ctx, "ar:githubRepo")

	registry, err := createRegistry(ctx, projectConfig, opts...)
	if err != nil {
		return fmt.Errorf("failed to create Artifact Registry: %w", err)
	}

	registry.ID().ApplyT(func(_ string) error {
		if config.GetBool(ctx, "ar:githubSACreate") {
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
			githubServiceAccount, serviceAccountMember, err := createGithubServiceAccount(ctx, projectConfig)
			if err != nil {
				return fmt.Errorf("failed to create Service Account: %w", err)
			}
			err = createGithubServiceAccountIAMBinding(ctx, projectConfig, githubServiceAccount.ID(), wifPool, githubRepo)
			if err != nil {
				return fmt.Errorf("failed to bind IAM role to Service Account: %w", err)
			}
			err = createRegistryIAMMember(ctx, projectConfig, githubServiceAccount, serviceAccountMember)
			if err != nil {
				return fmt.Errorf("failed to assign Artifact Registry writer role: %w", err)
			}

			ctx.Export("workloadIdentityProvider", wifProvider.Name)
			ctx.Export("githubServiceAccountEmail", githubServiceAccount.Email)
		}

		if config.GetBool(ctx, "ar:argoCDSACreate") {
			argoCDServiveAccount, err := createArgoCDServiceAccount(ctx, projectConfig)
			if err != nil {
				return err
			}
			ctx.Export("argoCDServiveAccountEmail", argoCDServiveAccount.Email)
		}
		ctx.Export("artifactRegistryURL", registry.RepositoryId)
		return nil
	})
	return nil
}
