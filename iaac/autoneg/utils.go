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
			err = createClusterRoleBindings(ctx, projectConfig, k8sProvider, serviceAccount, roleDef.Name, clusterRole)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createClusterRoleBindings(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
	clusterRole string,
	clusterRoleResource *rbacV1.ClusterRole,
) error {

	resourceName := fmt.Sprintf("%s-%s-binding", projectConfig.ResourceNamePrefix, clusterRole)
	_, err := rbacV1.NewClusterRoleBinding(ctx, resourceName, &rbacV1.ClusterRoleBindingArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-binding", clusterRole)),
			Labels: pulumi.StringMap{
				"app": pulumi.String("autoneg"),
			},
			Annotations: pulumi.StringMap{}, // Empty annotations
		},
		RoleRef: &rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("ClusterRole"),
			Name:     pulumi.String(clusterRole),
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
		pulumi.DependsOn([]pulumi.Resource{clusterRoleResource}),
	)

	if err != nil {
		return fmt.Errorf("failed to create cluster role binding %s: %w", clusterRole, err)
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
			err = createRoleBindings(ctx, projectConfig, k8sProvider, serviceAccount, roleDef.Name, role)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createRoleBindings(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	k8sProvider *kubernetes.Provider,
	serviceAccount *coreV1.ServiceAccount,
	role string,
	roleResource *rbacV1.Role,
) error {

	resourceName := fmt.Sprintf("%s-%s-binding", projectConfig.ResourceNamePrefix, role)
	_, err := rbacV1.NewRoleBinding(ctx, resourceName, &rbacV1.RoleBindingArgs{
		Metadata: &metaV1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("%s-binding", role)),
			Labels: pulumi.StringMap{
				"app": pulumi.String("autoneg"),
			},
			Annotations: pulumi.StringMap{}, // Empty annotations
		},
		RoleRef: &rbacV1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String(role),
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
		pulumi.DependsOn([]pulumi.Resource{roleResource}),
	)

	if err != nil {
		return fmt.Errorf("failed to create role binding %s: %w", role, err)
	}
	return nil
}
