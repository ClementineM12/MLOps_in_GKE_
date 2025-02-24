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
		Namespace:       pulumi.String("nginx-ingress"),
		CreateNamespace: pulumi.Bool(true),
		Chart:           pulumi.String("ingress-nginx"),
		Version:         pulumi.String("4.11.4"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(nginxIngressRepoURL),
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

func deployCertManager(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
) (*helm.Release, error) {

	resourceName := fmt.Sprintf("%s-cert-manager", projectConfig.ResourceNamePrefix)
	certManagerRelease, err := helm.NewRelease(ctx, resourceName, &helm.ReleaseArgs{
		Name:            pulumi.String("cert-manager"),
		Namespace:       pulumi.String("cert-manager"),
		CreateNamespace: pulumi.Bool(true),
		Chart:           pulumi.String("cert-manager"),
		Version:         pulumi.String("v1.17.0"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(certManagerRepoURL),
		},
		// Do not skip installing CRDs
		SkipCrds: pulumi.Bool(false),
		// Set the installCRDs value explicitly
		Values: pulumi.Map{
			"installCRDs": pulumi.Bool(true),
		},
		Timeout: pulumi.Int(300),
	}, pulumi.Provider(k8sProvider))

	if err != nil {
		return nil, err
	}

	if err := deployCertManagerIssuer(ctx, projectConfig, certManagerRelease, k8sProvider); err != nil {
		return nil, err
	}
	return certManagerRelease, nil
}

// deployCertManagerIssuer creates a Cert-Manager Issuer from a YAML manifest.
// The function takes an email address to inject into the manifest and a dependency resource,
// which should be the Cert-Manager Helm release, so that the issuer is created only after Cert-Manager is installed.
func deployCertManagerIssuer(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	certManagerRelease pulumi.Resource,
	k8sProvider *kubernetes.Provider,
) error {
	issuerYAML := fmt.Sprintf(`
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-production
  namespace: flyte
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
`, projectConfig.Email)

	_, err := yaml.NewConfigGroup(ctx, "cert-manager-issuer", &yaml.ConfigGroupArgs{
		YAML: []string{
			issuerYAML,
		},
	},
		pulumi.DependsOn([]pulumi.Resource{certManagerRelease}),
		pulumi.Provider(k8sProvider),
	)

	return err
}
