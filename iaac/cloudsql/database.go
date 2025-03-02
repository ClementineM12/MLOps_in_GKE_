package cloudsql

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCloudSQL(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	gcpNetwork *compute.Network,
	dependencies []pulumi.Resource,
) (*sql.DatabaseInstance, error) {

	projectNamePrefix := projectConfig.ResourceNamePrefix
	databaseInstancePrefix := projectConfig.CloudSQL.InstancePrefixName

	resourceName := fmt.Sprintf("%s-%s-db-instance", projectNamePrefix, databaseInstancePrefix)
	dbInstance, err := sql.NewDatabaseInstance(ctx, resourceName, &sql.DatabaseInstanceArgs{
		Name:               pulumi.Sprintf(fmt.Sprintf("%s-db-instance", databaseInstancePrefix)),
		DatabaseVersion:    pulumi.String("POSTGRES_14"),
		Project:            pulumi.String(projectConfig.ProjectId),
		Region:             pulumi.String(cloudRegion.Region),
		DeletionProtection: pulumi.Bool(false),
		Settings: &sql.DatabaseInstanceSettingsArgs{
			Tier: pulumi.String("db-custom-1-3840"),
			IpConfiguration: &sql.DatabaseInstanceSettingsIpConfigurationArgs{
				Ipv4Enabled:                             pulumi.Bool(false),
				EnablePrivatePathForGoogleCloudServices: pulumi.Bool(true),
				PrivateNetwork:                          gcpNetwork.SelfLink,
			},
		},
	}, pulumi.DependsOn(dependencies))
	if err != nil {
		return nil, err
	}

	database, err := createDatabase(ctx, projectConfig, dbInstance)
	if err != nil {
		return nil, err
	}
	randomPassword, err := createUser(ctx, projectConfig, dbInstance, database)
	if err != nil {
		return nil, err
	}

	projectConfig.CloudSQL.InstanceName = dbInstance.Name
	projectConfig.CloudSQL.Connection = dbInstance.FirstIpAddress
	projectConfig.CloudSQL.Password = randomPassword.Result

	return dbInstance, nil
}

func createDatabase(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	dbInstance *sql.DatabaseInstance,
) (*sql.Database, error) {

	projectNamePrefix := projectConfig.ResourceNamePrefix
	databaseInstancePrefix := projectConfig.CloudSQL.InstancePrefixName

	resourceName := fmt.Sprintf("%s-%s-db", projectNamePrefix, databaseInstancePrefix)
	return sql.NewDatabase(ctx, resourceName, &sql.DatabaseArgs{
		Instance: dbInstance.ID(),
		Name:     pulumi.String(projectConfig.CloudSQL.Database),
	})
}

func createUser(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	dbInstance *sql.DatabaseInstance,
	database *sql.Database,
) (*random.RandomPassword, error) {

	username := projectConfig.CloudSQL.User
	projectNamePrefix := projectConfig.ResourceNamePrefix
	databaseInstancePrefix := projectConfig.CloudSQL.InstancePrefixName

	// Generate a random password for the user
	resourceName := fmt.Sprintf("%s-%s-db-%s-user-password", projectNamePrefix, databaseInstancePrefix, username)
	randomPassword, err := random.NewRandomPassword(ctx, resourceName, &random.RandomPasswordArgs{
		Length:  pulumi.Int(16),
		Special: pulumi.Bool(false),
	})
	if err != nil {
		return nil, err
	}
	// Create the user
	resourceName = fmt.Sprintf("%s-%s-db-%s-user", projectNamePrefix, databaseInstancePrefix, username)
	if _, err = sql.NewUser(ctx, resourceName, &sql.UserArgs{
		Instance: dbInstance.ID(),
		Name:     pulumi.String(username),
		Password: randomPassword.Result,
	}, pulumi.DependsOn([]pulumi.Resource{database, dbInstance})); err != nil {
		return nil, err
	}
	return randomPassword, nil
}
