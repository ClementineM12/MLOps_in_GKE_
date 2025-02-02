package gke

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateGKE creates a Google Kubernetes Engine (GKE) cluster along with its associated node pool and Kubernetes provider for managing the cluster.
// This function sets up the cluster, initializes the node pool, and creates a Kubernetes provider using the generated kubeconfig.
// The Kubernetes provider is used to interact with the created GKE cluster.
func CreateGKE(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	gcpNetwork pulumi.StringInput,
	gcpSubnetwork pulumi.StringInput,
) (*kubernetes.Provider, error) {

	config := Configuration(ctx)

	gcpServiceAccount, serviceAccountMember, err := gkeConfigConnectorIAM(ctx, projectConfig)
	if err != nil {
		return nil, err
	}
	gcpGKECluster, k8sProvider, err := createGKE(ctx, config, projectConfig, cloudRegion, gcpNetwork, gcpSubnetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to create GKE: %w", err)
	}
	gcpGKENodePool, err := createGKENodePool(ctx, config, projectConfig, cloudRegion, gcpGKECluster.ID(), gcpServiceAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create GKE Node Pool: %w", err)
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
	return k8sProvider, err
}

// createGKE sets up the Google Kubernetes Engine (GKE) cluster in the specified region using the provided network, subnetwork, and project details.
// The function configures various cluster settings such as authorized networks, workload identity, and vertical pod autoscaling.
// It returns the created GKE cluster object, which is used by subsequent resources such as node pools and Kubernetes providers.
func createGKE(
	ctx *pulumi.Context,
	config *ClusterConfig,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	gcpNetwork pulumi.StringInput,
	gcpSubnetwork pulumi.StringInput,
) (*container.Cluster, *kubernetes.Provider, error) {

	cloudRegion.GKEClusterName = fmt.Sprintf("%s-gke-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
	gcpGKECluster, err := container.NewCluster(ctx, cloudRegion.GKEClusterName, &container.ClusterArgs{
		Project:               pulumi.String(projectConfig.ProjectId),
		Name:                  pulumi.String(cloudRegion.GKEClusterName),
		Network:               gcpNetwork,
		Subnetwork:            gcpSubnetwork,
		Location:              pulumi.String(cloudRegion.Region), // Since we are providing a region, the cluster will be regional
		RemoveDefaultNodePool: pulumi.Bool(true),
		InitialNodeCount:      pulumi.Int(1),
		MinMasterVersion:      pulumi.String(config.NodePool.MinMasterVersion),
		VerticalPodAutoscaling: &container.ClusterVerticalPodAutoscalingArgs{
			Enabled: pulumi.Bool(true),
		},
		ReleaseChannel: &container.ClusterReleaseChannelArgs{
			Channel: pulumi.String("REGULAR"),
		},
		ResourceLabels: config.NodePool.ResourceLabels,
		// MasterAuthorizedNetworksConfig: &container.ClusterMasterAuthorizedNetworksConfigArgs{
		// 	CidrBlocks: &container.ClusterMasterAuthorizedNetworksConfigCidrBlockArray{
		// 		&container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
		// 			CidrBlock:   pulumi.String(config.Cidr),
		// 			DisplayName: pulumi.String("Global Public Access"),
		// 		},
		// 	},
		// },
		WorkloadIdentityConfig: &container.ClusterWorkloadIdentityConfigArgs{
			WorkloadPool: pulumi.String(fmt.Sprintf("%s.svc.id.goog", projectConfig.ProjectId)),
		},
		AddonsConfig: &container.ClusterAddonsConfigArgs{
			ConfigConnectorConfig: &container.ClusterAddonsConfigConfigConnectorConfigArgs{
				Enabled: pulumi.Bool(projectConfig.Target == "management"),
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes Cluster: %w", err)
	}
	ctx.Export("kubeconfig", generateKubeconfig(gcpGKECluster.Endpoint, gcpGKECluster.Name, gcpGKECluster.MasterAuth))
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
	config *ClusterConfig,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	clusterID pulumi.StringInput,
	gcpServiceAccount *serviceaccount.Account,
) (*container.NodePool, error) {

	resourceName := fmt.Sprintf("%s-gke-%s-np", projectConfig.ResourceNamePrefix, cloudRegion.Region)
	gcpGKENodePool, err := container.NewNodePool(ctx, resourceName, &container.NodePoolArgs{
		Cluster:   clusterID,
		Name:      pulumi.String(resourceName),
		NodeCount: pulumi.Int(1),
		NodeConfig: &container.NodePoolNodeConfigArgs{
			Metadata:       config.NodePool.Metadata,
			Preemptible:    pulumi.Bool(config.NodePool.Preemptible),
			MachineType:    pulumi.String(config.NodePool.MachineType),
			OauthScopes:    config.NodePool.OauthScopes,
			Labels:         config.NodePool.ResourceLabels,
			DiskType:       pulumi.String(config.NodePool.DiskType),
			DiskSizeGb:     pulumi.Int(config.NodePool.DiskSizeGb),
			ServiceAccount: gcpServiceAccount.Email,
			KubeletConfig: &container.NodePoolNodeConfigKubeletConfigArgs{
				CpuCfsQuota:       pulumi.Bool(false),
				CpuCfsQuotaPeriod: pulumi.String(""),
				CpuManagerPolicy:  pulumi.String(""),
				PodPidsLimit:      pulumi.Int(1024),
			},
			// Add Workload Metadata Config for Workload Identity
			// Note: Enabling Workload Identity on an existing cluster does not automatically enable Workload Identity on the cluster's existing node pools.
			//           We recommend that you enable Workload Identity on all your cluster's node pools since Config Connector could run on any of them.
			//           https://cloud.google.com/config-connector/docs/how-to/install-upgrade-uninstall#prerequisites
			WorkloadMetadataConfig: config.NodePool.WorkloadMetadataConfig,
		},
		Autoscaling: &container.NodePoolAutoscalingArgs{
			LocationPolicy: pulumi.String("BALANCED"),
			MaxNodeCount:   pulumi.Int(config.NodePool.MaxNodeCount),
			MinNodeCount:   pulumi.Int(config.NodePool.MinNodeCount),
		},
		Management: &container.NodePoolManagementArgs{
			AutoRepair:  pulumi.Bool(config.Management.AutoRepair),
			AutoUpgrade: pulumi.Bool(config.Management.AutoUpgrade),
		},
	})

	return gcpGKENodePool, err
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
