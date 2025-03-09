package infracomponents

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

func certManagerIssuerYAML(
	projectConfig global.ProjectConfig,
	namespace string,
) map[string]string {

	issuerYAML := fmt.Sprintf(`
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: %s
  namespace: %s
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: %s
    privateKeySecretRef:
      name: %s
    solvers:
    - http01:
        ingress:
          ingressClassName: nginx
`, LetsEncrypt, namespace, projectConfig.Email, LetsEncrypt)

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
  name: %s-tls-cert
  namespace: %s
spec:
  secretName: %s-secret-tls
  issuerRef:
    name: %s
    kind: Issuer
  commonName: %s
  dnsNames:
    - %s
`, namespace, namespace, namespace, LetsEncrypt, DNS, DNS)

	return map[string]string{
		"certificate": certYAML,
	}
}
