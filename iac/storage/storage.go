package storage

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SetupObjectStorage creates a GCS bucket and returns the outputs
func CreateObjectStorage(
	ctx *pulumi.Context,
	resourceNamePrefix string,
) error {

	resourceName := fmt.Sprintf("%s-data-bucket", resourceNamePrefix)
	// Create a Google Cloud Storage bucket
	bucket, err := storage.NewBucket(ctx, resourceName, &storage.BucketArgs{
		Location:                 pulumi.String("EU"),
		StorageClass:             pulumi.String("STANDARD"),
		ForceDestroy:             pulumi.Bool(true),
		UniformBucketLevelAccess: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}

	// Add a bucket IAM policy
	_, err = storage.NewBucketIAMMember(ctx, "bucket-iam-member", &storage.BucketIAMMemberArgs{
		Bucket: bucket.Name,
		Role:   pulumi.String("roles/storage.admin"),
		Member: pulumi.String("allAuthenticatedUsers"),
	})
	if err != nil {
		return err
	}

	ctx.Export("bucketName", bucket.Name)
	return nil
}
