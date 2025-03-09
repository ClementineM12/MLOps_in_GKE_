package registry

import (
	"fmt"
	"mlops/global"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createRegistryServiceAccount(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	artifactRegistry global.ArtifactRegistryConfig,
) (*serviceaccount.Account, error) {

	formattedName := strings.Title(strings.ReplaceAll(artifactRegistry.RegistryName, "-", " "))

	resourceName := fmt.Sprintf("%s-%s-registry-svc", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
	serviceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("registry-svc"),
		DisplayName: pulumi.String(fmt.Sprintf("%s Registry Service Account", formattedName)),
		Project:     pulumi.String(projectConfig.ProjectId),
	})
	if err != nil {
		return nil, err
	}
	// Assign Artifact Registry Reader Role to the Service Account
	resourceName = fmt.Sprintf("%s-%s-registry-member-reader", projectConfig.ResourceNamePrefix, artifactRegistry.RegistryName)
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
