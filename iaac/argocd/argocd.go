package argocd

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	argocdNamespace        = "argocd"
	argocdHelmChartVersion = "7.7.21"
	argocdValuesPath       = "argocd/values/values.yaml"
	argocdHelm             = "https://argoproj.github.io/argo-helm"
)

func createNamespace(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*corev1.Namespace, error) {

	resourceName := fmt.Sprintf("%s-argocd-namespace", projectConfig.ResourceNamePrefix)
	namespace, err := corev1.NewNamespace(ctx, resourceName, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(argocdNamespace),
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace: %s", err)
	}
	return namespace, nil
}

// DeployArgoCD deploys Argo CD into the cluster using its Helm chart.
// It takes a Pulumi context and an existing Kubernetes provider, then returns
// a pointer to the Helm chart resource or an error.
func DeployArgoCD(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) error {

	namespace, err := createNamespace(ctx, projectConfig, k8sProvider)
	if err != nil {
		return err
	}
	_, err = getValues(argocdValuesPath)
	if err != nil {
		return err
	}

	resourceName := fmt.Sprintf("%s-argocd", projectConfig.ResourceNamePrefix)
	_, err = helm.NewChart(ctx, resourceName, helm.ChartArgs{
		Chart:   pulumi.String("argo-cd"),
		Version: pulumi.String(argocdHelmChartVersion),
		FetchArgs: helm.FetchArgs{
			Repo: pulumi.String(argocdHelm),
		},
		// Values:    values,
		Namespace: namespace.ID(),
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{k8sProvider}),
	)

	if err != nil {
		return fmt.Errorf("failed to deploy ArgoCD end-to-end: %s", err)
	}
	return err
}
