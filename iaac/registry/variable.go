package registry

import (
	"mlops/iam"
)

var RegistryIAM = map[string]iam.IAM{
	"registry": {
		DisplayName:          "Registry Service Account",
		CreateServiceAccount: true,
		Roles: []string{
			"roles/artifactregistry.reader",
		},
		CreateMember:       true,
		ResourceNamePrefix: "registry-reader",
	},
	"github": {
		DisplayName:          "GitHub Actions Service Account",
		CreateServiceAccount: true,
		Roles: []string{
			"roles/artifactregistry.writer",
		},
		CreateMember:       true,
		ResourceNamePrefix: "github",
	},
}
