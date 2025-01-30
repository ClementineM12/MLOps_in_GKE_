package istio

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	IstioBaseChart = "base"
	IstiodChart    = "istiod"
)

// InstallIstioHelm installs the Istio Service Mesh using Helm on a Kubernetes cluster.
// It installs the Istio base components, followed by the Istiod service, and returns the corresponding Helm releases for both.
// This function manages the setup of Istio in a Kubernetes cluster using Helm charts and dependencies.
func InstallIstioHelm(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	cloudRegion project.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpGKENodePool *container.NodePool,
) (*helm.Release, *helm.Release, error) {
	resourceNamePrefix := projectConfig.ResourceNamePrefix
	helmIstioBase, err := createIstioBase(ctx, resourceNamePrefix, cloudRegion, k8sProvider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Istio Service Mesh Base: %w", err)
	}
	helmIstioD, err := createIstiod(ctx, resourceNamePrefix, cloudRegion, k8sProvider, helmIstioBase, gcpGKENodePool)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Istio Base: %w", err)
	}
	return helmIstioBase, helmIstioD, nil
}

// createIstioBase installs the base components of the Istio Service Mesh using Helm.
// It pulls the Istio base chart from a public Helm repository and deploys it into the "istio-system" namespace.
// This function ensures that the fundamental Istio components are installed before other Istio features (like Istiod) can be set up.
func createIstioBase(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion project.CloudRegion,
	k8sProvider *kubernetes.Provider,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-istio-base-%s", resourceNamePrefix, cloudRegion.Region)
	helmIstioBase, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Description: pulumi.String("Istio Service Mesh - Install IstioBase"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://istio-release.storage.googleapis.com/charts"),
		},
		Chart:           pulumi.String(IstioBaseChart),
		Namespace:       pulumi.String("istio-system"),
		CleanupOnFail:   pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			"defaultRevision": pulumi.String("default"),
		},
	}, pulumi.Provider(k8sProvider))

	return helmIstioBase, err
}

// createIstiod installs the Istiod component of the Istio Service Mesh using Helm.
// Istiod is the central control plane component of Istio and handles management tasks for Istio proxies and resources.
// This function deploys Istiod into the "istio-system" namespace and ensures it runs after the Istio base components are installed.
func createIstiod(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion project.CloudRegion,
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
		Chart:           pulumi.String(IstiodChart),
		Namespace:       pulumi.String("istio-system"),
		CleanupOnFail:   pulumi.Bool(true),
		CreateNamespace: pulumi.Bool(true),
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{helmIstioBase}), pulumi.Parent(gcpGKENodePool))

	return helmIstioD, err
}

// CreateIstioIngressGateway installs the Istio Ingress Gateway using Helm on the Kubernetes cluster.
// The Ingress Gateway provides external access to services within the Istio service mesh.
// This function configures the Ingress Gateway with specific service annotations for load balancing in a GKE environment,
// such as enabling Network Endpoint Groups (NEG) and specifying an internal load balancer type.
func CreateIstioIngressGateway(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	cloudRegion project.CloudRegion,
	k8sProvider *kubernetes.Provider,
	helmIstioBase *helm.Release,
	helmIstioD *helm.Release,
	gcpGKENodePool *container.NodePool,
	gcpBackendService *compute.BackendService,
) error {
	resourceName := fmt.Sprintf("%s-istio-igw-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
	_, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:        pulumi.String("istio-ingressgateway"),
		Description: pulumi.String("Istio Service Mesh - Install Ingress Gateway"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://istio-release.storage.googleapis.com/charts"),
		},
		Chart:         pulumi.String("gateway"),
		Namespace:     pulumi.String("default"),
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
