package project

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// EnableGCPServices enables a list of GCP services for a given project.
func EnableGCPServices(
	ctx *pulumi.Context,
	projectConfig ProjectConfig,
) []pulumi.Resource {
	var gcpDependencies []pulumi.Resource

	for _, service := range gcpServices[projectConfig.Target] {
		resourceName := fmt.Sprintf("%s-project-service-%s", projectConfig.ResourceNamePrefix, service)

		// Create the service resource
		gcpService, err := projects.NewService(ctx, resourceName, &projects.ServiceArgs{
			DisableDependentServices: pulumi.Bool(true),
			Project:                  pulumi.String(projectConfig.ProjectId),
			Service:                  pulumi.String(service),
			DisableOnDestroy:         pulumi.Bool(false),
		})

		if err != nil {
			ctx.Log.Error(fmt.Sprintf("failed to enable service %s: %s", service, err), nil)
		}
		gcpDependencies = append(gcpDependencies, gcpService)
	}
	return gcpDependencies
}
