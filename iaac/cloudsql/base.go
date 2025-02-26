package cloudsql

import (
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployCloudSQL(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	gcpNetwork *compute.Network,
) (*sql.DatabaseInstance, error) {

	dependencies, err := createServiceNetworking(ctx, projectConfig, gcpNetwork)
	if err != nil {
		return nil, err
	}
	cloudSQL, err := createCloudSQL(ctx, projectConfig, cloudRegion, gcpNetwork, dependencies)
	if err != nil {
		return nil, err
	}

	return cloudSQL, nil
}
