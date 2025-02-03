package iam

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/container"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
)

// Deploy Cluster Ops components for GKE AutoNeg
func autoNegDeployClusterOps(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpGKENodePool *container.NodePool,
	helmIstioBase *helm.Release,
	helmIstioD *helm.Release,
) (*helm.Chart, error) {

	resourceName := fmt.Sprintf("%s-cluster-ops-%s", projectConfig.ResourceNamePrefix, cloudRegion.Id)
	helmClusterOps, err := helm.NewChart(ctx, resourceName, helm.ChartArgs{
		Chart:          pulumi.String("cluster-ops"),
		ResourcePrefix: cloudRegion.Id,
		Version:        pulumi.String("0.1.0"),
		Path:           pulumi.String("../apps/helm"),
		Values: pulumi.Map{
			"global": pulumi.Map{
				"labels": pulumi.Map{
					"region": pulumi.String(cloudRegion.Id),
				},
			},
			"app": pulumi.Map{
				"region": pulumi.String(cloudRegion.Id),
			},
			"autoneg": pulumi.Map{
				"serviceAccount": pulumi.Map{
					"annotations": pulumi.Map{
						"iam.gke.io/gcp-service-account": pulumi.String(fmt.Sprintf("autoneg-system@%s.iam.gserviceaccount.com", projectConfig.ProjectId)),
					}},
			},
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{helmIstioBase, helmIstioD}),
		pulumi.Parent(gcpGKENodePool),
	)
	if err != nil {
		return nil, err
	}
	return helmClusterOps, nil
}

// autoNegServiceAccountBind binds the Kubernetes AutoNeg Service Account to Workload Identity
func autoNegServiceAccountBind(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	cloudRegion *project.CloudRegion,
	k8sProvider *kubernetes.Provider,
	gcpGKENodePool *container.NodePool,
	helmClusterOps *helm.Chart,
	gcpServiceAccountAutoNeg pulumi.StringInput,
) error {

	resourceName := fmt.Sprintf("%s-iam-svc-k8s-%s", projectConfig.ResourceNamePrefix, cloudRegion.Id)
	_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
		ServiceAccountId: gcpServiceAccountAutoNeg,
		Role:             pulumi.String("roles/iam.workloadIdentityUser"),
		Members: pulumi.StringArray{
			pulumi.String(fmt.Sprintf("serviceAccount:%s.svc.id.goog[autoneg-system/autoneg-controller-manager]", projectConfig.ProjectId)),
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{gcpGKENodePool, helmClusterOps}),
		pulumi.Parent(gcpGKENodePool),
	)

	return err
}
