package argocd

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DeployArgoCD deploys Argo CD into the cluster using its Helm chart.
// It takes a Pulumi context and an existing Kubernetes provider, then returns
// a pointer to the Helm chart resource or an error.
func DeployArgoCD(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) (*helm.Chart, error) {
	chart, err := helm.NewChart(ctx, "argo-cd", helm.ChartArgs{
		Chart:   pulumi.String("argo-cd"),
		Version: pulumi.String("7.7.21"),
		FetchArgs: helm.FetchArgs{
			Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
		},
		Namespace: pulumi.String("argocd"),
		// Optional: add custom values to configure the deployment.
		// Values: pulumi.Map{
		//     "configs": pulumi.Map{
		//         "someKey": pulumi.String("someValue"),
		//     },
		// },
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return nil, err
	}
	return chart, nil
}
