package gke

import (
	"fmt"
	"mlops/vpc"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateGKE
func CreateGKE(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	cloudRegion *vpc.CloudRegion,
	gcpNetwork *compute.Network,
	gcpSubnetwork *compute.Subnetwork,
	gcpServiceAccount *serviceaccount.Account,
) (*container.NodePool, *kubernetes.Provider, error) {

	gcpGKECluster, err := createGKE(ctx, resourceNamePrefix, gcpProjectId, cloudRegion, gcpNetwork, gcpSubnetwork)
	if err != nil {
		return nil, nil, err
	}
	gcpGKENodePool, err := createGKENodePool(ctx, resourceNamePrefix, cloudRegion, gcpGKECluster, gcpServiceAccount)
	if err != nil {
		return nil, nil, err
	}
	k8sProvider, err := createKubernetesProvider(ctx, cloudRegion.GKEClusterName, gcpGKECluster)
	if err != nil {
		return nil, nil, err
	}
	return gcpGKENodePool, k8sProvider, nil
}

// createGKE creates a Google Kubernetes Engine Cluster
func createGKE(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	cloudRegion *vpc.CloudRegion,
	gcpNetwork *compute.Network,
	gcpSubnetwork *compute.Subnetwork,
) (*container.Cluster, error) {

	cloudRegion.GKEClusterName = fmt.Sprintf("%s-gke-%s", resourceNamePrefix, cloudRegion.Region)
	gcpGKECluster, err := container.NewCluster(ctx, cloudRegion.GKEClusterName, &container.ClusterArgs{
		Project:               pulumi.String(gcpProjectId),
		Name:                  pulumi.String(cloudRegion.GKEClusterName),
		Network:               gcpNetwork.ID(),
		Subnetwork:            gcpSubnetwork.ID(),
		Location:              pulumi.String(cloudRegion.Region),
		RemoveDefaultNodePool: pulumi.Bool(true),
		InitialNodeCount:      pulumi.Int(1),
		VerticalPodAutoscaling: &container.ClusterVerticalPodAutoscalingArgs{
			Enabled: pulumi.Bool(true),
		},
		IpAllocationPolicy: &container.ClusterIpAllocationPolicyArgs{},
		MasterAuthorizedNetworksConfig: &container.ClusterMasterAuthorizedNetworksConfigArgs{
			CidrBlocks: &container.ClusterMasterAuthorizedNetworksConfigCidrBlockArray{
				&container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
					CidrBlock:   pulumi.String("0.0.0.0/0"),
					DisplayName: pulumi.String("Global Public Access"),
				},
			},
		},
		WorkloadIdentityConfig: &container.ClusterWorkloadIdentityConfigArgs{
			WorkloadPool: pulumi.String(fmt.Sprintf("%s.svc.id.goog", gcpProjectId)),
		},
	}, pulumi.IgnoreChanges([]string{"gatewayApiConfig"}))

	return gcpGKECluster, err
}

// createGKENodePool
func createGKENodePool(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	cloudRegion *vpc.CloudRegion,
	gcpGKECluster *container.Cluster,
	gcpServiceAccount *serviceaccount.Account,
) (*container.NodePool, error) {

	resourceName := fmt.Sprintf("%s-gke-%s-np-01", resourceNamePrefix, cloudRegion.Region)
	gcpGKENodePool, err := container.NewNodePool(ctx, resourceName, &container.NodePoolArgs{
		Cluster:   gcpGKECluster.ID(),
		Name:      pulumi.String(resourceName),
		NodeCount: pulumi.Int(1),
		NodeConfig: &container.NodePoolNodeConfigArgs{
			Preemptible:    pulumi.Bool(false),
			MachineType:    pulumi.String("e2-medium"),
			ServiceAccount: gcpServiceAccount.Email,
			OauthScopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/cloud-platform"),
			},
		},
		Autoscaling: &container.NodePoolAutoscalingArgs{
			LocationPolicy: pulumi.String("BALANCED"),
			MaxNodeCount:   pulumi.Int(5),
			MinNodeCount:   pulumi.Int(1),
		},
	})
	return gcpGKENodePool, err
}

// CreateKubernetesProvider
func createKubernetesProvider(ctx *pulumi.Context, clusterName string, gkeCluster *container.Cluster) (*kubernetes.Provider, error) {
	resourceName := fmt.Sprintf("%s-kubeconfig", clusterName)
	kubeconfig := generateKubeconfig(gkeCluster.Endpoint, gkeCluster.Name, gkeCluster.MasterAuth)
	return kubernetes.NewProvider(ctx, resourceName, &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
	}, pulumi.DependsOn([]pulumi.Resource{gkeCluster}))
}

// generateKubeconfig generates a KubeConfig that will be used by Pulumi Kubernetes
func generateKubeconfig(clusterEndpoint pulumi.StringOutput, clusterName pulumi.StringOutput,
	clusterMasterAuth container.ClusterMasterAuthOutput) pulumi.StringOutput {
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
