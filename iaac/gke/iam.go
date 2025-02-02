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
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// gkeIAM configures the GKE IAM permissions and Config Connector
func gkeConfigConnectorIAM(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	// Create IAM Service Account
	serviceAccount, serviceAccountMember, err := createConfigConnectorServiceAccount(ctx, projectConfig)
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, fmt.Errorf("failed to created Config Connector service account: %s", err)
	}
	// Create Workload Identity Pool
	err = CreateWorkloadIdentityPool(ctx, projectConfig)
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, fmt.Errorf("failed to created Workload Identity Pool: %s", err)
	}
	// Assign IAM "Editor" Role to the Service Account
	err = createIAMRoleBinding(ctx, projectConfig, "cc-editor", "roles/editor", nil, serviceAccountMember, true)
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, fmt.Errorf("failed to bind editor IAM role to Config Connector service account: %s", err)
	}
	// Assign "Workload Identity User" role to the service account
	err = createIAMRoleBinding(ctx, projectConfig, "cc-workload", "roles/iam.workloadIdentityUser", serviceAccount, serviceAccountMember, false)
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, fmt.Errorf("failed to bind Workload IAM role to Config Connector service account: %w", err)
	}
	return serviceAccount, serviceAccountMember, nil
}

// CreateWorkloadIdentityPool creates Google Cloud Workload Identity Pool for GKE
func CreateWorkloadIdentityPool(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) error {

	resourceName := fmt.Sprintf("%s-gke-wip", projectConfig.ResourceNamePrefix)

	// Create the Workload Identity Pool
	_, err := iam.NewWorkloadIdentityPool(ctx, resourceName, &iam.WorkloadIdentityPoolArgs{
		Project:                pulumi.String(projectConfig.ProjectId),
		Description:            pulumi.String("GKE - Workload Identity Pool"),
		Disabled:               pulumi.Bool(false),
		DisplayName:            pulumi.String("GKE Workload Identity Pool"),
		WorkloadIdentityPoolId: pulumi.String(fmt.Sprintf("%s-gke-pool", projectConfig.ResourceNamePrefix)),
	})
	if err != nil {
		return err
	}
	return nil
}

func createConfigConnectorServiceAccount(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) (*serviceaccount.Account, pulumi.StringArrayOutput, error) {

	resourceName := fmt.Sprintf("%s-cc-svc", projectConfig.ResourceNamePrefix)
	// Create the Workload Identity Pool
	serviceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("svc-config-connector"),
		DisplayName: pulumi.String("Config Connector IAM Service Account"),
		Project:     pulumi.String(projectConfig.ProjectId),
	})
	if err != nil {
		return nil, pulumi.StringArrayOutput{}, err
	}
	serviceAccountMember := serviceAccount.Email.ApplyT(func(email string) []string {
		return []string{fmt.Sprintf("serviceAccount:%s", email)}
	}).(pulumi.StringArrayOutput)

	return serviceAccount, serviceAccountMember, nil
}

// createIAMRoleBinding is a generic function to create IAM role bindings for both projects and service accounts
func createIAMRoleBinding(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	resourceNameSuffix string,
	role string,
	serviceAccount *serviceaccount.Account, // Optional for SA-specific roles
	serviceAccountMember pulumi.StringArrayOutput,
	isProjectRole bool, // true for project roles, false for SA roles
) error {

	// Define the resource name dynamically
	resourceName := fmt.Sprintf("%s-%s-bind", projectConfig.ResourceNamePrefix, resourceNameSuffix)

	if isProjectRole {
		// Create IAM binding at the project level
		_, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
			Project: pulumi.String(projectConfig.ProjectId),
			Role:    pulumi.String(role),
			Members: serviceAccountMember,
		})
		if err != nil {
			return err
		}
	} else {
		// Create IAM binding for a specific Service Account
		if serviceAccount == nil {
			return fmt.Errorf("serviceAccount cannot be nil when assigning IAM roles at the SA level")
		}

		_, err := serviceaccount.NewIAMBinding(ctx, resourceName, &serviceaccount.IAMBindingArgs{
			ServiceAccountId: serviceAccount.Name, // Use Name reference
			Role:             pulumi.String(role),
			Members:          serviceAccountMember,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func createNamespace(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	namespaceName string,
	k8sProvider *kubernetes.Provider,
) error {
	resourceName := fmt.Sprintf("%s-cc-namespace", projectConfig.ResourceNamePrefix)

	_, err := corev1.NewNamespace(ctx, resourceName, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(namespaceName),
			Annotations: pulumi.StringMap{
				"cnrm.cloud.google.com/project-id": pulumi.String(projectConfig.ProjectId),
			},
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return fmt.Errorf("failed to create namespace: %s", err)
	}
	return nil
}

func applyResource(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
	serviceAccountMember pulumi.StringArrayOutput,
	k8sProvider *kubernetes.Provider,
) error {

	resourceName := fmt.Sprintf("%s-cc-config", projectConfig.ResourceNamePrefix)

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
		}, pulumi.Provider(k8sProvider))
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
