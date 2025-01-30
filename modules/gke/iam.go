package gke

// https://cloud.google.com/config-connector/docs/how-to/install-upgrade-uninstall#creating_a_new_cluster_with_the_enabled
// * IAM Service Account creation
// * IAM Role assignments
// * Workload Identity binding
// * Namespace creation and annotation
// * Config Connector installation in cluster mode

import (
	"bytes"
	"errors"
	"fmt"
	"mlops/project"
	"os/exec"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// gkeIAM configures the GKE IAM permissions and Config Connector
func gkeConfigConnectorIAM(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) error {

	namespaceName := "config-connector"

	// 1️⃣ Create IAM Service Account
	serviceAccount, serviceAccountMember, err := createConfigConnectorServiceAccount(ctx, projectConfig)
	if err != nil {
		return err
	}
	// 2️⃣ Assign IAM "Editor" Role to the Service Account (adjust if needed)
	err = createIAMRoleBinding(ctx, projectConfig, serviceAccountMember)
	if err != nil {
		return nil
	}
	// 3️⃣ Create Workload Identity Pool
	err = CreateWorkloadIdentityPool(ctx, projectConfig)
	if err != nil {
		return err
	}
	// 4️⃣ Bind IAM Service Account
	err = createIAMBinding(ctx, projectConfig, serviceAccount, serviceAccountMember)
	if err != nil {
		return err
	}
	// 5️⃣ Create Namespace for Config Connector
	err = createNamespace(ctx, projectConfig, namespaceName)
	if err != nil {
		return err
	}
	// 6️⃣ Apply Config Connector Configuration (Cluster Mode)
	err = applyResource(ctx, projectConfig, serviceAccountMember)
	if err != nil {
		return fmt.Errorf("failed to apply Config Connector configuration: %w", err)
	}

	// ctx.Export("serviceAccountEmail", serviceAccount.Email)
	// ctx.Export("workloadIdentityPool", identityPool.WorkloadIdentityPoolId)
	// ctx.Export("workloadIdentityProvider", identityPoolProvider.Name)

	return nil
}

// CreateWorkloadIdentityPool creates Google Cloud Workload Identity Pool for GKE
func CreateWorkloadIdentityPool(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) error {

	// Generate a unique suffix for the Workload Identity Pool ID
	randomSuffix := generateRandomString(6)
	resourceName := fmt.Sprintf("%s-wip-gke-cluster", projectConfig.ResourceNamePrefix)

	// Create the Workload Identity Pool
	_, err := iam.NewWorkloadIdentityPool(ctx, resourceName, &iam.WorkloadIdentityPoolArgs{
		Project:                pulumi.String(projectConfig.ProjectId),
		Description:            pulumi.String("GKE - Workload Identity Pool"),
		Disabled:               pulumi.Bool(false),
		DisplayName:            pulumi.String(resourceName),
		WorkloadIdentityPoolId: pulumi.String(fmt.Sprintf("%s-%s", projectConfig.ResourceNamePrefix, randomSuffix)),
	})
	if err != nil {
		return fmt.Errorf("failed to create Workload Identity Pool %s: %w", resourceName, err)
	}
	return nil
}

func createConfigConnectorServiceAccount(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	resourceName := fmt.Sprintf("%s-svc-config-connector", projectConfig.ResourceNamePrefix)
	// Create the Workload Identity Pool
	serviceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String(resourceName),
		DisplayName: pulumi.String("Config Connector IAM Service Account"),
		Project:     pulumi.String(projectConfig.ProjectId),
	})
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, fmt.Errorf("failed to create IAM service account: %w", err)
	}
	serviceAccountMember := serviceAccount.Email.ApplyT(func(email string) []string {
		return []string{fmt.Sprintf("serviceAccount:%s", email)}
	}).(pulumi.StringArrayOutput)

	return serviceAccount, serviceAccountMember, nil
}

func createIAMRoleBinding(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	serviceAccountMember pulumi.StringArrayOutput,
) error {

	resourceName := fmt.Sprintf("%s-config-connector-editor-role", projectConfig.ResourceNamePrefix)

	// Create the Config Connector Role Binding
	_, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Project: pulumi.String(projectConfig.ProjectId),
		Role:    pulumi.String("roles/editor"),
		Members: serviceAccountMember,
	})
	if err != nil {
		return fmt.Errorf("failed to assign editor role: %w", err)
	}
	return nil
}

func createIAMBinding(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	serviceAccount *serviceaccount.Account,
	serviceAccountMember pulumi.StringArrayOutput,
) error {
	resourceName := fmt.Sprintf("%s-config-connector-sa-iam-binding", projectConfig.ResourceNamePrefix)

	// members := serviceAccount.Email.ApplyT(func(email string) []string {
	// 	return []string{fmt.Sprintf("serviceAccount:%s", email)}
	// }).(pulumi.StringArrayOutput)

	// ✅ Pass `members` correctly into IAMBinding
	_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
		ServiceAccountId: serviceAccount.Name, // Use Name as reference
		Role:             pulumi.String("roles/iam.workloadIdentityUser"),
		Members:          serviceAccountMember,
	})
	if err != nil {
		return fmt.Errorf("failed to bind IAM role to service account: %w", err)
	}

	return nil
}

func createNamespace(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	namespaceName string,
) error {
	resourceName := fmt.Sprintf("%s-config-connector-namespace", projectConfig.ResourceNamePrefix)

	_, err := corev1.NewNamespace(ctx, resourceName, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(namespaceName),
			Annotations: pulumi.StringMap{
				"cnrm.cloud.google.com/project-id": pulumi.String(projectConfig.ProjectId),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %s", err)
	}
	return nil
}

func applyResource(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	serviceAccountMember pulumi.StringArrayOutput,
) error {

	resourceName := fmt.Sprintf("%s-config-connector-config", projectConfig.ResourceNamePrefix)

	configConnectorYaml := serviceAccountMember.ApplyT(func(members []string) string {
		if len(members) > 0 {
			return fmt.Sprintf(`
apiVersion: core.cnrm.cloud.google.com/v1beta1
kind: ConfigConnector
metadata:
  name: configconnector.core.cnrm.cloud.google.com
spec:
  mode: cluster
  googleServiceAccount: %s
  stateIntoSpec: Absent
`, members[0])
		}
		return "" // Handle empty array case
	}).(pulumi.StringOutput)

	configConnectorYaml.ApplyT(func(yamlContent string) (*yaml.ConfigGroup, error) {
		if yamlContent == "" {
			return nil, fmt.Errorf("empty YAML content")
		}
		return yaml.NewConfigGroup(ctx, resourceName, &yaml.ConfigGroupArgs{
			YAML: []string{yamlContent},
		})
	})
	return nil
}

func runCommand(ctx *pulumi.Context, args []string) {
	if len(args) == 0 {
		_ = ctx.Log.Error("❌ No command provided", nil)
		return
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdOutput

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	if err := cmd.Run(); err != nil {
		_ = ctx.Log.Error(fmt.Sprintf("❌ Command failed: %s\nError: %s", args, err), nil)
		return
	}

	ctx.Log.Info(fmt.Sprintf("✅ Command executed successfully:\n%s", cmdOutput.String()), nil)
}
