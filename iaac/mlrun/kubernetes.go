package mlrun

import (
	"fmt"
	"mlops/global"
	infracomponents "mlops/infra_components"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createKubernetesResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	infraComponents infracomponents.InfraComponents,
	k8sProvider *kubernetes.Provider,
) ([]pulumi.Resource, error) {

	dependencies := []pulumi.Resource{}
	_, err := createMLRunNamespace(ctx, projectConfig, k8sProvider)
	if err != nil {
		return dependencies, err
	}
	dependencies, err = infracomponents.CreateInfraComponents(ctx, projectConfig, namespace, k8sProvider, infraComponents)
	if err != nil {
		return dependencies, err
	}

	return dependencies, nil
}

func createMLRunNamespace(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*coreV1.Namespace, error) {

	resourceName := fmt.Sprintf("%s-mlrun-ns", projectConfig.ResourceNamePrefix)
	return coreV1.NewNamespace(ctx, resourceName, &coreV1.NamespaceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(namespace),
		},
	}, pulumi.Provider(k8sProvider))
}

// createGcpCredentialsSecret creates a generic Kubernetes secret
// that holds the GCP service account key JSON (used by Kaniko for repository creation).
// func createGcpCredentialsSecret(
// 	ctx *pulumi.Context,
// 	serviceAccountKey pulumi.StringOutput,
// 	dependsOn []pulumi.Resource,
// ) error {
// 	_, err := coreV1.NewSecret(ctx, serviceAccountSecretName, &coreV1.SecretArgs{
// 		Metadata: &metaV1.ObjectMetaArgs{
// 			Namespace: pulumi.String(namespace),
// 			Name:      pulumi.String(serviceAccountSecretName),
// 		},
// 		// The key "credentials" holds the JSON content.
// 		StringData: pulumi.StringMap{
// 			"credentials": serviceAccountKey,
// 		},
// 	}, pulumi.DependsOn(dependsOn))
// 	return err
// }
