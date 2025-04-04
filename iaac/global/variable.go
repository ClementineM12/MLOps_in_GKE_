package global

var (
	MLOpsAllowedTargets = []string{
		"flyte",
		"mlrun",
		"kubeflow",
	}

	// Recommended [https://googlecloudplatform.github.io/kubeflow-gke-docs/dev/docs/deploy/project-setup/#setting-up-a-project]
	gcpServices = []string{
		"serviceusage.googleapis.com",
		"compute.googleapis.com",
		"container.googleapis.com",
		"iam.googleapis.com",
		"servicenetworking.googleapis.com",
		"servicemanagement.googleapis.com",
		"cloudresourcemanager.googleapis.com",
		"ml.googleapis.com",
		"iap.googleapis.com",
		"sqladmin.googleapis.com",
		"meshconfig.googleapis.com",
		"krmapihosting.googleapis.com",
		"servicecontrol.googleapis.com",
		"endpoints.googleapis.com",
		"cloudbuild.googleapis.com",
		"certificatemanager.googleapis.com",
		"artifactregistry.googleapis.com",
		// OIDC
		"securitycenter.googleapis.com",
	}

	CloudRegions = []CloudRegion{
		{
			Id:                  "001",
			Country:             "Warsaw",
			Region:              "europe-central2",
			SubnetIp:            "10.128.50.0/24",
			MasterIpv4CidrBlock: "172.16.0.0/28",
		},
		{
			Id:                  "002",
			Country:             "Finland",
			Region:              "europe-north1",
			SubnetIp:            "10.128.100.0/24",
			MasterIpv4CidrBlock: "172.16.0.16/28",
		},
		{
			Id:                  "003",
			Country:             "Madrid",
			Region:              "europe-southwest1",
			SubnetIp:            "10.128.150.0/24",
			MasterIpv4CidrBlock: "172.16.0.32/28",
		},
		{
			Id:                  "004",
			Country:             "Belgium",
			Region:              "europe-west1",
			SubnetIp:            "10.128.200.0/24",
			MasterIpv4CidrBlock: "172.16.0.48/28",
		},
		{
			Id:                  "005",
			Country:             "London",
			Region:              "europe-west2",
			SubnetIp:            "10.128.250.0/24",
			MasterIpv4CidrBlock: "172.16.0.64/28",
		},
		{
			Id:                  "006",
			Country:             "Frankfurt",
			Region:              "europe-west3",
			SubnetIp:            "10.129.50.0/24",
			MasterIpv4CidrBlock: "172.16.0.80/28",
		},
		{
			Id:                  "007",
			Country:             "Netherlands",
			Region:              "europe-west4",
			SubnetIp:            "10.129.100.0/24",
			MasterIpv4CidrBlock: "172.16.0.96/28",
		},
		{
			Id:                  "008",
			Country:             "Zurich",
			Region:              "europe-west6",
			SubnetIp:            "10.129.150.0/24",
			MasterIpv4CidrBlock: "172.16.0.112/28",
		},
		{
			Id:                  "009",
			Country:             "Milan",
			Region:              "europe-west8",
			SubnetIp:            "10.129.200.0/24",
			MasterIpv4CidrBlock: "172.16.0.128/28",
		},
		{
			Id:                  "010",
			Country:             "Paris",
			Region:              "europe-west9",
			SubnetIp:            "10.129.250.0/24",
			MasterIpv4CidrBlock: "172.16.0.144/28",
		},
		{
			Id:                  "011",
			Country:             "Berlin",
			Region:              "europe-west10",
			SubnetIp:            "10.130.50.0/24",
			MasterIpv4CidrBlock: "172.16.0.160/28",
		},
		{
			Id:                  "012",
			Country:             "Turin",
			Region:              "europe-west12",
			SubnetIp:            "10.130.100.0/24",
			MasterIpv4CidrBlock: "172.16.0.176/28",
		},
	}
)
