package mlrun

import (
	"mlops/iam"
	infracomponents "mlops/infra_components"
)

var (
	registrySecretName = "gcr-registry-credentials"
	// serviceAccountSecretName = "mlrun-sa-credentials"
	bucketName   = "mlrun-project-bucket-01"
	registryName = "mlrun-artifacts"

	MLRunIAM = map[string]iam.IAM{
		"mlrun": {
			ResourceNamePrefix: "mlrun",
			DisplayName:        "MLRun Registry Service",
			Roles: []string{
				"roles/artifactregistry.writer",
				"roles/artifactregistry.reader",
				"roles/storage.objectAdmin",
			},
			CreateServiceAccount: true,
			CreateMember:         true,
			CreateKey:            true,
		},
	}
	// Define a map of Ingress configurations for your components.
	ingressMap = map[string]infracomponents.IngressConfig{
		"grafana": {
			DNS: "grafana",
			Paths: []infracomponents.IngressPathConfig{
				{
					Service: "grafana",
					Port:    3000,
				},
			},
		},
		"minio": {
			DNS: "minio",
			Paths: []infracomponents.IngressPathConfig{
				{
					Service: "minio",
					Port:    9001,
				},
			},
		},
		"mlrun": {
			DNS: "mlrun", // using the base domain for central UI
			Paths: []infracomponents.IngressPathConfig{
				{
					Path:    "pipelines",
					Service: "ml-pipeline-ui",
					Port:    3000,
				},
				{
					// Path:    "lab",
					Service: "mlrun-jupyter",
					Port:    8888,
				},
				{
					// No path provided means the root path for the central UI.
					Path:    "central",
					Service: "mlrun-ui",
					Port:    80,
				},
				{
					Path:    "nuclio",
					Service: "nuclio-dashboard",
					Port:    8070,
				},
			},
		},
	}
)
