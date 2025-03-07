import os
from minio import Minio
from minio.error import S3Error
import kagglehub  # Adjust the import as needed if you're using a specific Kaggle module

# Create MinIO client
minio_client = Minio(
    "minio.mlrun.svc.cluster.local:9000",
    access_key="mlrun",
    secret_key="mlrun1234",
    secure=False
)

def fetch_dataset(
    metadata_filename: str, 
    bucket_name: str,
    images_dir: str, 
) -> str:
    """
    Check if the dataset exists in a MinIO bucket.
    If not, download it from Kaggle and upload the metadata CSV and images to the bucket.
    
    Args:
        bucket_name (str): The name of the MinIO bucket.
        
    Returns:
        The local path to the downloaded dataset.
    """
    
    # Check if the bucket exists
    try:
        bucket_exists = minio_client.bucket_exists(bucket_name)
    except S3Error as err:
        print(f"Error checking bucket existence: {err}")
        bucket_exists = False

    if bucket_exists:
        print("Dataset already exists in MinIO. Skipping upload.")
        return None
    else:
        print("Dataset not found in MinIO. Downloading from Kaggle...")
        # Download the dataset from Kaggle. Adjust this call as needed.
        dataset_path = kagglehub.dataset_download("kmader/skin-cancer-mnist-ham10000")
        
        # Create the bucket in MinIO
        try:
            minio_client.make_bucket(bucket_name)
            print(f"Bucket '{bucket_name}' created.")
        except S3Error as err:
            print(f"Error creating bucket: {err}")
            raise
        
        # Upload metadata CSV
        local_metadata_path = os.path.join(dataset_path, metadata_filename)
        try:
            minio_client.fput_object(bucket_name, metadata_filename, local_metadata_path)
            print(f"Uploaded metadata CSV: {metadata_filename}")
        except S3Error as err:
            print(f"Error uploading metadata CSV: {err}")
            raise
        
        # Recursively upload images from all subdirectories
        for root, dirs, files in os.walk(dataset_path):
            for image_file in files:
                if image_file.lower().endswith(".jpg"):
                    local_image_path = os.path.join(root, image_file)
                    # All images are uploaded to the same bucket folder, ignoring subdirectory structure.
                    object_name = f"{images_dir}/{image_file}"
                    try:
                        minio_client.fput_object(bucket_name, object_name, local_image_path)
                    except S3Error as err:
                        print(f"Error uploading image {image_file}: {err}")
                        raise

        print("Dataset successfully uploaded to MinIO.")
        return dataset_path
