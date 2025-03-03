package infracomponents

import (
	"fmt"
	"mlops/global"

	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deployIngress deploys Ingress resources dynamically from a configuration map.
func deployIngress(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	namespace string,
	infraComponents InfraComponents,
) error {
	IngressMap := infraComponents.IngressMap
	// Loop over the map and create an Ingress resource for each configuration.
	for serviceRef, cfg := range IngressMap {
		if err := createIngress(ctx, projectConfig, serviceRef, namespace, cfg); err != nil {
			return err
		}
	}
	return nil
}

// buildIngressPaths converts a slice of IngressPathConfig into a slice of HTTPIngressPathArgs.
func buildIngressPaths(cfg IngressConfig) networkingv1.HTTPIngressPathArray {
	var paths networkingv1.HTTPIngressPathArray
	for _, rule := range cfg.Paths {
		path := rule.Path
		// if no path is provided, default to "/"
		if path == "" {
			path = "/"
		} else {
			// ensure the path starts with a slash
			if path[0] != '/' {
				path = "/" + path
			}
		}
		paths = append(paths, networkingv1.HTTPIngressPathArgs{
			Path:     pulumi.String(path),
			PathType: pulumi.String("Prefix"),
			Backend: networkingv1.IngressBackendArgs{
				Service: networkingv1.IngressServiceBackendArgs{
					Name: pulumi.String(rule.Service),
					Port: networkingv1.ServiceBackendPortArgs{
						Number: pulumi.Int(rule.Port),
					},
				},
			},
		})
	}
	return paths
}

// createIngress creates an Ingress resource using the provided configuration.
func createIngress(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	serviceRef string,
	namespace string,
	cfg IngressConfig,
) error {

	host := fmt.Sprintf("%s.%s", cfg.DNS, projectConfig.Domain)

	ingressName := serviceRef + "-ingress"
	_, err := networkingv1.NewIngress(ctx, ingressName, &networkingv1.IngressArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name:      pulumi.String(ingressName),
			Namespace: pulumi.String(namespace),
			Annotations: pulumi.StringMap{
				"cert-manager.io/issuer":                             pulumi.String("letsencrypt-production"),
				"kubernetes.io/ingress.class":                        pulumi.String("nginx"),
				"nginx.ingress.kubernetes.io/ssl-redirect":           pulumi.String("true"),
				"acme.cert-manager.io/http01-edit-in-place":          pulumi.String("true"),
				"nginx.ingress.kubernetes.io/whitelist-source-range": pulumi.String(projectConfig.WhitelistedIPs),
			},
		},
		Spec: networkingv1.IngressSpecArgs{
			IngressClassName: pulumi.String("nginx"),
			Rules: networkingv1.IngressRuleArray{
				networkingv1.IngressRuleArgs{
					Host: pulumi.String(host),
					Http: networkingv1.HTTPIngressRuleValueArgs{
						Paths: buildIngressPaths(cfg),
					},
				},
			},
			// TLS configuration: cert-manager will issue a certificate for this host.
			Tls: networkingv1.IngressTLSArray{
				networkingv1.IngressTLSArgs{
					Hosts: pulumi.StringArray{
						pulumi.String(host),
					},
					SecretName: pulumi.String(ingressName + "-tls"),
				},
			},
		},
	})
	return err
}
