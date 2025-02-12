package autoneg

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	namespace          = "autoneg-system"
	kubeRBACProxyImage = "gcr.io/kubebuilder/kube-rbac-proxy:v0.16.0"
)

// createAutoNEGKubernetesResources deploys the AutoNEG system with RBAC, service account, and controller deployment
func createAutoNEGKubernetesResources(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccountEmail pulumi.StringArrayOutput,
) pulumi.Output {

	// Create Namespace
	ns, err := createNamespace(ctx, projectConfig, k8sProvider)
	if err != nil {
		ctx.Log.Warn("Failed to create namespace for Auto NEG Controller", nil)
		return pulumi.StringOutput{}.ApplyT(func(_ string) string { return "Error: Namespace creation failed" })
	}

	// Create Service Account
	autoNegServiceAccount, err := createServiceAccount(ctx, projectConfig, k8sProvider, serviceAccountEmail, ns)
	if err != nil {
		ctx.Log.Warn("Failed to create Service Account for Auto NEG Controller", nil)
		return pulumi.StringOutput{}.ApplyT(func(_ string) string { return "Error: Service Account creation failed" })
	}

	// Create Roles and Bindings
	err = createAutoNEGRBAC(ctx, projectConfig, k8sProvider, autoNegServiceAccount)
	if err != nil {
		ctx.Log.Warn("Failed to create RBAC for Auto NEG Controller", nil)
		return pulumi.StringOutput{}.ApplyT(func(_ string) string { return "Error: RBAC setup failed" })
	}

	// Deploy AutoNEG Service
	err = createAutoNegService(ctx, projectConfig, k8sProvider, ns)
	if err != nil {
		ctx.Log.Warn("Failed to create AutoNEG Service", nil)
		return pulumi.StringOutput{}.ApplyT(func(_ string) string { return "Error: AutoNEG Service creation failed" })
	}

	// Deploy AutoNEG Controller
	negDeployment, err := createAutoNegDeployment(ctx, projectConfig, k8sProvider, autoNegServiceAccount, ns)
	if err != nil {
		ctx.Log.Warn("Failed to create AutoNEG Deployment", nil)
		return pulumi.StringOutput{}.ApplyT(func(_ string) string { return "Error: AutoNEG Deployment failed" })
	}

	return negDeployment.Status.ReadyReplicas().ApplyT(func(replicaCount *int) string {
		if replicaCount != nil && *replicaCount > 0 {
			return fmt.Sprintf("%d replicas ready", *replicaCount)
		}
		return "AutoNEG deployment in progress"
	})
}

// Create Kubernetes Namespace
func createNamespace(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*coreV1.Namespace, error) {

	resourceName := fmt.Sprintf("%s-autoneg-namespace", projectConfig.ResourceNamePrefix)
	return coreV1.NewNamespace(ctx, resourceName, &coreV1.NamespaceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(namespace),
			Labels: pulumi.StringMap{
				"app":           pulumi.String("autoneg"),
				"control-plane": pulumi.String("controller-manager"),
			},
		},
	}, pulumi.Provider(k8sProvider))
}

// Create Kubernetes Service Account
func createServiceAccount(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccountEmail pulumi.StringArrayOutput,
	ns *coreV1.Namespace,
) (*coreV1.ServiceAccount, error) {

	resourceName := fmt.Sprintf("%s-autoneg-sa", projectConfig.ResourceNamePrefix)
	return coreV1.NewServiceAccount(ctx, resourceName, &coreV1.ServiceAccountArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Namespace: ns.Metadata.Name(),
			Name:      pulumi.String("autoneg-service-account"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("autoneg"),
			},
			Annotations: pulumi.StringMap{
				"iam.gke.io/gcp-service-account": serviceAccountEmail.Index(pulumi.Int(0)).ToStringOutput(),
			},
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ns}))
}

// Create RBAC Roles and Bindings
func createAutoNEGRBAC(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
) error {

	err := createClusterRoles(ctx, projectConfig, k8sProvider, serviceAccount)
	if err != nil {
		return err
	}
	err = createRoles(ctx, projectConfig, k8sProvider, serviceAccount)
	if err != nil {
		return err
	}
	return nil
}

// Create AutoNEG Service for Metrics Collection
func createAutoNegService(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	ns *coreV1.Namespace,
) error {

	resourceName := fmt.Sprintf("%s-autoneg-metrics-service", projectConfig.ResourceNamePrefix)
	_, err := coreV1.NewService(ctx, resourceName, &coreV1.ServiceArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Namespace: ns.Metadata.Name(),
			Name:      pulumi.String("autoneg-controller-manager-metrics-service"),
			Labels: pulumi.StringMap{
				"app":           pulumi.String("autoneg"),
				"control-plane": pulumi.String("controller-manager"),
			},
			Annotations: pulumi.StringMap{
				"prometheus.io/port":   pulumi.String("8443"),
				"prometheus.io/scrape": pulumi.String("true"),
			},
		},
		Spec: &coreV1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Ports: coreV1.ServicePortArray{
				&coreV1.ServicePortArgs{
					Name:       pulumi.String("https"),
					Port:       pulumi.Int(8443),
					TargetPort: pulumi.String("https"),
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: pulumi.StringMap{
				"app":           pulumi.String("autoneg"),
				"control-plane": pulumi.String("controller-manager"),
			},
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{ns}))
	return err
}

// Deploy AutoNEG Controller
func createAutoNegDeployment(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	autoNegServiceAccount *coreV1.ServiceAccount,
	ns *coreV1.Namespace,
) (*v1.Deployment, error) {

	resourceName := fmt.Sprintf("%s-autoneg-controller-manager", projectConfig.ResourceNamePrefix)
	negDeployment, err := v1.NewDeployment(ctx, resourceName, &v1.DeploymentArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Namespace: ns.Metadata.Name(),
			Name:      pulumi.String("autoneg-controller-manager"),
		},
		Spec: &v1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metaV1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("autoneg"),
				},
			},
			Template: &coreV1.PodTemplateSpecArgs{
				Metadata: &metaV1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("autoneg"),
					},
				},
				Spec: &coreV1.PodSpecArgs{
					ServiceAccountName: autoNegServiceAccount.Metadata.Name(),
					Containers: coreV1.ContainerArray{
						&coreV1.ContainerArgs{
							Name:  pulumi.String("manager"),
							Image: pulumi.String(kubeRBACProxyImage),
							Args: pulumi.StringArray{
								pulumi.String("--health-probe-bind-address=:8081"),
								pulumi.String("--metrics-bind-address=127.0.0.1:8080"),
								pulumi.String("--leader-elect"),
							},
							Ports: coreV1.ContainerPortArray{
								&coreV1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8081),
									Name:          pulumi.String("metrics"),
								},
							},
						},
					},
				},
			},
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{autoNegServiceAccount}))
	return negDeployment, err
}
