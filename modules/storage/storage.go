package storage

import (
	"fmt"
	"mlops/project"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SetupObjectStorage creates a GCS bucket and returns the outputs
func CreateObjectStorage(
	ctx *pulumi.Context,
	projectConfig project.ProjectConfig,
) {

	var bucketLocation string
	if len(projectConfig.EnabledRegions) > 1 {
		bucketLocation = "EU"
	} else {
		bucketLocation = projectConfig.EnabledRegions[0].Region
	}

	resourceName := fmt.Sprintf("%s-data-bucket", projectConfig.ResourceNamePrefix)
	// Create a Google Cloud Storage bucket
	bucket, err := storage.NewBucket(ctx, resourceName, &storage.BucketArgs{
		Location:                 pulumi.String(bucketLocation),
		StorageClass:             pulumi.String("STANDARD"),
		ForceDestroy:             pulumi.Bool(true),
		UniformBucketLevelAccess: pulumi.Bool(true),
	})
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Storage creation: %s", err), nil)
	}

	// Add a bucket IAM policy
	_, err = storage.NewBucketIAMMember(ctx, "bucket-iam-member", &storage.BucketIAMMemberArgs{
		Bucket: bucket.Name,
		Role:   pulumi.String("roles/storage.admin"),
		Member: pulumi.String("allAuthenticatedUsers"),
	})
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Storage IAM creation: %s", err), nil)
	}

	ctx.Export("bucketName", bucket.Name)
}
