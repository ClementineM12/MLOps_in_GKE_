package flyte

import (
	"mlops/global"
	infracomponents "mlops/infra_components"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createKubernetesResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	cloudSQL *sql.DatabaseInstance,
) ([]pulumi.Resource, error) {

	dependencies, err := infracomponents.CreateInfraComponents(ctx, projectConfig, k8sProvider, infraComponents)
	if err != nil {
		return []pulumi.Resource{}, err
	}

	dependsOn := append(dependencies,
		cloudSQL,
	)

	return dependsOn, nil
}
