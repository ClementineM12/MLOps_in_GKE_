
import kagglehub
import os 
from google.cloud import storage

def fetch_dataset():
    """
    Check if the dataset exists in a GCS bucket. 
    If not, download it from Kaggle and upload the metadata CSV and images to the bucket.
    """
    bucket_name = "datasets"
    metadata_file = "HAM10000_metadata.csv"
    images_prefix = "images/"

    # Initialize GCS client
    client = storage.Client()

    # Get bucket reference and check existence
    bucket = client.bucket(bucket_name)
    if bucket.exists():
        print("Dataset already exists in GCS. Skipping upload.")
    else:
        print("Dataset not found in GCS. Downloading from Kaggle...")
        # Download dataset from Kaggle
        dataset_path = kagglehub.dataset_download("kmader/skin-cancer-mnist-ham10000")
        
        # Create GCS bucket
        bucket = client.create_bucket(bucket_name)
        
        # Upload metadata CSV
        local_metadata_path = os.path.join(dataset_path, metadata_file)
        blob = bucket.blob(metadata_file)
        blob.upload_from_filename(local_metadata_path)
        
        # Upload images (assumes images are in the dataset_path directory)
        for image_file in os.listdir(dataset_path):
            if image_file.endswith(".jpg"):
                local_image_path = os.path.join(dataset_path, image_file)
                blob = bucket.blob(f"{images_prefix}{image_file}")
                blob.upload_from_filename(local_image_path)
                
        print("Dataset successfully uploaded to GCS.")

