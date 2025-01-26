package gke

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
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
	gcpNetwork *compute.Network,
	gcpSubnetwork *compute.Subnetwork,
	gcpServiceAccount *serviceaccount.Account,
) (*container.NodePool, *kubernetes.Provider, error) {

	config := Configuration(ctx)

	gcpGKECluster, err := createGKE(ctx, config, projectConfig, cloudRegion, gcpNetwork, gcpSubnetwork)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GKE: %w", err)
	}
	gcpGKENodePool, err := createGKENodePool(ctx, config, projectConfig, cloudRegion, gcpGKECluster, gcpServiceAccount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GKE Node Pool: %w", err)
	}
	k8sProvider, err := createKubernetesProvider(ctx, cloudRegion.GKEClusterName, gcpGKECluster)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes Provider configuration: %w", err)
	}
	return gcpGKENodePool, k8sProvider, nil
}

// createGKE sets up the Google Kubernetes Engine (GKE) cluster in the specified region using the provided network, subnetwork, and project details.
// The function configures various cluster settings such as authorized networks, workload identity, and vertical pod autoscaling.
// It returns the created GKE cluster object, which is used by subsequent resources such as node pools and Kubernetes providers.
func createGKE(
	ctx *pulumi.Context,
	config *ClusterConfig,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	gcpNetwork *compute.Network,
	gcpSubnetwork *compute.Subnetwork,
) (*container.Cluster, error) {

	cloudRegion.GKEClusterName = fmt.Sprintf("%s-%s-k8s-%s", projectConfig.ResourceNamePrefix, config.Name, cloudRegion.Region)
	gcpGKECluster, err := container.NewCluster(ctx, cloudRegion.GKEClusterName, &container.ClusterArgs{
		Project:               pulumi.String(projectConfig.ProjectId),
		Name:                  pulumi.String(cloudRegion.GKEClusterName),
		Network:               gcpNetwork.ID(),
		Subnetwork:            gcpSubnetwork.ID(),
		Location:              pulumi.String(cloudRegion.Region), // Since we are providing a region, the cluster will be regional
		RemoveDefaultNodePool: pulumi.Bool(true),
		InitialNodeCount:      pulumi.Int(1),
		VerticalPodAutoscaling: &container.ClusterVerticalPodAutoscalingArgs{
			Enabled: pulumi.Bool(true),
		},
		IpAllocationPolicy: &container.ClusterIpAllocationPolicyArgs{},
		MasterAuthorizedNetworksConfig: &container.ClusterMasterAuthorizedNetworksConfigArgs{
			CidrBlocks: &container.ClusterMasterAuthorizedNetworksConfigCidrBlockArray{
				&container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
					CidrBlock:   pulumi.String(config.Cidr),
					DisplayName: pulumi.String("Global Public Access"),
				},
			},
		},
		WorkloadIdentityConfig: &container.ClusterWorkloadIdentityConfigArgs{
			WorkloadPool: pulumi.String(fmt.Sprintf("%s.svc.id.goog", projectConfig.ProjectId)),
		},
	}, pulumi.IgnoreChanges([]string{"gatewayApiConfig"}))

	// Export the Cluster name and endpoint
	ctx.Export("clusterName", gcpGKECluster.Name)
	ctx.Export("clusterEndpoint", gcpGKECluster.Endpoint)

	return gcpGKECluster, err
}

// createGKENodePool creates a node pool within the specified GKE cluster. It configures the node pool with settings such as machine type, preemptibility,
// and service account credentials. The node pool also supports autoscaling with the specified minimum and maximum node count.
// This function returns the created node pool object, which is used to manage the nodes within the GKE cluster.
func createGKENodePool(
	ctx *pulumi.Context,
	config *ClusterConfig,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	gcpGKECluster *container.Cluster,
	gcpServiceAccount *serviceaccount.Account,
) (*container.NodePool, error) {

	resourceName := fmt.Sprintf("%s-%s-k8s-%s-np", projectConfig.ResourceNamePrefix, config.Name, cloudRegion.Region)
	gcpGKENodePool, err := container.NewNodePool(ctx, resourceName, &container.NodePoolArgs{
		Cluster:   gcpGKECluster.ID(),
		Name:      pulumi.String(resourceName),
		NodeCount: pulumi.Int(1),
		NodeConfig: &container.NodePoolNodeConfigArgs{
			Preemptible:    pulumi.Bool(config.NodePool.Preemptible),
			MachineType:    pulumi.String(config.NodePool.MachineType),
			ServiceAccount: gcpServiceAccount.Email,
			OauthScopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/devstorage.read_only"),
				pulumi.String("https://www.googleapis.com/auth/logging.write"),
				pulumi.String("https://www.googleapis.com/auth/monitoring"),
				pulumi.String("https://www.googleapis.com/auth/trace.append"),
			},
			DiskType:   pulumi.String(config.NodePool.DiskType),
			DiskSizeGb: pulumi.Int(config.NodePool.DiskSizeGb),
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
	kubeconfig := generateKubeconfig(gkeCluster.Endpoint, gkeCluster.Name, gkeCluster.MasterAuth)
	return kubernetes.NewProvider(ctx, resourceName, &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
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

	context := pulumi.Sprintf("%s", clusterName)

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
