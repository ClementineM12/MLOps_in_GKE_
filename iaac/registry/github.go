package registry

// Package registry provides functions to create and configure a GitHub Service Account,
// set up Workload Identity Federation (WIF) bindings, and assign IAM roles for
// accessing Google Cloud services, such as Artifact Registry.

import (
	"fmt"
	"mlops/global"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createGithubServiceAccount(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	formattedName := strings.Title(strings.ReplaceAll(artifactRegistry.RegistryName, "-", " "))

	resourceName := fmt.Sprintf("%s-%s-github-svc", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
	serviceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("github-svc"),
		DisplayName: pulumi.String(fmt.Sprintf("%s GitHub Actions", formattedName)),
		Project:     pulumi.String(projectConfig.ProjectId),
	})

	serviceAccountMember := serviceAccount.Email.ApplyT(func(email string) []string {
		return []string{fmt.Sprintf("serviceAccount:%s", email)}
	}).(pulumi.StringArrayOutput)

	return serviceAccount, serviceAccountMember, err
}

func createGithubServiceAccountIAMBinding(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
	serviceAccount pulumi.StringInput,
	wifPool *iam.WorkloadIdentityPool,
) error {

	resourceName := fmt.Sprintf("%s-%s-github-svc-wip-binding", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
	// Allow the Service Account to be used via Workload Identity Federation
	_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
		ServiceAccountId: serviceAccount,
		Role:             pulumi.String("roles/iam.workloadIdentityUser"),
		Members: pulumi.StringArray{
			pulumi.Sprintf("principalSet://iam.googleapis.com/%s/attribute.repository/%s", wifPool.Name, artifactRegistry.GithubRepo),
		},
	}, pulumi.DependsOn([]pulumi.Resource{wifPool}))

	return err
}

func createRegistryIAMMember(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
	serviceAccount *serviceaccount.Account,
	serviceAccountMember pulumi.StringArrayOutput,
) error {

	resourceName := fmt.Sprintf("%s-%s-registry-member-writer", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
	member := serviceAccountMember.Index(pulumi.Int(0))

	// Give the service account permissions to push to Artifact Registry
	_, err := projects.NewIAMMember(ctx, resourceName, &projects.IAMMemberArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String("roles/artifactregistry.writer"),
		Member:  member,
	}, pulumi.DependsOn([]pulumi.Resource{serviceAccount}))

	return err
}

func createWorkloadIdentityPool(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
	randomString string,
) (*iam.WorkloadIdentityPool, error) {

	formattedName := strings.Title(strings.ReplaceAll(artifactRegistry.RegistryName, "-", " "))

	resourceName := fmt.Sprintf("%s-%s-github-wip", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
	wifPool, err := iam.NewWorkloadIdentityPool(ctx, resourceName, &iam.WorkloadIdentityPoolArgs{
		Project:                pulumi.String(projectConfig.ProjectId),
		Description:            pulumi.String("Github - Workload Identity Pool"),
		Disabled:               pulumi.Bool(false),
		DisplayName:            pulumi.String(fmt.Sprintf("%s GitHub", formattedName)),
		WorkloadIdentityPoolId: pulumi.String(fmt.Sprintf("%s-%s-github-pool-%s", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName, randomString)),
	})
	if err != nil {
		return nil, err
	}
	return wifPool, nil
}

func createWorkloadIdentityPoolProvider(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
	wifPool *iam.WorkloadIdentityPool,
) (*iam.WorkloadIdentityPoolProvider, error) {

	formattedName := strings.Title(strings.ReplaceAll(artifactRegistry.RegistryName, "-", " "))

	resourceName := fmt.Sprintf("%s-%s-github-wip-provider", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
	wifProvider, err := iam.NewWorkloadIdentityPoolProvider(ctx, resourceName, &iam.WorkloadIdentityPoolProviderArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		// Setting this to wifPool.ID() causes error with double ID
		WorkloadIdentityPoolId:         wifPool.WorkloadIdentityPoolId,
		WorkloadIdentityPoolProviderId: pulumi.String("github-wip-provider"),
		DisplayName:                    pulumi.String(fmt.Sprintf("%s GitHub", formattedName)),
		AttributeMapping: pulumi.StringMap{
			"attribute.actor":      pulumi.String("assertion.actor"),
			"google.subject":       pulumi.String("assertion.sub"),
			"attribute.repository": pulumi.String("assertion.repository"),
		},
		AttributeCondition: pulumi.String(fmt.Sprintf(`attribute.repository=="%s"`, artifactRegistry.GithubRepo)),
		Oidc: &iam.WorkloadIdentityPoolProviderOidcArgs{
			IssuerUri: pulumi.String("https://token.actions.githubusercontent.com"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{wifPool}))
	if err != nil {
		return nil, err
	}
	return wifProvider, nil
}
