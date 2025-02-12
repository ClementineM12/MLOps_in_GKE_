package database

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// var flyte_ksas = ["default"] // The KSA that Task Pods will use
//  As recommended by Google, we make use of Workload Identity as a mechanism to enable KSAs
//   to impersonate GSAs and access GCP resources. To do so, the following process is implemented in this module:
//  1. Create a GSA. This is the Principal that will be impersonated by the KSAs
//  2. Create the custom roles that include the permissions for flyteadmin, the dataplane (flytepropeller) and the workers (the Task Pods)
//  3. Grant the custom role to each GSA
//  4. Define an IAM binding at the SA level to associate the GSA with the KSA as a Workload Identity user

func createCloudSQL(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	region global.CloudRegion,
	privateNetwork 
) error {

	resourceName := fmt.Sprintf("%s-db-instance", projectConfig.ResourceNamePrefix)
	dbInstance, err := sql.NewDatabaseInstance(ctx, resourceName, &sql.DatabaseInstanceArgs{
		Name:               pulumi.Sprintf("DB instance"),
		DatabaseVersion:    pulumi.String("POSTGRES_14"),
		Project:            pulumi.String(projectConfig.ProjectId),
		Region:             pulumi.String(region.Region),
		DeletionProtection: pulumi.Bool(false),
		Settings: &sql.DatabaseInstanceSettingsArgs{
			Tier: pulumi.String("db-custom-1-3840"),
			IpConfiguration: &sql.DatabaseInstanceSettingsIpConfigurationArgs{
				Ipv4Enabled:                             pulumi.Bool(false),
				EnablePrivatePathForGoogleCloudServices: pulumi.Bool(true),
				PrivateNetwork:                          pulumi.String("your-private-network-self-link"),
			},
		},
	})

	// Create the database
	resourceName = fmt.Sprintf("%s-db", projectConfig.ResourceNamePrefix)
	_, err = sql.NewDatabase(ctx, resourceName, &sql.DatabaseArgs{
		Instance: dbInstance.Name,
		Name:     pulumi.String("mlop"),
	})
	if err != nil {
		return err
	}
	// Generate a random password for the user
	randomPassword, err := random.NewRandomPassword(ctx, "mlopUserPassword", &random.RandomPasswordArgs{
		Length:  pulumi.Int(16),
		Special: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	// Create the user
	_, err = sql.NewUser(ctx, "mlop-user", &sql.UserArgs{
		Instance: dbInstance.Name,
		Name:     pulumi.String("mlop"),
		Password: randomPassword.Result,
	})
	if err != nil {
		return err
	}
	// Output the database instance connection name
	ctx.Export("dbInstanceConnectionName", dbInstance.ConnectionName)
	ctx.Export("flyteUserPassword", randomPassword.Result)
	return nil
}
