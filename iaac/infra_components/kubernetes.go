package infracomponents

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deployNginxController(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-nginx-controller", projectConfig.ResourceNamePrefix)
	return helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("ingress-nginx"),
		Namespace:       pulumi.String(NginxControllerNamespace),
		CreateNamespace: pulumi.Bool(true),
		Chart:           pulumi.String(NginxControllerHelmChart),
		Version:         pulumi.String(NginxControllerHelmChartVersion),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(NginxControllerHelmChartRepo),
		},
		Values: pulumi.Map{
			"controller": pulumi.Map{
				"service": pulumi.Map{
					"externalTrafficPolicy": pulumi.String("Local"),
				},
			},
		},
	}, pulumi.Provider(k8sProvider))
}

// configGroup aggregates the YAML resources and creates a ConfigGroup for each.
// The certManagerRelease parameter is used as a dependency so that the resources are created only after cert-manager is deployed.
func configGroup(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	namespace string,
	certManagerRelease pulumi.Resource,
	k8sProvider *kubernetes.Provider,
	infraComponents InfraComponents,
) error {
	// Create a map to hold resource names and YAML manifests.
	resources := make(map[string]string)

	if infraComponents.CertManagerIssuer {
		merge(resources, certManagerIssuerYAML(projectConfig, namespace))
	}
	if infraComponents.Certificate && infraComponents.Domain != "" {
		merge(resources, certificateYAML(projectConfig, namespace))
	}

	// Iterate over the collected resources and create a ConfigGroup for each.
	for name, resourceYAML := range resources {
		resourceName := fmt.Sprintf("%s-cert-manager-%s", projectConfig.ResourceNamePrefix, name)
		_, err := yaml.NewConfigGroup(ctx, resourceName, &yaml.ConfigGroupArgs{
			YAML: []string{resourceYAML},
		},
			pulumi.DependsOn([]pulumi.Resource{certManagerRelease}),
			pulumi.Provider(k8sProvider),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func deployCertManager(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	namespace string,
	k8sProvider *kubernetes.Provider,
	infraComponents InfraComponents,
	opts ...pulumi.ResourceOption,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-cert-manager", projectConfig.ResourceNamePrefix)
	certManagerRelease, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("cert-manager"),
		Namespace:       pulumi.String(CertManagerNamespace),
		CreateNamespace: pulumi.Bool(true),
		Chart:           pulumi.String(CertManagerHelmChart),
		Version:         pulumi.String(CertManagerHelmChartVersion),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(CertManagerHelmChartRepo),
		},
		// Do not skip installing CRDs
		SkipCrds: pulumi.Bool(false),
		// Set the installCRDs value explicitly
		Values: pulumi.Map{
			"installCRDs": pulumi.Bool(true),
		},
		Timeout: pulumi.Int(300),
	}, append(opts, pulumi.Provider(k8sProvider))...)

	if err != nil {
		return nil, err
	}
	if err := configGroup(ctx, projectConfig, namespace, certManagerRelease, k8sProvider, infraComponents); err != nil {
		return nil, err
	}
	return certManagerRelease, nil
}

func certManagerIssuerYAML(
	projectConfig global.ProjectConfig,
	namespace string,
) map[string]string {

	issuerYAML := fmt.Sprintf(`
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-production
  namespace: %s
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: %s
    privateKeySecretRef:
      name: letsencrypt-production
    solvers:
    - http01:
        ingress:
          ingressClassName: nginx
`, namespace, projectConfig.Email)

	return map[string]string{
		"issuer": issuerYAML,
	}
}

// certificateYAML returns the YAML manifest for the Certificate.
func certificateYAML(
	projectConfig global.ProjectConfig,
	namespace string,
) map[string]string {
	DNS := fmt.Sprintf("%s.%s", namespace, projectConfig.Domain)
	certYAML := fmt.Sprintf(`
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: mlrun-tls-cert
  namespace: %s
spec:
  secretName: mlrun-secret-tls
  issuerRef:
    name: letsencrypt-production
    kind: Issuer
  commonName: %s
  dnsNames:
    - %s
`, namespace, DNS, DNS)

	return map[string]string{
		"certificate": certYAML,
	}
}
