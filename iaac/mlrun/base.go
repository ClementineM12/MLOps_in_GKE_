package mlrun

import (
	"fmt"
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
	application      = "mlrun"
	domainPrefix     = "mlrun"
	namespace        = "mlrun"
	deploy           = true
	helmChart        = "mlrun-ce"
	helmChartVersion = "0.7.3"
	helmChartRepo    = "https://mlrun.github.io/ce"
)

func CreateMLRunResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
) error {
	registryEndpoint := fmt.Sprintf("%s-docker.pkg.dev", cloudRegion.Region)
	registryURL := fmt.Sprintf("%s/%s/%s", registryEndpoint, projectConfig.ProjectId, registryName)
	domain := fmt.Sprintf("%s.%s", domainPrefix, projectConfig.Domain)

	infraComponents := infracomponents.InfraComponents{
		CertManager:       true,
		NginxIngress:      true,
		CertManagerIssuer: true,
		Certificate:       true,
		Domain:            domain,
	}

	registry, err := registry.CreateArtifactRegistry(ctx, projectConfig, registryName, pulumi.DependsOn([]pulumi.Resource{}))
	if err != nil {
		return err
	}
	serviceAccounts, err := iam.CreateIAMResources(ctx, projectConfig, MLRunIAM)
	if err != nil {
		return err
	}
	if err := createDockerRegistrySecret(ctx, projectConfig, serviceAccounts, registry, registryEndpoint, k8sProvider); err != nil {
		return err
	}

	gcsBucket := storage.CreateObjectStorage(ctx, projectConfig, bucketName)
	dependencies, err := createKubernetesResources(ctx, projectConfig, infraComponents, k8sProvider)
	if err != nil {
		return err
	}

	dependencies = append(dependencies, gcsBucket)
	MLRunConfig := MLRunConfig{
		registryEndpoint:   registryEndpoint,
		registryURL:        registryURL,
		gcsBucketName:      bucketName,
		registrySecretName: registrySecretName,
		domain:             domain,
	}
	if deploy {
		if err = deployMLRun(ctx, projectConfig, k8sProvider, MLRunConfig, dependencies); err != nil {
			return err
		}
		deployIngress(ctx, projectConfig)
	}

	return nil
}

func deployMLRun(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	MLRunConfig MLRunConfig,
	dependencies []pulumi.Resource,
) error {

	// Path to the values.yaml file.
	valuesFilePath := "../helm/mlrun/values/values.yaml"
	// Build the replacement map using resolved strings.
	userSettings := map[string]interface{}{
		"gcsbucket":          MLRunConfig.gcsBucketName,
		"hostName":           MLRunConfig.domain,
		"registryURL":        MLRunConfig.registryURL,
		"registrySecretName": MLRunConfig.registrySecretName,
		"whitelistedIPs":     projectConfig.WhitelistedIPs,
		"minioRootPassword":  "minio123",
	}

	// Get the substituted values map.
	valuesMap, err := global.GetValues(valuesFilePath, userSettings)
	if err != nil {
		return err
	}

	// Deploy the Helm release for Flyte-Core.
	resourceName := fmt.Sprintf("%s-mlrun", projectConfig.ResourceNamePrefix)
	_, err = helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:      pulumi.String(application),
		Namespace: pulumi.String(namespace),
		Version:   pulumi.String(helmChartVersion),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(helmChartRepo),
		},
		Chart:   pulumi.String(helmChart),
		Values:  valuesMap,
		Timeout: pulumi.Int(400),
	},
		pulumi.DependsOn(dependencies),
		pulumi.Provider(k8sProvider),
	)
	if err != nil {
		return fmt.Errorf("failed to deploy MLRun Helm chart: %w", err)
	}
	return nil
}
