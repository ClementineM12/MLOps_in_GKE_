package autoneg

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	coreV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metaV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	rbacV1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createClusterRoles(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
) error {

	for _, roleDef := range ClusterRoles {
		resourceName := fmt.Sprintf("%s-%s", projectConfig.ResourceNamePrefix, roleDef.Name)
		clusterRole, err := rbacV1.NewClusterRole(ctx, resourceName, &rbacV1.ClusterRoleArgs{
			Metadata: &metaV1.ObjectMetaArgs{
				Name: pulumi.String(roleDef.Name),
				Labels: pulumi.StringMap{
					"app": pulumi.String("autoneg"),
				},
			},
			Rules: roleDef.RBAC,
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return fmt.Errorf("failed to create cluster role %s: %w", roleDef.Name, err)
		}

		if roleDef.Bind {
			err = createClusterRoleBindings(ctx, projectConfig, k8sProvider, serviceAccount, clusterRole)
		}
	}

	return nil
}

func createClusterRoleBindings(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
	clusterRole *rbacV1.ClusterRole,
) error {

	resourceName := fmt.Sprintf("%s-%s-binding", projectConfig.ResourceNamePrefix, clusterRole.Metadata.Name())
	_, err := rbacV1.NewClusterRoleBinding(ctx, resourceName, &rbacV1.ClusterRoleBindingArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-binding", clusterRole.Metadata.Name())),
			Labels: pulumi.StringMap{
				"app": pulumi.String("autoneg"),
			},
			Annotations: pulumi.StringMap{}, // Empty annotations
		},
		RoleRef: &rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("ClusterRole"),
			Name:     clusterRole.Metadata.Name().Elem(),
		},
		Subjects: rbacV1.SubjectArray{
			&rbacV1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      serviceAccount.Metadata.Name().Elem(),
				Namespace: serviceAccount.Metadata.Namespace(),
			},
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{clusterRole}),
	)

	if err != nil {
		return fmt.Errorf("failed to create cluster role binding %s: %w", clusterRole.ID(), err)
	}
	return nil
}

func createRoles(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
) error {

	for _, roleDef := range Roles {
		resourceName := fmt.Sprintf("%s-%s", projectConfig.ResourceNamePrefix, roleDef.Name)
		role, err := rbacV1.NewRole(ctx, resourceName, &rbacV1.RoleArgs{
			Metadata: &metaV1.ObjectMetaArgs{
				Name:      pulumi.String(roleDef.Name),
				Namespace: serviceAccount.Metadata.Namespace(),
				Labels: pulumi.StringMap{
					"app": pulumi.String("autoneg"),
				},
			},
			Rules: roleDef.RBAC,
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return fmt.Errorf("failed to create cluster role %s: %w", roleDef.Name, err)
		}

		if roleDef.Bind {
			err = createRoleBindings(ctx, projectConfig, k8sProvider, serviceAccount, role)
		}
	}

	return nil
}

func createRoleBindings(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
	role *rbacV1.Role,
) error {

	resourceName := fmt.Sprintf("%s-%s-binding", projectConfig.ResourceNamePrefix, role.Metadata.Name())
	_, err := rbacV1.NewRoleBinding(ctx, resourceName, &rbacV1.RoleBindingArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-binding", role.Metadata.Name())),
			Labels: pulumi.StringMap{
				"app": pulumi.String("autoneg"),
			},
			Annotations: pulumi.StringMap{}, // Empty annotations
		},
		RoleRef: &rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     role.Metadata.Name().Elem(),
		},
		Subjects: rbacV1.SubjectArray{
			&rbacV1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      serviceAccount.Metadata.Name().Elem(),
				Namespace: serviceAccount.Metadata.Namespace(),
			},
		},
	},
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn([]pulumi.Resource{role}),
	)

	if err != nil {
		return fmt.Errorf("failed to create role binding %s: %w", role.ID(), err)
	}
	return nil
}
