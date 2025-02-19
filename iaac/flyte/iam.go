package flyte

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createFlyteIAM(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (map[string]struct {
	AccountId pulumi.StringOutput
	Member    pulumi.StringArrayOutput
}, error) {

	serviceAccounts, err := createServiceAccounts(ctx, projectConfig)
	if err != nil {
		return nil, err
	}
	if err := createIAMBindings(ctx, projectConfig, serviceAccounts); err != nil {
		return nil, err
	}

	ctx.Export("flyteadminServiceAccount", serviceAccounts["flyteadmin"].AccountId)
	ctx.Export("flytepropellerServiceAccount", serviceAccounts["flytepropeller"].AccountId)
	ctx.Export("flyteschedulerServiceAccount", serviceAccounts["flytescheduler"].AccountId)
	ctx.Export("datacatalogServiceAccount", serviceAccounts["datacatalog"].AccountId)
	ctx.Export("flyteworkersServiceAccount", serviceAccounts["flyteworkers"].AccountId)
	ctx.Export("artifactregistryWriterServiceAccount", serviceAccounts["artifactregistry-writer"].AccountId)

	return serviceAccounts, nil
}

func createServiceAccounts(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) (map[string]struct {
	AccountId pulumi.StringOutput
	Member    pulumi.StringArrayOutput
}, error) {

	accounts := make(map[string]struct {
		AccountId pulumi.StringOutput
		Member    pulumi.StringArrayOutput
	})
	roles := []string{
		"flyteadmin",
		"flytepropeller",
		"flytescheduler",
		"datacatalog",
		"flyteworkers",
		"artifactregistry-writer",
	}

	for _, role := range roles {
		resourceName := fmt.Sprintf("%s-%s-iam-svc", projectConfig.ResourceNamePrefix, role)
		account, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
			AccountId: pulumi.String(fmt.Sprintf("flyte-%s", role)),
			Project:   pulumi.String(projectConfig.ProjectId),
		})
		if err != nil {
			return nil, err
		}

		member := account.Email.ApplyT(func(email string) []string {
			return []string{fmt.Sprintf("serviceAccount:%s", email)}
		}).(pulumi.StringArrayOutput)

		accounts[role] = struct {
			AccountId pulumi.StringOutput
			Member    pulumi.StringArrayOutput
		}{
			AccountId: account.AccountId,
			Member:    member,
		}
	}
	return accounts, nil
}

func createIAMBindings(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	serviceAccounts map[string]struct {
		AccountId pulumi.StringOutput
		Member    pulumi.StringArrayOutput
	},
) error {

	roles := map[string][]string{
		"flyteadmin": {
			"iam.serviceAccounts.signBlob",
			"storage.buckets.get",
			"storage.objects.create",
			"storage.objects.delete",
			"storage.objects.get",
			"storage.objects.getIamPolicy",
			"storage.objects.update",
		},
		"flytepropeller": {
			"storage.buckets.get",
			"storage.objects.create",
			"storage.objects.delete",
			"storage.objects.get",
			"storage.objects.list",
			"storage.objects.getIamPolicy",
			"storage.objects.update",
		},
		"flytescheduler": {
			"storage.buckets.get",
			"storage.objects.create",
			"storage.objects.delete",
			"storage.objects.get",
			"storage.objects.getIamPolicy",
			"storage.objects.update",
		},
		"datacatalog": {
			"storage.buckets.get",
			"storage.objects.create",
			"storage.objects.delete",
			"storage.objects.get",
			"storage.objects.update",
		},
		"flyteworkers": {
			"storage.buckets.get",
			"storage.objects.create",
			"storage.objects.delete",
			"storage.objects.get",
			"storage.objects.list",
			"storage.objects.update",
		},
	}

	for role, permissions := range roles {
		resourceName := fmt.Sprintf("%s-%s-iam-role", projectConfig.ResourceNamePrefix, role)
		_, err := projects.NewIAMCustomRole(ctx, resourceName, &projects.IAMCustomRoleArgs{
			Title:       pulumi.String(role),
			Permissions: pulumi.ToStringArray(permissions),
			Project:     pulumi.String(projectConfig.ProjectId),
		})
		if err != nil {
			return err
		}
	}

	for role := range serviceAccounts {
		resourceName := fmt.Sprintf("%s-%s-iam-role-binding", projectConfig.ResourceNamePrefix, role)
		_, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
			Project: pulumi.String(projectConfig.ProjectId),
			Role:    pulumi.String("projects/" + projectConfig.ProjectId + "/roles/" + role + "-role"),
			Members: serviceAccounts[role].Member,
		})
		if err != nil {
			return err
		}
	}

	// Additional IAM bindings for artifact registry access
	resourceName := fmt.Sprintf("%s-flyte-workers-iam-role-binding-registry", projectConfig.ResourceNamePrefix)
	_, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String("roles/artifactregistry.reader"),
		Members: serviceAccounts["flyteworkers"].Member,
	})
	if err != nil {
		return err
	}

	resourceName = fmt.Sprintf("%s-flyte-artifact-registry-writer", projectConfig.ResourceNamePrefix)
	_, err = projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String("roles/artifactregistry.writer"),
		Members: serviceAccounts["artifactregistry-writer"].Member,
	})
	if err != nil {
		return err
	}

	// Workload Identity Bindings
	ksaBindings := map[string]string{
		"flyteadmin":     "flyte/flyteadmin",
		"flytepropeller": "flyte/flytepropeller",
		"flytescheduler": "flyte/flytescheduler",
		"datacatalog":    "flyte/datacatalog",
	}

	identityNamespace := pulumi.Sprintf("%s.svc.id.goog", projectConfig.ProjectId)

	for role, ksa := range ksaBindings {
		resourceName := fmt.Sprintf("%s-%s-workload-identity-binding", projectConfig.ResourceNamePrefix, role)
		_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
			ServiceAccountId: serviceAccounts[role].AccountId,
			Role:             pulumi.String("roles/iam.workloadIdentityUser"),
			Members: pulumi.StringArray{
				pulumi.Sprintf("serviceAccount:%s[%s]", identityNamespace, ksa),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
