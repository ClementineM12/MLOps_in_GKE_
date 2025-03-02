package mlrun

import (
	"mlops/iam"
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
			CreateKey:            true,
		},
	}
)
