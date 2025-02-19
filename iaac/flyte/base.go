package flyte

import (
	"fmt"
	"io/ioutil"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v2"
)

var (
	application = "flyte"
)

func CreateFlyteResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpNetwork *compute.Network,
	gcpSubnetwork *compute.Subnetwork,
	gcsBucket pulumi.StringInput,
) error {

	serviceNetworking, err := createServiceNetworking(ctx, gcpNetwork)
	if err != nil {
		return err
	}
	serviceNetworkConnection, err := createServiceNetworkingConnection(ctx, gcpNetwork, serviceNetworking)
	if err != nil {
		return err
	}
	serviceAccounts, err := createFlyteIAM(ctx, projectConfig)
	if err != nil {
		return err
	}
	dependencies, err := createKubernetesResources(ctx, projectConfig, k8sProvider)
	if err != nil {
		return err
	}
	cloudSQL, err := createCloudSQL(ctx, projectConfig, cloudRegion, gcpSubnetwork, serviceNetworkConnection)
	if err != nil {
		return err
	}

	deployFlyteCore(ctx, projectConfig, k8sProvider, cloudSQL, gcsBucket, serviceAccounts, dependencies)
	return nil
}

func deployFlyteCore(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	cloudSQL struct {
		InstanceName pulumi.StringOutput
		Connection   pulumi.StringOutput
		Password     pulumi.StringOutput
	},
	gcsBucket pulumi.StringInput,
	serviceAccounts map[string]struct {
		AccountId pulumi.StringOutput
		Member    pulumi.StringArrayOutput
	},
	dependsOn []pulumi.Resource,
) error {

	// Load the values.yaml file
	valuesFilePath := "../../helm/flyte/values/values.yaml"
	yamlData, err := ioutil.ReadFile(valuesFilePath)
	if err != nil {
		return fmt.Errorf("failed to read values.yaml: %w", err)
	}

	// Convert YAML file to a map
	var valuesMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &valuesMap)
	if err != nil {
		return fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	// Inject dynamic values into the parsed YAML map
	userSettings := map[string]interface{}{
		"googleProjectId":              projectConfig.ProjectId,
		"dbHost":                       cloudSQL.Connection,
		"dbPassword":                   cloudSQL.Password,
		"bucketName":                   gcsBucket,
		"hostName":                     fmt.Sprintf("%s.%s", application, projectConfig.Domain),
		"flyteadminServiceAccount":     serviceAccounts["flyteadmin"],
		"flytepropellerServiceAccount": serviceAccounts["flytepropeller"],
		"flyteschedulerServiceAccount": serviceAccounts["flytescheduler"],
		"datacatalogServiceAccount":    serviceAccounts["datacatalog"],
		"flyteworkersServiceAccount":   serviceAccounts["flyteworkers"],
	}
	valuesMap["userSettings"] = userSettings

	// Deploy Helm release for Flyte-Core
	resourceName := fmt.Sprintf("%s-flyte-core", projectConfig.ResourceNamePrefix)
	_, err = helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:      pulumi.String("flyte-core"),
		Namespace: pulumi.String("flyte"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://flyteorg.github.io/flyte"),
		},
		Chart:  pulumi.String("flyte-core"),
		Values: pulumi.ToMap(valuesMap),
	},
		pulumi.DependsOn(dependsOn),
		pulumi.Provider(k8sProvider),
	)

	if err != nil {
		return fmt.Errorf("failed to deploy Flyte-Core Helm chart: %w", err)
	}

	return nil
}
