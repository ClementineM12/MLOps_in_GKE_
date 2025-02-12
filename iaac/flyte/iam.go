package database

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Project & Namespace
		projectID := "your-gcp-project-id"
		identityNamespace := "your-identity-namespace" // e.g., "your-project.svc.id.goog"

		// List of Google Service Accounts (GSAs) for Flyte
		gsaNames := []string{
			"flyteadmin", "flytepropeller", "flytescheduler",
			"datacatalog", "flyteworkers", "artifactregistry-writer",
		}

		// Create GSAs
		gsaMap := make(map[string]*serviceaccount.Account)
		for _, gsa := range gsaNames {
			account, err := serviceaccount.NewAccount(ctx, gsa, &serviceaccount.AccountArgs{
				AccountId: pulumi.String(fmt.Sprintf("%s-%s", "flyte", gsa)),
				Project:   pulumi.String(projectID),
			})
			if err != nil {
				return err
			}
			gsaMap[gsa] = account
		}

		// Define IAM Role Permissions
		rolePermissions := map[string][]string{
			"flyteadmin": {
				"iam.serviceAccounts.signBlob",
				"storage.buckets.get", "storage.objects.create",
				"storage.objects.delete", "storage.objects.get",
				"storage.objects.getIamPolicy", "storage.objects.update",
			},
			"flytepropeller": {
				"storage.buckets.get", "storage.objects.create",
				"storage.objects.delete", "storage.objects.get",
				"storage.objects.list", "storage.objects.getIamPolicy",
				"storage.objects.update",
			},
			"flytescheduler": {
				"storage.buckets.get", "storage.objects.create",
				"storage.objects.delete", "storage.objects.get",
				"storage.objects.getIamPolicy", "storage.objects.update",
			},
			"datacatalog": {
				"storage.buckets.get", "storage.objects.create",
				"storage.objects.delete", "storage.objects.get",
				"storage.objects.update",
			},
			"flyteworkers": {
				"storage.buckets.get", "storage.objects.create",
				"storage.objects.delete", "storage.objects.get",
				"storage.objects.list", "storage.objects.update",
			},
		}

		// Create Custom IAM Roles
		roleMap := make(map[string]*iam.CustomRole)
		for roleName, permissions := range rolePermissions {
			role, err := iam.NewCustomRole(ctx, roleName, &iam.CustomRoleArgs{
				Project:     pulumi.String(projectID),
				RoleId:      pulumi.String(roleName),
				Title:       pulumi.String(roleName),
				Permissions: pulumi.ToStringArray(permissions),
			})
			if err != nil {
				return err
			}
			roleMap[roleName] = role
		}

		// Bind IAM Roles to GSAs
		for roleName, role := range roleMap {
			_, err := iam.NewMember(ctx, fmt.Sprintf("%s-binding", roleName), &iam.MemberArgs{
				Project: pulumi.String(projectID),
				Role:    role.Name,
				Member:  pulumi.Sprintf("serviceAccount:%s", gsaMap[roleName].Email),
			})
			if err != nil {
				return err
			}
		}

		// Artifact Registry Binding for Flyte Workers
		_, err := iam.NewMember(ctx, "flyteworkers-registry", &iam.MemberArgs{
			Project: pulumi.String(projectID),
			Role:    pulumi.String("roles/artifactregistry.reader"),
			Member:  pulumi.Sprintf("serviceAccount:%s", gsaMap["flyteworkers"].Email),
		})
		if err != nil {
			return err
		}

		// Artifact Registry Writer Binding
		_, err = iam.NewMember(ctx, "artifactregistry-writer", &iam.MemberArgs{
			Project: pulumi.String(projectID),
			Role:    pulumi.String("roles/artifactregistry.writer"),
			Member:  pulumi.Sprintf("serviceAccount:%s", gsaMap["artifactregistry-writer"].Email),
		})
		if err != nil {
			return err
		}

		// Workload Identity Bindings (KSA to GSA)
		ksaBindings := map[string]string{
			"flyteadmin":     "flyte/flyteadmin",
			"flytepropeller": "flyte/flytepropeller",
			"flytescheduler": "flyte/flytescheduler",
			"datacatalog":    "flyte/datacatalog",
			"flyteworkers":   "flyte/default",
		}

		for gsaName, ksa := range ksaBindings {
			_, err := iam.NewMember(ctx, fmt.Sprintf("%s-wi-binding", gsaName), &iam.MemberArgs{
				Project: pulumi.String(projectID),
				Role:    pulumi.String("roles/iam.workloadIdentityUser"),
				Member:  pulumi.Sprintf("serviceAccount:%s[%s]", identityNamespace, ksa),
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}
