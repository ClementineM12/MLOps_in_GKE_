import argparse
import os
from google.cloud import storage
import kagglehub


def fetch_dataset(
    retrieved_metadata_filename: str, 
    metadata_filename: str,
    bucket: str, 
    images_dir: str,
    data_path: str,
) -> None:
    """
    Check if the dataset exists in a GCP bucket.
    If not, download it from Kaggle and upload the metadata CSV and images to the bucket under a given data_path.
    
    Args:
        retrieved_metadata_filename (str): The name of the metadata CSV file.
        metadata_filename (str): The name we will set for the metadata CSV file.
        bucket (str): The name of the GCP bucket.
        images_dir (str): The folder name to store images within the data_path.
        data_path (str): The folder path in the bucket where all data should be uploaded.
        
    Returns:
        None
    """
    # Initialize the GCS client.
    client = storage.Client()
    
    # Get the bucket. BUCKET MUST BE CREATED
    bucket_obj = storage.Bucket(client, bucket)
    # Download the dataset from Kaggle.
    dataset_path = kagglehub.dataset_download("kmader/skin-cancer-mnist-ham10000")
    
    # Define the destination path for metadata CSV inside the bucket.
    metadata_object_name = os.path.join(data_path, metadata_filename)
    local_metadata_path = os.path.join(dataset_path, retrieved_metadata_filename)
    
    # Upload metadata CSV
    try:
        blob = bucket_obj.blob(metadata_object_name)
        blob.upload_from_filename(local_metadata_path)
        print(f"Uploaded metadata CSV to: {metadata_object_name}")
    except Exception as err:
        print(f"Error uploading metadata CSV: {err}")
        raise
    
    # Recursively upload images from all subdirectories.
    for root, dirs, files in os.walk(dataset_path):
        for image_file in files:
            if image_file.lower().endswith(".jpg"):
                local_image_path = os.path.join(root, image_file)
                # Upload images to a specific folder inside the given data_path.
                object_name = os.path.join(data_path, images_dir, image_file)
                try:
                    blob = bucket_obj.blob(object_name)
                    blob.upload_from_filename(local_image_path)
                except Exception as err:
                    print(f"Error uploading image {image_file}: {err}")
                    raise
    
    print("Dataset successfully uploaded to GCP bucket.")
    return


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Load data into GCS")
    parser.add_argument("--retrieved_metadata_filename", type=str, required=True, help="Filename for the retrieved metadata")
    parser.add_argument("--metadata_filename", type=str, required=True, help="Metadata filename")
    parser.add_argument("--bucket", type=str, required=True, help="GCS bucket name")
    parser.add_argument("--images_dir", type=str, required=True, help="Images directory")
    parser.add_argument("--data_path", type=str, required=True, help="Data path")
    
    args = parser.parse_args()
    
    fetch_dataset(
        retrieved_metadata_filename=args.retrieved_metadata_filename,
        metadata_filename=args.metadata_filename,
        bucket=args.bucket,
        images_dir=args.images_dir,
        data_path=args.data_path
    )