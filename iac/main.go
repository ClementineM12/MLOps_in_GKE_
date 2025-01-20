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

var CloudRegions = []vpc.CloudRegion{
	{
		Id:       "001",
		Enabled:  true,
		Region:   "europe-west6",
		SubnetIp: "10.128.100.0/24",
	},
	{
		Id:       "002",
		Enabled:  false,
		Region:   "europe-north1",
		SubnetIp: "10.129.100.0/24",
	},
	{
		Id:       "003",
		Enabled:  false,
		Region:   "europe-west8",
		SubnetIp: "10.130.100.0/24",
	},
	{
		Id:       "004",
		Enabled:  false,
		Region:   "europe-west9",
		SubnetIp: "10.130.150.0/24",
	},
	{
		Id:       "005",
		Enabled:  false,
		Region:   "europe-west3",
		SubnetIp: "10.130.200.0/24",
	},
	{
		Id:       "006",
		Enabled:  false,
		Region:   "europe-central2",
		SubnetIp: "10.130.250.0/24",
	},
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Global Variables
		var SSL bool
		var domain string
		var gcpProjectId string
		var resourceNamePrefix string
		var createWorkloadIdentityPool bool

		// Instanciate Pulumi Configuration
		cfg := config.New(ctx, "")

		// Check configuration values
		domain = cfg.Get("domainName")
		gcpProjectId, err := ConfigureProjectId(ctx)
		if err != nil {
			return err
		}
		resourceNamePrefix, err = ConfigureResourcePrefix(cfg)
		if err != nil {
			return err
		}
		SSL = ConfigureSSL(domain)

		// Enable Google API's on the Specified Project.
		gcpDependencies, err := project.EnableGCPServices(ctx, resourceNamePrefix, gcpProjectId)
		if err != nil {
			return err
		}
		// Create Global Load Balancer Static IP
		gcpGlobalAddress, err := vpc.CreateLoadBalancerStaticIP(ctx, resourceNamePrefix, gcpProjectId, gcpDependencies)
		if err != nil {
			return err
		}
		// Read Cluster Configuration YAML file
		// err = RetrieveClusterConfiguration(cfg)

		// -------------------------- IAM ------------------------------
		// Create Google Cloud Service Account
		gcpServiceAccount, err := gcpIAM.CreateServiceAccount(ctx, resourceNamePrefix, gcpProjectId, "Admin")
		if err != nil {
			return err
		}
		// Create AutoNeg Service Account: custom IAM Role that will be used by the AutoNeg Kubernetes Deployment
		gcpServiceAccountAutoNeg, err := gcpIAM.CreateServiceAccount(ctx, resourceNamePrefix, gcpProjectId, "AutoNEG")
		if err != nil {
			return err
		}
		// -------------------------------------------------------------

		// -----------------  Workload Identity Pool --------------------
		if createWorkloadIdentityPool {
			err = project.CreateWorkloadIdentityPool(ctx, resourceNamePrefix, gcpProjectId)
			if err != nil {
				return err
			}
		}
		// -------------------------------------------------------------

		// ---------------------  Object storage ------------------------
		// Create Object Storage
		err = storage.CreateObjectStorage(ctx, resourceNamePrefix)
		if err != nil {
			return err
		}

		// -------------------------------------------------------------

		// ---------------------------  VPC ----------------------------
		// Create Google Cloud VPC Network (Global Resource)
		gcpNetwork, err := vpc.CreateVPC(ctx, resourceNamePrefix, gcpProjectId, gcpDependencies)
		if err != nil {
			return nil
		}

		// Create Load Balancer Backend Service
		gcpBackendService, err := vpc.CreateLoadBalancerBackendService(ctx, resourceNamePrefix, gcpProjectId, gcpDependencies)
		if err != nil {
			return nil
		}

		// Create Managed SSL Certificate
		if SSL {
			err = vpc.ConfigureSSLCertificate(ctx, resourceNamePrefix, gcpProjectId, domain, gcpBackendService, gcpGlobalAddress, gcpDependencies)
			if err != nil {
				return err
			}
		}
		// Create URL Maps
		err = vpc.CreateLoadBalancerURLMapHTTP(ctx, resourceNamePrefix, gcpProjectId, domain, SSL, gcpGlobalAddress, gcpBackendService)
		if err != nil {
			return nil
		}
		// -------------------------------------------------------------------

		// Process Each Cloud Region;
		for _, cloudRegion := range CloudRegions {
			if !cloudRegion.Enabled {
				// Logging Region Skipping
				fmt.Printf("[ INFORMATION ] - Cloud Region: %s - SKIPPING\n", cloudRegion.Region)
				continue
			}

			// Logging Region Processing
			fmt.Printf("[ INFORMATION ] - Cloud Region: %s - PROCESSING\n", cloudRegion.Region)

			// Create VPC Subnet for Cloud Region
			gcpSubnetwork, err := vpc.CreateVPCSubnet(ctx, resourceNamePrefix, gcpProjectId, cloudRegion, gcpNetwork)
			if err != nil {
				return err
			}

			// ---------------------------  GKE ----------------------------
			gcpGKENodePool, k8sProvider, err := gke.CreateGKE(ctx, resourceNamePrefix, gcpProjectId, &cloudRegion, gcpNetwork, gcpSubnetwork, gcpServiceAccount)
			if err != nil {
				return err
			}
			// -------------------------------------------------------------------

			// Install Istio Service Mesh
			helmIstioBase, helmIstioD, err := istio.InstallIstioHelm(ctx, resourceNamePrefix, cloudRegion, k8sProvider, gcpGKENodePool)
			if err != nil {
				return err
			}

			// Create New Namespace in the GKE Clusters for Application Deployments
			resourceName := fmt.Sprintf("%s-k8s-ns-app-%s", resourceNamePrefix, cloudRegion.Region)
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
			err = istio.CreateIstioIngressGateway(ctx, resourceNamePrefix, cloudRegion, k8sProvider, helmIstioBase, helmIstioD, gcpGKENodePool, k8sAppNamespace, gcpBackendService)
			if err != nil {
				return err
			}

			// Deploy Cluster Ops components for GKE AutoNeg
			resourceName = fmt.Sprintf("%s-cluster-ops-%s", resourceNamePrefix, cloudRegion.Region)
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
								"iam.gke.io/gcp-service-account": pulumi.String(fmt.Sprintf("autoneg-system@%s.iam.gserviceaccount.com", gcpProjectId)),
							}},
					},
				},
			}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{helmIstioBase, helmIstioD}), pulumi.Parent(gcpGKENodePool))
			if err != nil {
				return err
			}

			// Bind Kubernetes AutoNeg Service Account to Workload Identity
			resourceName = fmt.Sprintf("%s-iam-svc-k8s-%s", resourceNamePrefix, cloudRegion.Region)
			_, err = serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
				ServiceAccountId: gcpServiceAccountAutoNeg.Name,
				Role:             pulumi.String("roles/iam.workloadIdentityUser"),
				Members: pulumi.StringArray{
					pulumi.String(fmt.Sprintf("serviceAccount:%s.svc.id.goog[autoneg-system/autoneg-controller-manager]", gcpProjectId)),
				},
			}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{gcpGKENodePool, helmClusterOps}), pulumi.Parent(gcpGKENodePool))
			if err != nil {
				return err
			}
		}

		return nil
	})
}
