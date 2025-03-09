package flyte

import (
	"encoding/json"
	"fmt"
	"mlops/global"
	"mlops/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func configureSAIAMPolicy(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
	serviceAccounts map[string]iam.ServiceAccountInfo,
) error {

	flyteProjects := []string{"flytesnacks"}
	flyteDomains := []string{"development", "staging", "production"}
	flyteKSAs := []string{"default"} // The KSA that Task Pods will use

	// Compute the cartesian product (setproduct) to generate worker WI members.
	var flyteWorkerWIMembers []string
	for _, proj := range flyteProjects {
		for _, dom := range flyteDomains {
			for _, ksa := range flyteKSAs {
				// Format as "project-domain/ksa", e.g., "flytesnacks-development/default"
				member := fmt.Sprintf("%s-%s/%s", proj, dom, ksa)
				flyteWorkerWIMembers = append(flyteWorkerWIMembers, member)
			}
		}
	}

	// Assume this comes from your GKE module (or set it manually).
	identityNamespace := projectConfig.ProjectId
	// Format each member as: "serviceAccount:<identity_namespace>.svc.id.goog[<member>]"
	var formattedMembers []string
	for _, m := range flyteWorkerWIMembers {
		formattedMember := fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s]", identityNamespace, m)
		formattedMembers = append(formattedMembers, formattedMember)
	}

	// Build the IAM policy JSON.
	policy := map[string]interface{}{
		"bindings": []map[string]interface{}{
			{
				"role":    "roles/iam.workloadIdentityUser",
				"members": formattedMembers,
			},
		},
	}
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	policyData := string(policyBytes)

	// Create the IAM policy for the service account.
	_, err = serviceaccount.NewIAMPolicy(ctx, "flyte-worker-workload-identity", &serviceaccount.IAMPolicyArgs{
		ServiceAccountId: serviceAccounts["flyteworkers"].ServiceAccount.ID(),
		PolicyData:       pulumi.String(policyData),
	})
	if err != nil {
		return err
	}
	return nil
}
