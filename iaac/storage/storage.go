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
	bucketName string,
) *storage.Bucket {

	var bucketLocation string
	if len(projectConfig.EnabledRegions) > 1 {
		bucketLocation = "EU"
	} else {
		bucketLocation = projectConfig.EnabledRegions[0].Region
	}

	resourceName := fmt.Sprintf("%s-data-bucket", projectConfig.ResourceNamePrefix)
	if bucketName == "" {
		bucketName = resourceName
	} else {
		resourceName = fmt.Sprintf("%s-%s", projectConfig.ResourceNamePrefix, bucketName)
	}

	bucket, err := storage.NewBucket(ctx, resourceName, &storage.BucketArgs{
		Name:                     pulumi.String(bucketName),
		Location:                 pulumi.String(bucketLocation),
		StorageClass:             pulumi.String("STANDARD"),
		ForceDestroy:             pulumi.Bool(true),
		UniformBucketLevelAccess: pulumi.Bool(true),
	})
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Storage creation: %s", err), nil)
	}
	return bucket
}
