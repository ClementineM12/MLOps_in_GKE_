from minio import Minio
import kagglehub
import os 

# ---------------------
# Data Retrieval from Kaggle & MinIO
# ---------------------
def fetch_dataset():
    """Check if dataset exists in MinIO. If not, download from Kaggle and upload to MinIO."""
    minio_client = Minio(
        "minio:9000",
        access_key="jupyter_user",
        secret_key="12341234",
        secure=False
    )
    
    bucket_name = "datasets"
    metadata_file = "HAM10000_metadata.csv"
    images_folder = "HAM10000_images_part_1"
    images_prefix = "images/"
    
    # Check if bucket exists
    found = minio_client.bucket_exists(bucket_name)
    if found:
        print("Dataset already exists in MinIO. Skipping upload.")
    else:
        print("Dataset not found in MinIO. Downloading from Kaggle...")
        dataset_path = kagglehub.dataset_download("kmader/skin-cancer-mnist-ham10000")
        
        # Create bucket
        minio_client.make_bucket(bucket_name)
        
        # Upload metadata CSV
        local_metadata_path = os.path.join(dataset_path, metadata_file)
        minio_client.fput_object(bucket_name, metadata_file, local_metadata_path)
        
        # Upload images
        image_folder = dataset_path
        for image_file in os.listdir(image_folder):
            if image_file.endswith(".jpg"):
                local_image_path = os.path.join(image_folder, image_file)
                minio_image_path = f"{images_prefix}{image_file}"
                minio_client.fput_object(bucket_name, minio_image_path, local_image_path)
                
        print("Dataset successfully uploaded to MinIO.")

fetch_dataset()
