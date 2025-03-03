package mlrun

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
	serviceAccount := serviceAccounts["mlrun"]
	serviceAccountKey := serviceAccount.Key

	// Use the complete JSON key. Adjust this if your struct has a different field (e.g. JsonKey).
	// Here we assume serviceAccountKey.PrivateKey holds the full JSON key.
	dockerConfigData := serviceAccountKey.PrivateKey.ApplyT(func(key string) (string, error) {
		// Standard for Google registries: username is "_json_key"
		username := "_json_key"
		// Build the auth string "username:password" and base64-encode it.
		authStr := fmt.Sprintf("%s:%s", username, key)
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))

		// Build the docker config JSON in the format expected by Kubernetes.
		// This matches the output of: kubectl create secret docker-registry ...
		config := map[string]interface{}{
			"auths": map[string]interface{}{
				registryURL: map[string]string{
					"username": username,
					"password": key,
					"email":    projectConfig.Email,
					"auth":     encodedAuth,
				},
			},
		}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return "", err
		}
		// When using the Data field, Kubernetes expects the value to be base64 encoded.
		return base64.StdEncoding.EncodeToString(configBytes), nil
	}).(pulumi.StringOutput)

	resourceName := fmt.Sprintf("%s-mlrun-gcr-registry-credentials", projectConfig.ResourceNamePrefix)
	_, err := coreV1.NewSecret(ctx, resourceName, &coreV1.SecretArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Namespace: pulumi.String("mlrun"),
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
		return fmt.Errorf("error creating Kubernetes secret: %w", err)
	}
	return nil
}
