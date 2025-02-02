package mlrun

// import (
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
// 	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
// 	helm "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
// 	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
// )

// // Args holds the configuration parameters for deploying MLRun.
// type Args struct {
// 	// Kubernetes namespace where MLRun will be deployed.
// 	Namespace string
// 	// Docker registry credentials to pull images.
// 	RegistryServer   string // e.g. "https://registry.hub.docker.com/"
// 	RegistryUsername string
// 	RegistryPassword string
// 	RegistryEmail    string

// 	// Values for configuring MLRun.
// 	RegistryURL         string   // e.g. "index.docker.io/<your-username>"
// 	ExternalHostAddress string   // Host machine address or IP (e.g. minikube ip)
// 	NuclioIPAddresses   []string // List of external IP addresses for Nuclio dashboard.
// 	PipelinesEnabled    bool     // true to enable pipelines, false to disable (e.g. on Apple silicon)
// 	RedisURL            string   // Optional: Redis URL for the online feature store.
// 	ChartVersion        pulumi.String   // Optional: Specify the MLRun chart version (if empty, the latest available is used).
// }

// // Deploy installs MLRun in the given Kubernetes cluster. It performs the following:
// //   1. Creates the target namespace.
// //   2. Creates a docker-registry secret with the provided credentials.
// //   3. Deploys the MLRun Community Edition Helm chart in the namespace.
// func Deploy(ctx *pulumi.Context, k8sProvider *kubernetes.Provider, args Args) (*helm.Chart, error) {
// 	// 1. Create the target namespace.
// 	ns, err := corev1.NewNamespace(ctx, args.Namespace, &corev1.NamespaceArgs{
// 		Metadata: &pulumi.ObjectMetaArgs{
// 			Name: pulumi.String(args.Namespace),
// 		},
// 	}, pulumi.Provider(k8sProvider))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create namespace %s: %w", args.Namespace, err)
// 	}

// 	// 2. Create the docker-registry secret.
// 	// The secret must be of type "kubernetes.io/dockerconfigjson" with a key ".dockerconfigjson"
// 	// containing the base64-encoded JSON of the docker config.
// 	authString := fmt.Sprintf("%s:%s", args.RegistryUsername, args.RegistryPassword)
// 	authEncoded := base64.StdEncoding.EncodeToString([]byte(authString))

// 	dockerConfig := map[string]interface{}{
// 		"auths": map[string]interface{}{
// 			args.RegistryServer: map[string]string{
// 				"username": args.RegistryUsername,
// 				"password": args.RegistryPassword,
// 				"email":    args.RegistryEmail,
// 				"auth":     authEncoded,
// 			},
// 		},
// 	}
// 	dockerConfigJSON, err := json.Marshal(dockerConfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal docker config: %w", err)
// 	}
// 	// Note: Kubernetes expects the data value to be base64-encoded automatically,
// 	// but here we provide the raw JSON string as the secret’s data value.
// 	secret, err := corev1.NewSecret(ctx, "registry-credentials", &corev1.SecretArgs{
// 		Metadata: &pulumi.ObjectMetaArgs{
// 			Name:      pulumi.String("registry-credentials"),
// 			Namespace: ns.Metadata.Name,
// 		},
// 		Type: pulumi.String("kubernetes.io/dockerconfigjson"),
// 		Data: pulumi.StringMap{
// 			".dockerconfigjson": pulumi.String(string(dockerConfigJSON)),
// 		},
// 	}, pulumi.Provider(k8sProvider))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create docker registry secret: %w", err)
// 	}

// 	// 3. Build the Helm chart values.
// 	// These values correspond to the "helm install" command documented in MLRun's docs.
// 	values := pulumi.Map{
// 		"global": pulumi.Map{
// 			"registry": pulumi.Map{
// 				"url":        pulumi.String(args.RegistryURL),
// 				"secretName": pulumi.String(secret.Metadata.Name().Elem()), // should resolve to "registry-credentials"
// 			},
// 			"externalHostAddress": pulumi.String(args.ExternalHostAddress),
// 		},
// 		"nuclio": pulumi.Map{
// 			"dashboard": pulumi.Map{
// 				"externalIPAddresses": pulumi.ToStringArray(args.NuclioIPAddresses),
// 			},
// 		},
// 		"pipelines": pulumi.Map{
// 			"enabled": pulumi.Bool(args.PipelinesEnabled),
// 		},
// 	}

// 	// Optionally add Redis configuration if provided.
// 	if args.RedisURL != "" {
// 		values["mlrun"] = pulumi.Map{
// 			"api": pulumi.Map{
// 				"extraEnvKeyValue": pulumi.Map{
// 					"MLRUN_REDIS__URL": pulumi.String(args.RedisURL),
// 				},
// 			},
// 		}
// 	}

// 	// 4. Deploy the MLRun CE Helm chart.
// 	// If ChartVersion is provided, use it; otherwise, omit the version to get the latest.
// 	var chartVersion pulumi.StringPtrInput
// 	if args.ChartVersion != "" {
// 		chartVersion = pulumi.StringPtr(args.ChartVersion)
// 	} else {
// 		chartVersion = nil
// 	}

// 	chart, err := helm.NewChart(ctx, "mlrun-ce", helm.ChartArgs{
// 		// The chart name within the repo is "mlrun-ce".
// 		Chart:   pulumi.String("mlrun-ce"),
// 		Version: chartVersion,
// 		// The repository URL as per the MLRun docs.
// 		FetchArgs: helm.FetchArgs{
// 			Repo: pulumi.String("https://mlrun.github.io/ce"),
// 		},
// 		// Deploy the chart into the specified namespace.
// 		Namespace: pulumi.String(args.Namespace),
// 		// Pass in the custom values.
// 		Values: values,
// 		// Wait for the deployment to finish, with a longer timeout (960 seconds) as recommended.
// 		Wait:    pulumi.Bool(true),
// 		Timeout: pulumi.Int(960),
// 	}, pulumi.Provider(k8sProvider))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to deploy mlrun helm chart: %w", err)
// 	}

// 	return chart, nil
// }
