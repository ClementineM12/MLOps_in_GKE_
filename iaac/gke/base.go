package gke

import (
	"fmt"
	"mlops/global"

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

	gcpServiceAccount, serviceAccountMember, err := gkeConfigConnectorIAM(ctx, projectConfig)
	if err != nil {
		return nil, nil, err
	}
	gcpGKECluster, k8sProvider, err := createGKE(ctx, config, projectConfig, cloudRegion, gcpNetwork, gcpSubnetwork)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GKE: %w", err)
	}
	gcpGKENodePool, err := createGKENodePool(ctx, config, projectConfig, cloudRegion, gcpGKECluster.ID(), gcpServiceAccount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GKE Node Pool: %w", err)
	}
	if projectConfig.Target == "management" {
		namespaceName := "config-connector"

		gcpGKENodePool.ID().ApplyT(func(_ string) error {
			err := createNamespace(ctx, projectConfig, namespaceName, k8sProvider)
			if err != nil {
				return fmt.Errorf("failed to create Config Connector namespace: %w", err)
			}
			err = applyResource(ctx, projectConfig, serviceAccountMember, k8sProvider)
			if err != nil {
				return fmt.Errorf("failed to apply Config Connector configuration: %w", err)
			}
			return nil
		})
	}
	return k8sProvider, gcpGKENodePool, err
}
