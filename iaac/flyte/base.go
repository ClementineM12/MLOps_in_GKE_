package flyte

import (
	"fmt"
	"mlops/cloudsql"
	"mlops/global"
	"mlops/iam"
	infracomponents "mlops/infra_components"
	"mlops/storage"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	application      = "flyte"
	namespace        = "flyte"
	deploy           = true
	helmChart        = "flyte-core"
	helmChartVersion = "v1.15.0"
	helmChartRepo    = "https://flyteorg.github.io/flyte"
)

func CreateFlyteResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
) error {
	domain := fmt.Sprintf("%s.%s", application, projectConfig.Domain)
	infraComponents := infracomponents.InfraComponents{
		CertManager:  true,
		NginxIngress: true,
		Domain:       domain,
	}
	serviceAccounts, err := iam.CreateIAMResources(ctx, projectConfig, FlyteIAM)
	if err != nil {
		return err
	}
	gcsBucket := storage.CreateObjectStorage(ctx, projectConfig, "flyte-project-bucket-01")
	cloudSQL, err := cloudsql.DeployCloudSQL(ctx, projectConfig, cloudRegion, gcpNetwork)
	if err != nil {
		return err
	}
	dependencies, err := createKubernetesResources(ctx, projectConfig, infraComponents, k8sProvider, cloudSQL)
	if err != nil {
		return err
	}
	if deploy {
		if err = deployFlyteCore(ctx, projectConfig, k8sProvider, gcsBucket.Name, domain, serviceAccounts, dependencies); err != nil {
			return err
		}
	}
	configureSAIAMPolicy(ctx, projectConfig, serviceAccounts)

	return nil
}

func deployFlyteCore(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	gcsBucket pulumi.StringInput,
	domain string,
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
			"hostName":                  domain,
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
			Name:      pulumi.String(application),
			Namespace: pulumi.String(namespace),
			Version:   pulumi.String(helmChartVersion),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String(helmChartRepo),
			},
			Chart:  pulumi.String(helmChart),
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
