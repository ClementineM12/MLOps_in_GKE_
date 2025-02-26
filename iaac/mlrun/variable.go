package mlrun

import (
	"mlops/iam"
	infracomponents "mlops/infra_components"
)

var (
	infraComponents = infracomponents.InfraComponents{
		CertManager:  true,
		NginxIngress: true,
	}

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
