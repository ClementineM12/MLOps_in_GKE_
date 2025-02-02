package registry

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createGithubServiceAccount(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	resourceName := fmt.Sprintf("%s-github-svc", projectConfig.ResourceNamePrefix)

	// Create a Service Account
	serviceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("github-svc"),
		DisplayName: pulumi.String("GitHub Actions Service Account"),
		Project:     pulumi.String(projectConfig.ProjectId),
	})

	serviceAccountMember := serviceAccount.Email.ApplyT(func(email string) []string {
		return []string{fmt.Sprintf("serviceAccount:%s", email)}
	}).(pulumi.StringArrayOutput)

	return serviceAccount, serviceAccountMember, err
}

func createGithubServiceAccountIAMBinding(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	serviceAccount pulumi.StringInput,
	wifPool pulumi.StringInput,
	githubRepo string,
) error {

	resourceName := fmt.Sprintf("%s-github-svc-wip-binding", projectConfig.ResourceNamePrefix)

	// Allow the Service Account to be used via Workload Identity Federation
	_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
		ServiceAccountId: serviceAccount,
		// Role:             pulumi.String("roles/iam.workloadIdentityPoolAdmin"),
		Members: pulumi.StringArray{
			pulumi.Sprintf("principalSet://iam.googleapis.com/%s/attribute.repository/%s", wifPool, githubRepo),
		},
	})
	return err
}

func createRegistryIAMMember(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	serviceAccount *serviceaccount.Account,
	serviceAccountMember pulumi.StringArrayOutput,
) error {

	resourceName := fmt.Sprintf("%s-helm-artifacts-registry-member-writer", projectConfig.ResourceNamePrefix)
	member := serviceAccountMember.Index(pulumi.Int(0))

	// Give the service account permissions to push to Artifact Registry
	_, err := projects.NewIAMMember(ctx, resourceName, &projects.IAMMemberArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String("roles/artifactregistry.writer"),
		Member:  member,
	}, pulumi.DependsOn([]pulumi.Resource{serviceAccount}))

	return err
}
