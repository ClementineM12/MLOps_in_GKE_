package istio

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DeployIstio installs the Istio Service Mesh using Helm on a Kubernetes cluster.
// It installs the Istio base components, followed by the Istiod service, and returns the corresponding Helm releases for both.
// This function manages the setup of Istio in a Kubernetes cluster using Helm charts and dependencies.
func DeployIstio(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	cloudRegion project.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpGKENodePool *container.NodePool,
	gcpBackendService *compute.BackendService,
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
	// Deploy Istio Ingress Gateway into the GKE Clusters
	err = createIstioIngressGateway(ctx, projectConfig, cloudRegion, k8sProvider, helmIstioBase, helmIstioD, gcpGKENodePool, gcpBackendService)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Ingress Gateway for Istio Service Mesh: %w", err)
	}
	return helmIstioBase, helmIstioD, nil
}
