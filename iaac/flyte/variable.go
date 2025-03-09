package flyte

import (
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	FlyteIAM = map[string]iam.IAM{
		"flyteadmin": {
			Permissions: pulumi.StringArray{
				pulumi.String("iam.serviceAccounts.signBlob"),
				pulumi.String("storage.buckets.get"),
				pulumi.String("storage.objects.create"),
				pulumi.String("storage.objects.delete"),
				pulumi.String("storage.objects.get"),
				pulumi.String("storage.objects.getIamPolicy"),
				pulumi.String("storage.objects.update"),
			},
			CreateRole:           true,
			CreateServiceAccount: true,
			WorkloadIdentityBinding: []string{
				"flyte/flyteadmin",
			},
			ResourceNamePrefix: "flyte",
		},
		"flytepropeller": {
			Permissions: pulumi.StringArray{
				pulumi.String("storage.buckets.get"),
				pulumi.String("storage.objects.create"),
				pulumi.String("storage.objects.delete"),
				pulumi.String("storage.objects.get"),
				pulumi.String("storage.objects.list"),
				pulumi.String("storage.objects.getIamPolicy"),
				pulumi.String("storage.objects.update"),
			},
			CreateRole:           true,
			CreateServiceAccount: true,
			WorkloadIdentityBinding: []string{
				"flyte/flytepropeller",
			},
			ResourceNamePrefix: "flyte",
		},
		"flytescheduler": {
			Permissions: pulumi.StringArray{
				pulumi.String("storage.buckets.get"),
				pulumi.String("storage.objects.create"),
				pulumi.String("storage.objects.delete"),
				pulumi.String("storage.objects.get"),
				pulumi.String("storage.objects.getIamPolicy"),
				pulumi.String("storage.objects.update"),
			},
			CreateRole:           true,
			CreateServiceAccount: true,
			WorkloadIdentityBinding: []string{
				"flyte/flytescheduler",
			},
			ResourceNamePrefix: "flyte",
		},
		"datacatalog": {
			Permissions: pulumi.StringArray{
				pulumi.String("storage.buckets.get"),
				pulumi.String("storage.objects.create"),
				pulumi.String("storage.objects.delete"),
				pulumi.String("storage.objects.get"),
				pulumi.String("storage.objects.update"),
			},
			CreateRole:           true,
			CreateServiceAccount: true,
			WorkloadIdentityBinding: []string{
				"flyte/datacatalog",
			},
			ResourceNamePrefix: "flyte",
		},
		"flyteworkers": {
			Permissions: pulumi.StringArray{
				pulumi.String("storage.buckets.get"),
				pulumi.String("storage.objects.create"),
				pulumi.String("storage.objects.delete"),
				pulumi.String("storage.objects.get"),
				pulumi.String("storage.objects.list"),
				pulumi.String("storage.objects.update"),
			},
			CreateRole:           true,
			CreateServiceAccount: true,
			RoleBindings:         []string{"roles/artifactregistry.reader"},
			ResourceNamePrefix:   "flyte",
		},
		"artifactregistry-writer": {
			CreateRole:           false,
			CreateServiceAccount: true,
			RoleBindings:         []string{"roles/artifactregistry.writer"},
			ResourceNamePrefix:   "flyte",
		},
	}

	cloudSQLConfig = global.CloudSQLConfig{
		User:               "flyteadmin",
		Database:           "flyte",
		InstancePrefixName: "flyte",
	}
)
