package registry

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createRegistry gives the Service Account permissions to push to Artifact Registry
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
	registry, err := artifactregistry.NewRepository(ctx, resourceName, &artifactregistry.RepositoryArgs{
		Project:      pulumi.String(projectConfig.ProjectId),
		Location:     pulumi.String(region),
		RepositoryId: pulumi.String("helm-charts"),
		Format:       pulumi.String("DOCKER"), // Artifact Registry supports OCI Helm Charts
	}, opts...)

	fmt.Printf("\033[1;32m[INFO] - Artifact Registry will be created in: [ %s ]\n\033[0m", region)

	return registry, err
}
