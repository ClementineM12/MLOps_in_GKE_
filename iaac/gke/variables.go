package gke

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"mlops/iam"
)

var (
	GKEDefaultVersion = "1.31.5-gke.1169000"

	AdministrationIAM = map[string]iam.IAM{
		"admin": {
			ResourceNamePrefix: "gke",
			DisplayName:        "GKE Admin",
			Roles: []string{
				"roles/logging.logWriter",
				"roles/monitoring.metricWriter",
				"roles/meshtelemetry.reporter",
				"roles/cloudtrace.agent",
				"roles/monitoring.viewer",
				"roles/storage.objectViewer",
				"roles/container.defaultNodeServiceAccount",
			},
			CreateServiceAccount: true,
			CreateMember:         true,
		},
	}

	nodePoolsConfig = NodePoolConfigs{
		"highmem": NodePoolConfig{
			MachineType:      "e2-standard-8",
			DiskSizeGb:       100,
			InitialNodeCount: 0,
			MinNodeCount:     0,
			MaxNodeCount:     5,
			Preemptible:      false,
			LocationPolicy:   "ANY",
			ResourceLabels: pulumi.StringMap{
				"type-dedicated": pulumi.String("memory-optimized"),
			},
			Labels: pulumi.StringMap{
				"dedicated": pulumi.String("highmem"),
			},
		},
		"cpu": NodePoolConfig{
			MachineType:      "c2-standard-16",
			DiskSizeGb:       100,
			InitialNodeCount: 0,
			MinNodeCount:     0,
			MaxNodeCount:     5,
			Preemptible:      false,
			LocationPolicy:   "ANY",
			ResourceLabels: pulumi.StringMap{
				"type-dedicated": pulumi.String("cpu-optimized"),
			},
			Labels: pulumi.StringMap{
				"dedicated": pulumi.String("cpu"),
			},
		},
	}
)
