package storage

import (
	"fmt"
	"mlops/global"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SetupObjectStorage creates a GCS bucket and returns the outputs
func CreateObjectStorage(
	ctx *pulumi.Context,
	projectConfig global.ProjectConfig,
) pulumi.StringOutput {

	var bucketLocation string
	if len(projectConfig.EnabledRegions) > 1 {
		bucketLocation = "EU"
	} else {
		bucketLocation = projectConfig.EnabledRegions[0].Region
	}

	resourceName := fmt.Sprintf("%s-data-bucket", projectConfig.ResourceNamePrefix)
	bucket, err := storage.NewBucket(ctx, resourceName, &storage.BucketArgs{
		Location:                 pulumi.String(bucketLocation),
		StorageClass:             pulumi.String("STANDARD"),
		ForceDestroy:             pulumi.Bool(true),
		UniformBucketLevelAccess: pulumi.Bool(true),
	})
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Storage creation: %s", err), nil)
	}

	ctx.Export("bucketName", bucket.Name)
	return bucket.Name
}
