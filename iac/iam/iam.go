package iam

import (
	"fmt"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type svc struct {
	resourceNameSuffix string
	AccountId          string
	DisplayName        string
	Members            string
	Description        string
	Title              string
	IAMRoleId          string
	Permissions        pulumi.StringArray
	createRole         bool
}

var SVC = map[string]svc{
	"AutoNEG": {
		resourceNameSuffix: "autoneg",
		AccountId:          "autoneg-system",
		DisplayName:        "GKE at Scale - AutoNEG Service Account",
		Members:            "serviceAccount:autoneg-system@%s.iam.gserviceaccount.com",
		Description:        "Custom IAM Role - GKE AutoNeg",
		Title:              "GKE at Scale - AutoNEG",
		IAMRoleId:          "autoneg_system",
		Permissions: pulumi.StringArray{
			pulumi.String("compute.backendServices.get"),
			pulumi.String("compute.backendServices.update"),
			pulumi.String("compute.regionBackendServices.get"),
			pulumi.String("compute.regionBackendServices.update"),
			pulumi.String("compute.networkEndpointGroups.use"),
			pulumi.String("compute.healthChecks.useReadOnly"),
			pulumi.String("compute.regionHealthChecks.useReadOnly"),
		},
		createRole: true,
	},
	"Admin": {
		resourceNameSuffix: "admin",
		AccountId:          "svc-gke-at-scale-admin",
		DisplayName:        "GKE at Scale - Admin Service Account",
		createRole:         false,
	},
}

// CreateServiceAccount
func CreateServiceAccount(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	target string,
) (*serviceaccount.Account, error) {

	// Retrieve the corresponding service configuration from the map
	selectedSVC, found := SVC[target]
	if !found {
		return nil, fmt.Errorf("service '%s' not found", target)
	}

	gcpServiceAccount, err := createServiceAccount(ctx, resourceNamePrefix, &selectedSVC, gcpProjectId)
	if err != nil {
		return nil, err
	}
	if selectedSVC.createRole {
		gcpIAMRole, err := createIAMRole(ctx, resourceNamePrefix, &selectedSVC, gcpProjectId)
		if err != nil {
			return nil, err
		}
		_, err = createIAMRoleBinding(ctx, resourceNamePrefix, gcpProjectId, &selectedSVC, gcpIAMRole, gcpServiceAccount)
		if err != nil {
			return nil, err
		}
	}
	return gcpServiceAccount, nil
}

// createServiceAccount handles the creation of a Service Account
func createServiceAccount(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	svc *svc,
	gcpProjectId string,
) (*serviceaccount.Account, error) {

	resourceName := fmt.Sprintf("%s-svc-%s", resourceNamePrefix, svc.resourceNameSuffix)
	gcpServiceAccount, err := serviceaccount.NewAccount(ctx, resourceName, &serviceaccount.AccountArgs{
		Project:     pulumi.String(gcpProjectId),
		AccountId:   pulumi.String(svc.AccountId),
		DisplayName: pulumi.String(svc.DisplayName),
	})
	return gcpServiceAccount, err
}

// CreateIAMRole creates a Custom IAM Role that will be used by the Kubernetes Deployment.
// If svc selected is AutoNEG => This Role allows the AutoNeg CRD to link the Istio Ingress Gateway Service Ip to Load Balancer NEGs ( GCLB )
func createIAMRole(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	svc *svc,
	gcpProjectId string,
) (*projects.IAMCustomRole, error) {

	resourceName := fmt.Sprintf("%s-iam-custom-role-%s", resourceNamePrefix, svc.resourceNameSuffix)
	gcpIAMRole, err := projects.NewIAMCustomRole(ctx, resourceName, &projects.IAMCustomRoleArgs{
		Project:     pulumi.String(gcpProjectId),
		Description: pulumi.String(svc.Description),
		Permissions: svc.Permissions,
		RoleId:      pulumi.String(fmt.Sprintf("%s_iam_role_%s", resourceNamePrefix, svc.IAMRoleId)),
		Title:       pulumi.String(svc.Title),
	})
	return gcpIAMRole, err
}

// CreateIAMRoleBinding creates the IAM Role Binding to link to the Service Account to Custom Role.
func createIAMRoleBinding(
	ctx *pulumi.Context,
	resourceNamePrefix string,
	gcpProjectId string,
	svc *svc,
	gcpIAMRole *projects.IAMCustomRole,
	gcpServiceAccount *serviceaccount.Account,
) (*projects.IAMBinding, error) {

	resourceName := fmt.Sprintf("%s-iam-role-binding-%s", resourceNamePrefix, svc.resourceNameSuffix)
	gcpIAMRoleBinding, err := projects.NewIAMBinding(ctx, resourceName, &projects.IAMBindingArgs{
		Members: pulumi.StringArray{
			pulumi.String(fmt.Sprintf(svc.Members, gcpProjectId)),
		},
		Project: pulumi.String(gcpProjectId),
		Role:    gcpIAMRole.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{gcpServiceAccount}))

	return gcpIAMRoleBinding, err
}
