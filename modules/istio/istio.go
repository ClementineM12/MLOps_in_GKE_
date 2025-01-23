package istio

import (
	"fmt"
	"mlops/vpc"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func InstallIstioHelm(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion vpc.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpGKENodePool *container.NodePool,
) (*helm.Release, *helm.Release, error) {
	helmIstioBase, err := createIstioBase(ctx, resourceNamePrefix, cloudRegion, k8sProvider)
	if err != nil {
		return nil, nil, err
	}
	helmIstioD, err := createIstiod(ctx, resourceNamePrefix, cloudRegion, k8sProvider, helmIstioBase, gcpGKENodePool)
	if err != nil {
		return nil, nil, err
	}
	return helmIstioBase, helmIstioD, nil
}

// createIstioBase installs the Istio Service Mesh Base
func createIstioBase(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion vpc.CloudRegion,
	k8sProvider *kubernetes.Provider,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-istio-base-%s", resourceNamePrefix, cloudRegion.Region)
	helmIstioBase, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Description: pulumi.String("Istio Service Mesh - Install IstioBase"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://istio-release.storage.googleapis.com/charts"),
		},
		Chart:           pulumi.String("base"),
		Namespace:       pulumi.String("istio-system"),
		CleanupOnFail:   pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			"defaultRevision": pulumi.String("default"),
		},
	}, pulumi.Provider(k8sProvider))

	return helmIstioBase, err
}

// createIstiod installs the Istio Service Mesh Istiod
func createIstiod(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion vpc.CloudRegion,
	k8sProvider *kubernetes.Provider,
	helmIstioBase *helm.Release,
	gcpGKENodePool *container.NodePool,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-istio-istiod-%s", resourceNamePrefix, cloudRegion.Region)
	helmIstioD, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Description: pulumi.String("Istio Service Mesh - Install Istiod"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://istio-release.storage.googleapis.com/charts"),
		},
		Chart:           pulumi.String("istiod"),
		Namespace:       pulumi.String("istio-system"),
		CleanupOnFail:   pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{helmIstioBase}), pulumi.Parent(gcpGKENodePool))

	return helmIstioD, err
}

func CreateIstioIngressGateway(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion vpc.CloudRegion,
	k8sProvider *kubernetes.Provider,
	helmIstioBase *helm.Release,
	helmIstioD *helm.Release,
	gcpGKENodePool *container.NodePool,
	k8sAppNamespace *k8s.Namespace,
	gcpBackendService *compute.BackendService,
) error {
	resourceName := fmt.Sprintf("%s-istio-igw-%s", resourceNamePrefix, cloudRegion.Region)
	_, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:        pulumi.String("istio-ingressgateway"),
		Description: pulumi.String("Istio Service Mesh - Install Ingress Gateway"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://istio-release.storage.googleapis.com/charts"),
		},
		Chart:         pulumi.String("gateway"),
		Namespace:     k8sAppNamespace.Metadata.Name(),
		CleanupOnFail: pulumi.Bool(true),
		Values: pulumi.Map{
			"service": pulumi.Map{
				//"type": pulumi.String("LoadBalancer"),
				"type": pulumi.String("ClusterIP"),
				"annotations": pulumi.Map{
					"cloud.google.com/neg":                 pulumi.String("{\"exposed_ports\": {\"80\":{}}}"),
					"controller.autoneg.dev/neg":           pulumi.Sprintf("{\"backend_services\":{\"80\":[{\"name\":\"%s\",\"max_rate_per_endpoint\":100}]}}", gcpBackendService.Name),
					"networking.gke.io/load-balancer-type": pulumi.String("Internal"),
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{helmIstioBase, helmIstioD}), pulumi.Parent(gcpGKENodePool))
	return err
}
