package gke

import (
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateGKE creates a Google Kubernetes Engine (GKE) cluster along with its associated node pool and Kubernetes provider for managing the cluster.
// This function sets up the cluster, initializes the node pool, and creates a Kubernetes provider using the generated kubeconfig.
// The Kubernetes provider is used to interact with the created GKE cluster.
func CreateGKEResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	gcpNetwork pulumi.StringInput,
	gcpSubnetwork pulumi.StringInput,
) (*kubernetes.Provider, *container.NodePool, error) {

	config := Configuration(ctx)

	serviceAccount, err := iam.CreateIAMResources(ctx, projectConfig, AdministrationIAM)
	if err != nil {
		return nil, nil, err
	}
	gcpGKECluster, k8sProvider, err := createGKE(ctx, config, projectConfig, cloudRegion, gcpNetwork, gcpSubnetwork)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GKE: %w", err)
	}
	GKENodePools, err := createGKENodePool(ctx, config, projectConfig, cloudRegion, gcpGKECluster.ID(), serviceAccount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GKE Node Pool: %w", err)
	}

	baseNodePool, exists := GKENodePools["base"]
	if !exists {
		return nil, nil, fmt.Errorf("base node pool not found")
	}
	return k8sProvider, baseNodePool, err
}
