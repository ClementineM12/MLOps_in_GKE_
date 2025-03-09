package flyte

import (
	"fmt"
	"mlops/cloudsql"
	"mlops/global"
	"mlops/iam"
	infracomponents "mlops/infra_components"
	"mlops/registry"
	"mlops/storage"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	deploy = true

	application      = "flyte"
	namespace        = "flyte"
	helmChart        = "flyte-core"
	helmChartVersion = "v1.15.0"
	helmChartRepo    = "https://flyteorg.github.io/flyte"
	bucketName       = "flyte-project-bucket-01"
	registryName     = "flyte"
)

func CreateFlyteResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
) error {
	// Construct the domain.
	domain := fmt.Sprintf("%s.%s", application, projectConfig.Domain)
	var dependencies []pulumi.Resource

	// Assign the CloudSQL configuration.
	projectConfig.CloudSQL = &cloudSQLConfig

	infraComponents := infracomponents.InfraComponents{
		CertManager:       true,
		NginxIngress:      true,
		Domain:            domain,
		CertManagerIssuer: true,
	}
	artifactRegistryConfig := global.ArtifactRegistryConfig{
		RegistryName:               registryName,
		GithubServiceAccountCreate: true,
	}

	// Create IAM resources.
	serviceAccounts, err := iam.CreateIAMResources(ctx, projectConfig, FlyteIAM)
	if err != nil {
		return err
	}

	// Create the GCS bucket for object storage.
	gcsBucket := storage.CreateObjectStorage(ctx, projectConfig, bucketName)

	_, err = registry.CreateArtifactRegistry(ctx, projectConfig, artifactRegistryConfig, pulumi.DependsOn([]pulumi.Resource{}))
	if err != nil {
		return err
	}
	// Deploy CloudSQL and obtain its dependencies.
	cloudSQL, cloudSQLDependencies, err := cloudsql.DeployCloudSQL(ctx, projectConfig, cloudRegion, gcpNetwork)
	if err != nil {
		return err
	}

	// Create other Kubernetes resources (e.g., ingress, certificates).
	kubernetesDependencies, letsEncrypt, err := createKubernetesResources(ctx, projectConfig, infraComponents, k8sProvider, cloudSQL)
	if err != nil {
		return err
	}

	// If deploying Flyte core, add the dependencies and call the deployment.
	if deploy {
		// Append the Kubernetes and CloudSQL dependencies to our dependencies slice.
		dependencies = append(dependencies, kubernetesDependencies...)
		dependencies = append(dependencies, cloudSQLDependencies...)

		// Deploy Flyte core, ensuring it waits for the dependencies.
		if err := deployFlyteCore(ctx, projectConfig, k8sProvider, gcsBucket.Name, domain, serviceAccounts, letsEncrypt, dependencies); err != nil {
			return err
		}
	}

	// Configure the Service Account IAM policy.
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
	letsEncrypt string,
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
		projectConfig.CloudSQL.DatabaseName,
	).ApplyT(func(vals []interface{}) (interface{}, error) {
		adminEmail := vals[0].(string)
		propellerEmail := vals[1].(string)
		schedulerEmail := vals[2].(string)
		datacatalogEmail := vals[3].(string)
		workersEmail := vals[4].(string)
		gcsBucket := vals[5].(string)
		dbHost := vals[6].(string)
		dbPassword := vals[7].(string)
		dbName := vals[8].(string)

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
			"dbName":                    dbName,
			"dbUsername":                projectConfig.CloudSQL.User,
			"whitelistedIPs":            projectConfig.WhitelistedIPs,
			"LetsEncrypt":               letsEncrypt,
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
