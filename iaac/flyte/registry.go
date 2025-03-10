package flyte

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
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
	k8sProvider *kubernetes.Provider,
) error {
	// Retrieve the service account info for mlrun.
	serviceAccount := serviceAccounts["flyteworkers"]

	// Decode the private key using the standard library.
	decodedPrivateKey := serviceAccount.Key.PrivateKey.ApplyT(func(encoded string) (string, error) {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return "", fmt.Errorf("error decoding private key: %w", err)
		}
		return string(decoded), nil
	}).(pulumi.StringOutput)

	// Build the docker config JSON as an output.
	dockerConfigData := decodedPrivateKey.ApplyT(func(pk string) (string, error) {
		username := "_json_key"
		authStr := fmt.Sprintf("%s:%s", username, pk)
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))

		// Build the docker config JSON structure.
		config := map[string]interface{}{
			"auths": map[string]interface{}{
				registryURL: map[string]string{
					"username": username,
					"password": pk,
					"email":    projectConfig.Email,
					"auth":     encodedAuth,
				},
			},
		}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return "", fmt.Errorf("error marshalling docker config: %w", err)
		}

		// The secret data should be base64-encoded.
		encodedData := base64.StdEncoding.EncodeToString(configBytes)
		return encodedData, nil
	}).(pulumi.StringOutput)

	for _, namespace := range flyteNamespaces {
		resourceName := fmt.Sprintf("%s-%s-gcr-creds", projectConfig.ResourceNamePrefix, namespace)
		_, err := coreV1.NewSecret(ctx, resourceName, &coreV1.SecretArgs{
			Metadata: &metaV1.ObjectMetaArgs{
				Namespace: pulumi.String(namespace),
				Name:      pulumi.String(registrySecretName),
			},
			Type: pulumi.String("kubernetes.io/dockerconfigjson"),
			Data: pulumi.StringMap{
				".dockerconfigjson": dockerConfigData,
			},
		},
			pulumi.Provider(k8sProvider),
			pulumi.DependsOn([]pulumi.Resource{serviceAccount.ServiceAccount, registry}),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
