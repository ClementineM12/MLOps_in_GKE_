package registry

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createArgoCDServiceAccount(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (*serviceaccount.Account, error) {

	resourceName := fmt.Sprintf("%s-argocd-svc", projectConfig.ResourceNamePrefix)
	serviceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("argocd-svc"),
		DisplayName: pulumi.String("ArgoCD Service Account"),
		Project:     pulumi.String(projectConfig.ProjectId),
	})
	if err != nil {
		return nil, err
	}
	// Assign Artifact Registry Reader Role to the Service Account
	resourceName = fmt.Sprintf("%s-argocd-member-reader", projectConfig.ResourceNamePrefix)
	_, err = projects.NewIAMMember(ctx, resourceName, &projects.IAMMemberArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String("roles/artifactregistry.reader"),
		Member:  pulumi.Sprintf("serviceAccount:%s", serviceAccount.Email),
	}, pulumi.DependsOn([]pulumi.Resource{serviceAccount}))
	if err != nil {
		return nil, err
	}

	return serviceAccount, err
}
