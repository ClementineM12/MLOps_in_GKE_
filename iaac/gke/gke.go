package gke

import (
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var (
	GKEDeletionProtection    = false // We set this to false since we need to be able to destroy the cluster without interuptions, else the `pulumi destroy` will fail
	GKERemoveDefaultNodePool = true
	GKEReleaseChannel        = "REGULAR"
)

// createGKE sets up the Google Kubernetes Engine (GKE) cluster in the specified region using the provided network, subnetwork, and project details.
// The function configures various cluster settings such as authorized networks, workload identity, and vertical pod autoscaling.
// It returns the created GKE cluster object, which is used by subsequent resources such as node pools and Kubernetes providers.
func createGKE(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	gcpNetwork pulumi.StringInput,
	gcpSubnetwork pulumi.StringInput,
) (*container.Cluster, *kubernetes.Provider, error) {

	privateNodesEnabled := config.GetBool(ctx, "gke:privateNodes")

	privateClusterConfig := &container.ClusterPrivateClusterConfigArgs{}
	if privateNodesEnabled {
		// https://cloud.google.com/kubernetes-engine/docs/how-to/legacy/network-isolation
		privateClusterConfig = &container.ClusterPrivateClusterConfigArgs{
			EnablePrivateNodes:    pulumi.Bool(privateNodesEnabled),
			MasterIpv4CidrBlock:   pulumi.String(cloudRegion.MasterIpv4CidrBlock),
			EnablePrivateEndpoint: pulumi.Bool(false),
		}
	} else {
		privateClusterConfig = &container.ClusterPrivateClusterConfigArgs{
			EnablePrivateNodes: pulumi.Bool(privateNodesEnabled),
		}
	}
	cloudRegion.GKEClusterName = fmt.Sprintf("%s-gke-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
	gcpGKECluster, err := container.NewCluster(ctx, cloudRegion.GKEClusterName, &container.ClusterArgs{
		Project:               pulumi.String(projectConfig.ProjectId),
		Name:                  pulumi.String(cloudRegion.GKEClusterName),
		Network:               gcpNetwork,
		Subnetwork:            gcpSubnetwork,
		Location:              pulumi.String(cloudRegion.Region), // Since we are providing a region, the cluster will be regional
		DeletionProtection:    pulumi.Bool(GKEDeletionProtection),
		RemoveDefaultNodePool: pulumi.Bool(GKERemoveDefaultNodePool),
		InitialNodeCount:      pulumi.Int(1),
		// EnableShieldedNodes:   pulumi.Bool(privateNodesEnabled),
		PrivateClusterConfig: privateClusterConfig,
		VerticalPodAutoscaling: &container.ClusterVerticalPodAutoscalingArgs{
			Enabled: pulumi.Bool(true),
		},
		ReleaseChannel: &container.ClusterReleaseChannelArgs{
			Channel: pulumi.String(GKEReleaseChannel),
		},
		WorkloadIdentityConfig: &container.ClusterWorkloadIdentityConfigArgs{
			WorkloadPool: pulumi.String(fmt.Sprintf("%s.svc.id.goog", projectConfig.ProjectId)),
		},
		// HorizontalPodAutoscaling & HttpLoadBalancing are also enabled by default
		AddonsConfig: &container.ClusterAddonsConfigArgs{
			ConfigConnectorConfig: &container.ClusterAddonsConfigConfigConnectorConfigArgs{
				Enabled: pulumi.Bool(projectConfig.Target == "management"),
			},
			GcePersistentDiskCsiDriverConfig: &container.ClusterAddonsConfigGcePersistentDiskCsiDriverConfigArgs{
				Enabled: pulumi.Bool(true),
			},
		},
		LoggingService:    pulumi.String("logging.googleapis.com/kubernetes"),
		MonitoringService: pulumi.String("monitoring.googleapis.com/kubernetes"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes Cluster: %w", err)
	}
	// ctx.Export("kubeconfig", generateKubeconfig(gcpGKECluster.Endpoint, gcpGKECluster.Name, gcpGKECluster.MasterAuth))
	k8sProvider, err := createKubernetesProvider(ctx, cloudRegion.GKEClusterName, gcpGKECluster)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes Provider configuration: %w", err)
	}
	return gcpGKECluster, k8sProvider, nil
}

// createGKENodePool creates a node pool within the specified GKE cluster. It configures the node pool with settings such as machine type, preemptibility,
// and service account credentials. The node pool also supports autoscaling with the specified minimum and maximum node count.
// This function returns the created node pool object, which is used to manage the nodes within the GKE cluster.
func createGKENodePool(
	ctx *pulumi.Context,
	ClusterConfig *ClusterConfig,
	projectConfig global.ProjectConfig,
	cloudRegion *global.CloudRegion,
	clusterID pulumi.StringInput,
	serviceAccount map[string]iam.ServiceAccountInfo,
) (map[string]*container.NodePool, error) {

	// Create a map to hold the created node pools
	nodePools := make(map[string]*container.NodePool)

	for key, nodePool := range ClusterConfig.NodePools {
		resourceName := ""
		if key == "base" {
			resourceName = fmt.Sprintf("%s-gke-%s-np", projectConfig.ResourceNamePrefix, cloudRegion.Region)
		} else {
			resourceName = fmt.Sprintf("%s-gke-%s-%s-np", projectConfig.ResourceNamePrefix, cloudRegion.Region, nodePool.KeyName)
		}

		// Create the node pool using the provided configuration.
		np, err := container.NewNodePool(ctx, resourceName, &container.NodePoolArgs{
			Cluster:          clusterID,
			Name:             pulumi.String(resourceName),
			InitialNodeCount: pulumi.Int(nodePool.InitialNodeCount),
			NodeConfig: &container.NodePoolNodeConfigArgs{
				Metadata:       nodePool.Metadata,
				Preemptible:    pulumi.Bool(nodePool.Preemptible),
				MachineType:    pulumi.String(nodePool.MachineType),
				OauthScopes:    nodePool.OauthScopes,
				Labels:         nodePool.Labels, // Comma added here
				DiskType:       pulumi.String(nodePool.DiskType),
				DiskSizeGb:     pulumi.Int(nodePool.DiskSizeGb),
				ServiceAccount: serviceAccount["admin"].ServiceAccount.Email,
				ResourceLabels: mergeStringMaps(
					pulumi.StringMap{
						"goog-gke-node-pool-provisioning-model": pulumi.String("on-demand"),
					},
					nodePool.ResourceLabels,
				),
				KubeletConfig: &container.NodePoolNodeConfigKubeletConfigArgs{
					CpuCfsQuota:       pulumi.Bool(false),
					CpuCfsQuotaPeriod: pulumi.String(""),
					CpuManagerPolicy:  pulumi.String(""),
					PodPidsLimit:      pulumi.Int(1024),
				},
				WorkloadMetadataConfig: nodePool.WorkloadMetadataConfig,
			},
			Autoscaling: &container.NodePoolAutoscalingArgs{
				LocationPolicy: pulumi.String(nodePool.LocationPolicy),
				MinNodeCount:   pulumi.Int(nodePool.MinNodeCount),
				MaxNodeCount:   pulumi.Int(nodePool.MaxNodeCount),
			},
			Management: &container.NodePoolManagementArgs{
				AutoRepair:  pulumi.Bool(ClusterConfig.Management.AutoRepair),
				AutoUpgrade: pulumi.Bool(ClusterConfig.Management.AutoUpgrade),
			},
		}, pulumi.DependsOn([]pulumi.Resource{serviceAccount["admin"].ServiceAccount}))

		if err != nil {
			return nil, fmt.Errorf("failed to create %s node pool: %w", key, err)
		}

		// Save the created node pool in the map keyed by the node pool's KeyName.
		nodePools[key] = np
	}

	return nodePools, nil
}

// createKubernetesProvider creates a Kubernetes provider using the kubeconfig generated from the GKE cluster's endpoint and authentication credentials.
// The provider is used to manage Kubernetes resources on the created cluster and interacts with the cluster based on its API server details.
// This function is essential for using Pulumi to deploy Kubernetes resources onto the GKE cluster.
func createKubernetesProvider(
	ctx *pulumi.Context,
	clusterName string,
	gkeCluster *container.Cluster,
) (*kubernetes.Provider, error) {

	resourceName := fmt.Sprintf("%s-kubeconfig", clusterName)

	return kubernetes.NewProvider(ctx, resourceName, &kubernetes.ProviderArgs{
		Kubeconfig: generateKubeconfig(gkeCluster.Endpoint, gkeCluster.Name, gkeCluster.MasterAuth),
	}, pulumi.DependsOn([]pulumi.Resource{gkeCluster}))
}

// generateKubeconfig generates a kubeconfig formatted string based on the GKE cluster's endpoint, cluster name, and master authentication credentials.
// This kubeconfig is then used by the Kubernetes provider to authenticate and manage resources in the GKE cluster.
// The kubeconfig also includes information about cluster certificates and authentication mechanisms.
func generateKubeconfig(
	clusterEndpoint pulumi.StringOutput,
	clusterName pulumi.StringOutput,
	clusterMasterAuth container.ClusterMasterAuthOutput,
) pulumi.StringOutput {
	context := pulumi.Sprintf("%s-kubeconfig", clusterName)

	return pulumi.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: gke-gcloud-auth-plugin
      installHint: Install gke-gcloud-auth-plugin for use with kubectl by following
        https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
      provideClusterInfo: true
`,
		clusterMasterAuth.ClusterCaCertificate().Elem(),
		clusterEndpoint, context, context, context, context, context, context)
}
