package flyte

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/servicenetworking"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCloudSQL(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	gcpSubnetwork *compute.Subnetwork,
	serviceNetworkConnection *servicenetworking.Connection,
) (struct {
	InstanceName pulumi.StringOutput
	Connection   pulumi.StringOutput
	Password     pulumi.StringOutput
}, error) {

	resourceName := fmt.Sprintf("%s-db-instance", projectConfig.ResourceNamePrefix)
	dbInstance, err := sql.NewDatabaseInstance(ctx, resourceName, &sql.DatabaseInstanceArgs{
		Name:               pulumi.Sprintf("DB instance"),
		DatabaseVersion:    pulumi.String("POSTGRES_14"),
		Project:            pulumi.String(projectConfig.ProjectId),
		Region:             pulumi.String(cloudRegion.Region),
		DeletionProtection: pulumi.Bool(false),
		Settings: &sql.DatabaseInstanceSettingsArgs{
			Tier: pulumi.String("db-custom-1-3840"),
			IpConfiguration: &sql.DatabaseInstanceSettingsIpConfigurationArgs{
				Ipv4Enabled:                             pulumi.Bool(false),
				EnablePrivatePathForGoogleCloudServices: pulumi.Bool(true),
				PrivateNetwork:                          gcpSubnetwork.SelfLink,
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{serviceNetworkConnection}))
	if err != nil {
		return struct {
			InstanceName pulumi.StringOutput
			Connection   pulumi.StringOutput
			Password     pulumi.StringOutput
		}{}, err
	}

	// Create the database
	resourceName = fmt.Sprintf("%s-db", projectConfig.ResourceNamePrefix)
	_, err = sql.NewDatabase(ctx, resourceName, &sql.DatabaseArgs{
		Instance: dbInstance.Name,
		Name:     pulumi.String("mlop"),
	})
	if err != nil {
		return struct {
			InstanceName pulumi.StringOutput
			Connection   pulumi.StringOutput
			Password     pulumi.StringOutput
		}{}, err
	}
	// Generate a random password for the user
	randomPassword, err := random.NewRandomPassword(ctx, "mlopUserPassword", &random.RandomPasswordArgs{
		Length:  pulumi.Int(16),
		Special: pulumi.Bool(false),
	})
	if err != nil {
		return struct {
			InstanceName pulumi.StringOutput
			Connection   pulumi.StringOutput
			Password     pulumi.StringOutput
		}{}, err
	}
	// Create the user
	_, err = sql.NewUser(ctx, "mlop-user", &sql.UserArgs{
		Instance: dbInstance.Name,
		Name:     pulumi.String("mlop"),
		Password: randomPassword.Result,
	})
	if err != nil {
		return struct {
			InstanceName pulumi.StringOutput
			Connection   pulumi.StringOutput
			Password     pulumi.StringOutput
		}{}, err
	}

	// Output the database instance connection name
	ctx.Export("dbInstanceConnectionName", dbInstance.ConnectionName)

	return struct {
		InstanceName pulumi.StringOutput
		Connection   pulumi.StringOutput
		Password     pulumi.StringOutput
	}{
		InstanceName: dbInstance.Name,
		Connection:   dbInstance.ConnectionName,
		Password:     randomPassword.Result,
	}, nil
}
