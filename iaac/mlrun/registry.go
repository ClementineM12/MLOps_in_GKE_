package mlrun

import (
	"encoding/json"
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createDockerRegistrySecret constructs the required Docker config JSON using
// the provided service account key JSON and creates a Kubernetes secret.
// The secret is of type "kubernetes.io/dockerconfigjson".
func createDockerRegistrySecret(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	serviceAccounts map[string]iam.ServiceAccountInfo,
	registry *artifactregistry.Repository,
	registryURL string,
) error {
	// Retrieve the service account info for mlrun.
	serviceAccount := serviceAccounts["mlrun"]
	serviceAccountKey := serviceAccount.Key

	// Build the docker config JSON using the service account key.
	dockerConfigJson := serviceAccountKey.PrivateKey.ApplyT(func(k string) (string, error) {
		type dockerAuth struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
		}
		type dockerConfig struct {
			Auths map[string]dockerAuth `json:"auths"`
		}

		// Create the JSON structure that Docker expects.
		config := dockerConfig{
			Auths: map[string]dockerAuth{
				// Use the Artifact Registry endpoint.
				registryURL: {
					Username: "_json_key",
					Password: k,
					Email:    projectConfig.Email,
				},
			},
		}

		b, err := json.Marshal(config)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}).(pulumi.StringOutput)

	// Create the Kubernetes secret in the "mlrun" namespace.
	_, err := coreV1.NewSecret(ctx, "mlrun-gcr-registry-credentials", &coreV1.SecretArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Namespace: pulumi.String(namespace),
			Name:      pulumi.String(registrySecretName),
		},
		Type: pulumi.String("kubernetes.io/dockerconfigjson"),
		StringData: pulumi.StringMap{
			".dockerconfigjson": dockerConfigJson,
		},
	}, pulumi.DependsOn([]pulumi.Resource{serviceAccount.ServiceAccount, registry}))
	if err != nil {
		return fmt.Errorf("error creating Kubernetes secret: %w", err)
	}
	return nil
}
