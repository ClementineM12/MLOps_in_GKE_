package flyte

import (
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	application = "flyte"
	deploy      = true
)

func CreateFlyteResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcsBucket pulumi.StringInput,
	cloudSQL *sql.DatabaseInstance,
) error {

	serviceAccounts, err := iam.CreateIAMResources(ctx, projectConfig, FlyteIAM)
	if err != nil {
		return err
	}
	dependencies, err := createKubernetesResources(ctx, projectConfig, k8sProvider, cloudSQL)
	if err != nil {
		return err
	}
	if deploy {
		if err = deployFlyteCore(ctx, projectConfig, k8sProvider, gcsBucket, serviceAccounts, dependencies); err != nil {
			return err
		}
	}
	return nil
}

func deployFlyteCore(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	gcsBucket pulumi.StringInput,
	serviceAccounts map[string]iam.ServiceAccountInfo,
	dependencies []pulumi.Resource,
) error {
	// Wait for all service account emails to resolve.
	// This ensures we have plain string values for our substitutions.
	pulumi.All(
		serviceAccounts["flyteadmin"].Email,
		serviceAccounts["flytepropeller"].Email,
		serviceAccounts["flytescheduler"].Email,
		serviceAccounts["datacatalog"].Email,
		serviceAccounts["flyteworkers"].Email,
		gcsBucket,
		projectConfig.CloudSQL.Connection,
		projectConfig.CloudSQL.Password,
	).ApplyT(func(vals []interface{}) (interface{}, error) {
		adminEmail := vals[0].(string)
		propellerEmail := vals[1].(string)
		schedulerEmail := vals[2].(string)
		datacatalogEmail := vals[3].(string)
		workersEmail := vals[4].(string)
		gcsBucket := vals[5].(string)
		dbHost := vals[6].(string)
		dbPassword := vals[7].(string)

		// Path to the values.yaml file.
		valuesFilePath := "../helm/flyte/values/values.yaml"
		// Build the replacement map using resolved strings.
		userSettings := map[string]interface{}{
			"gcpProjectId":              projectConfig.ProjectId,
			"dbHost":                    dbHost,
			"dbPassword":                dbPassword,
			"gcsbucket":                 gcsBucket,
			"hostName":                  fmt.Sprintf("%s.%s", application, projectConfig.Domain),
			"AdminServiceAccount":       adminEmail,
			"PropellerServiceAccount":   propellerEmail,
			"SchedulerServiceAccount":   schedulerEmail,
			"DatacatalogServiceAccount": datacatalogEmail,
			"WorkersServiceAccount":     workersEmail,
			"whitelistedIPs":            projectConfig.WhitelistedIPs,
		}

		// Get the substituted values map.
		valuesMap, err := global.GetValues(valuesFilePath, userSettings)
		if err != nil {
			return nil, err
		}

		// Deploy the Helm release for Flyte-Core.
		resourceName := fmt.Sprintf("%s-flyte-core", projectConfig.ResourceNamePrefix)
		_, err = helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
			Name:      pulumi.String("flyte-core"),
			Namespace: pulumi.String("flyte"),
			Version:   pulumi.String("v1.15.0"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://flyteorg.github.io/flyte"),
			},
			Chart:  pulumi.String("flyte-core"),
			Values: valuesMap,
		},
			pulumi.DependsOn(dependencies),
			pulumi.Provider(k8sProvider),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy Flyte-Core Helm chart: %w", err)
		}
		return nil, nil
	})
	return nil
}
