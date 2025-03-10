package infracomponents

// import (
// 	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
// 	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
// 	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
// )

// // deployMinioHelm deploys MinIO using its official Helm chart.
// func deployMinioHelm(
// 	ctx *pulumi.Context,
// 	k8sProvider *kubernetes.Provider,
// ) error {
// 	// Deploy the MinIO Helm chart via OCI
// 	_, err := helm.NewRelease(ctx, "minio", &helm.ReleaseArgs{
// 		Name:            pulumi.String("minio"),
// 		Namespace:       pulumi.String("minio"),
// 		CreateNamespace: pulumi.Bool(true),
// 		Chart:           pulumi.String("minio"),
// 		Version:         pulumi.String("15.0.6"),
// 		RepositoryOpts: &helm.RepositoryOptsArgs{
// 			Repo: pulumi.String("oci://registry-1.docker.io/bitnamicharts"),
// 		},
// 		Timeout: pulumi.Int(300),
// 	}, pulumi.Provider(k8sProvider))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
