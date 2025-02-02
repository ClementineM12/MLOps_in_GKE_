package registry

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/iam"
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
	// Create an Artifact Registry
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

func createRegistry(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	opts ...pulumi.ResourceOption,
) (*artifactregistry.Repository, error) {

	var region string
	if len(project.CloudRegions) > 1 {
		region = projectConfig.EnabledRegions[0].Region
	} else {
		region = "europe"
	}
	resourceName := fmt.Sprintf("%s-helm-artifacts-registry", projectConfig.ResourceNamePrefix)

	// Give the service account permissions to push to Artifact Registry
	registry, err := artifactregistry.NewRepository(ctx, resourceName, &artifactregistry.RepositoryArgs{
		Project:      pulumi.String(projectConfig.ProjectId),
		Location:     pulumi.String(region),
		RepositoryId: pulumi.String("helm-charts"),
		Format:       pulumi.String("DOCKER"), // Artifact Registry supports OCI Helm Charts
	}, opts...)

	fmt.Printf("\033[1;32m[INFO] - Artifact Registry will be created in: [ %s ]\n\033[0m", region)

	return registry, err
}

func createWorkloadIdentityPool(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) (*iam.WorkloadIdentityPool, error) {

	resourceName := fmt.Sprintf("%s-github-wip", projectConfig.ResourceNamePrefix)

	// Create the Workload Identity Pool
	wifPool, err := iam.NewWorkloadIdentityPool(ctx, resourceName, &iam.WorkloadIdentityPoolArgs{
		Project:                pulumi.String(projectConfig.ProjectId),
		Description:            pulumi.String("Github - Workload Identity Pool"),
		Disabled:               pulumi.Bool(false),
		DisplayName:            pulumi.String("GitHub Workload Identity Pool"),
		WorkloadIdentityPoolId: pulumi.String(fmt.Sprintf("%s-github-pool", projectConfig.ResourceNamePrefix)),
	})
	if err != nil {
		return nil, err
	}
	return wifPool, nil
}

func createWorkloadIdentityPoolProvider(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	wifPool *iam.WorkloadIdentityPool,
	githubRepo string,
) (*iam.WorkloadIdentityPoolProvider, error) {

	resourceName := fmt.Sprintf("%s-github-wip-provider", projectConfig.ResourceNamePrefix)

	// Create a Workload Identity Provider for GitHub
	wifProvider, err := iam.NewWorkloadIdentityPoolProvider(ctx, resourceName, &iam.WorkloadIdentityPoolProviderArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		// Setting this to wifPool.ID() causes error with double ID
		WorkloadIdentityPoolId:         wifPool.WorkloadIdentityPoolId,
		WorkloadIdentityPoolProviderId: pulumi.String("github-wip-provider"),
		DisplayName:                    pulumi.String("GitHub Actions Identity Provider"),
		AttributeMapping: pulumi.StringMap{
			"attribute.actor":      pulumi.String("assertion.actor"),
			"google.subject":       pulumi.String("assertion.sub"),
			"attribute.repository": pulumi.String("assertion.repository"),
		},
		AttributeCondition: pulumi.String(fmt.Sprintf(`attribute.repository=="%s"`, githubRepo)),
		Oidc: &iam.WorkloadIdentityPoolProviderOidcArgs{
			IssuerUri: pulumi.String("https://token.actions.githubusercontent.com"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{wifPool}))
	if err != nil {
		return nil, err
	}
	return wifProvider, nil
}
