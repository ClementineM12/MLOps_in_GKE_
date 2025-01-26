package main

import (
	"mlops/gke"
	gcpIAM "mlops/iam"
	"mlops/istio"
	"mlops/project"
	"mlops/storage"
	"mlops/vpc"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		projectConfig := project.GenerateProjectConfig(ctx)
		project.EnableGCPServices(ctx, projectConfig)

		if config.GetBool(ctx, "storage:create") {
			storage.CreateObjectStorage(ctx, projectConfig)
		}
		// CreateResources(ctx, projectConfig, gcpDependencies)
		return nil
	})
}

func CreateResources(ctx *pulumi.Context, projectConfig project.ProjectConfig, gcpDependencies []pulumi.Resource) error {
	// -------------------------- IAM ------------------------------
	gcpServiceAccount := gcpIAM.CreateServiceAccount(ctx, projectConfig, "Admin")
	// Create AutoNeg Service Account: custom IAM Role that will be used by the AutoNeg Kubernetes Deployment
	gcpServiceAccountAutoNeg := gcpIAM.CreateServiceAccount(ctx, projectConfig, "AutoNEG")

	// -----------------  Workload Identity Pool --------------------
	err := project.CreateWorkloadIdentityPool(ctx, projectConfig)
	if err != nil {
		return err
	}
	// ---------------------------  VPC ----------------------------
	gcpNetwork, err := vpc.CreateVPC(ctx, projectConfig, gcpDependencies)
	if err != nil {
		return nil
	}
	gcpBackendService, err := vpc.CreateLoadBalancerBackendService(ctx, projectConfig, gcpDependencies)
	if err != nil {
		return nil
	}
	gcpGlobalAddress, err := vpc.CreateLoadBalancerStaticIP(ctx, projectConfig, gcpDependencies)
	if err != nil {
		return err
	}
	if projectConfig.SSL {
		err = vpc.ConfigureSSLCertificate(ctx, projectConfig, gcpBackendService, gcpGlobalAddress, gcpDependencies)
		if err != nil {
			return err
		}
	}
	err = vpc.CreateLoadBalancerURLMapHTTP(ctx, projectConfig, gcpGlobalAddress, gcpBackendService)
	if err != nil {
		return nil
	}
	// Process Each Cloud Region;
	for _, cloudRegion := range projectConfig.EnabledRegions {

		// Create VPC Subnet for Cloud Region
		gcpSubnetwork, err := vpc.CreateVPCSubnet(ctx, projectConfig, cloudRegion, gcpNetwork)
		if err != nil {
			return err
		}

		// ---------------------------  GKE ----------------------------
		gcpGKENodePool, k8sProvider, err := gke.CreateGKE(ctx, projectConfig, &cloudRegion, gcpNetwork, gcpSubnetwork, gcpServiceAccount)
		if err != nil {
			return err
		}
		helmIstioBase, helmIstioD, err := istio.InstallIstioHelm(ctx, projectConfig, cloudRegion, k8sProvider, gcpGKENodePool)
		if err != nil {
			return err
		}

		// Create New Namespace in the GKE Clusters for Application Deployments
		resourceName := fmt.Sprintf("%s-k8s-ns-app-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
		k8sAppNamespace, err := k8s.NewNamespace(ctx, resourceName, &k8s.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String("app-team"),
				Labels: pulumi.StringMap{
					"istio-injection": pulumi.String("enabled"),
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{helmIstioD}))
		if err != nil {
			return err
		}

		// Deploy Istio Ingress Gateway into the GKE Clusters
		err = istio.CreateIstioIngressGateway(ctx, projectConfig, cloudRegion, k8sProvider, helmIstioBase, helmIstioD, gcpGKENodePool, k8sAppNamespace, gcpBackendService)
		if err != nil {
			return err
		}

		// Deploy Cluster Ops components for GKE AutoNeg
		resourceName = fmt.Sprintf("%s-cluster-ops-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
		helmClusterOps, err := helm.NewChart(ctx, resourceName, helm.ChartArgs{
			Chart:          pulumi.String("cluster-ops"),
			ResourcePrefix: cloudRegion.Id,
			Version:        pulumi.String("0.1.0"),
			Path:           pulumi.String("../apps/helm"),
			Values: pulumi.Map{
				"global": pulumi.Map{
					"labels": pulumi.Map{
						"region": pulumi.String(cloudRegion.Region),
					},
				},
				"app": pulumi.Map{
					"region": pulumi.String(cloudRegion.Region),
				},
				"autoneg": pulumi.Map{
					"serviceAccount": pulumi.Map{
						"annotations": pulumi.Map{
							"iam.gke.io/gcp-service-account": pulumi.String(fmt.Sprintf("autoneg-system@%s.iam.gserviceaccount.com", projectConfig.ProjectId)),
						}},
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{helmIstioBase, helmIstioD}), pulumi.Parent(gcpGKENodePool))
		if err != nil {
			return err
		}

		// Bind Kubernetes AutoNeg Service Account to Workload Identity
		resourceName = fmt.Sprintf("%s-iam-svc-k8s-%s", projectConfig.ResourceNamePrefix, cloudRegion.Region)
		_, err = serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
			ServiceAccountId: gcpServiceAccountAutoNeg.Name,
			Role:             pulumi.String("roles/iam.workloadIdentityUser"),
			Members: pulumi.StringArray{
				pulumi.String(fmt.Sprintf("serviceAccount:%s.svc.id.goog[autoneg-system/autoneg-controller-manager]", projectConfig.ProjectId)),
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{gcpGKENodePool, helmClusterOps}), pulumi.Parent(gcpGKENodePool))
		if err != nil {
			return err
		}
	}
	return nil
}
