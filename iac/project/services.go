package project

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var GCPServices = []string{
	"compute.googleapis.com",
	"container.googleapis.com",
}

var ManagementGCPServices = []string{
	"serviceusage.googleapis.com",
	"compute.googleapis.com",
	"container.googleapis.com",
	"iam.googleapis.com",
	"servicemanagement.googleapis.com",
	"cloudresourcemanager.googleapis.com",
	"ml.googleapis.com",
	"iap.googleapis.com",
	"sqladmin.googleapis.com",
	"meshconfig.googleapis.com",
	"krmapihosting.googleapis.com",
	"servicecontrol.googleapis.com",
	"endpoints.googleapis.com",
	"cloudbuild.googleapis.c",
}

// EnableGCPServices enables a list of GCP services for a given project.
func EnableGCPServices(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
) ([]pulumi.Resource, error) {

	var errorList []error
	var gcpDependencies []pulumi.Resource

	for _, service := range GCPServices {
		resourceName := fmt.Sprintf("%s-project-service-%s", resourceNamePrefix, service)

		ctx.Log.Info(fmt.Sprintf("Enabling GCP service: %s", service), nil)

		// Create the service resource
		gcpService, err := projects.NewService(ctx, resourceName, &projects.ServiceArgs{
			DisableDependentServices: pulumi.Bool(true),           // Disable dependent services if the parent service is disabled
			Project:                  pulumi.String(gcpProjectId), // Specify the project ID
			Service:                  pulumi.String(service),      // The service to enable
			DisableOnDestroy:         pulumi.Bool(false),          // Keep the service enabled on destroy
		})

		if err != nil {
			errorList = append(errorList, fmt.Errorf("failed to enable service %s: %w", service, err))
			continue
		}

		// Append successful service enablement to dependencies
		gcpDependencies = append(gcpDependencies, gcpService)
	}

	if len(errorList) > 0 {
		for _, e := range errorList {
			ctx.Log.Error(e.Error(), nil)
		}
		return gcpDependencies, fmt.Errorf("one or more services failed to enable")
	}

	return gcpDependencies, nil
}
