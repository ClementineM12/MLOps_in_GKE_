from google.cloud import storage
import pandas as pd
import io 

def get_metadata_file(
    bucket: str, 
    metadata_filename: str,
    data_path: str,
    target: str = "csv",
) -> pd.DataFrame:
    """
    Retrieve the metadata file from the source GCS bucket.
    
    Args:
        bucket (str): The GCS bucket name.
        metadata_filename (str): The name of the metadata file.
        target (str): The format of the file to load ('csv' or 'pickle').
        data_path (str): The base path in the bucket where the file is stored, if any.
        
    Returns:
        pd.DataFrame: The loaded metadata DataFrame.
    """
    try:
        storage_client = storage.Client()
        bucket_obj = storage_client.bucket(bucket)
        # Combine data_path and metadata_filename if a data_path is provided.
        object_name = f"{data_path}/{metadata_filename}" if data_path else metadata_filename
        blob = bucket_obj.blob(object_name)
        data = blob.download_as_bytes()

        if target == "csv":
            df = pd.read_csv(io.StringIO(data.decode('utf-8')))
        elif target == "pickle":
            df = pd.read_pickle(io.BytesIO(data))
        else:
            raise ValueError(f"Unsupported target type: {target}")

        print(f"Retrieved metadata file '{metadata_filename}' from bucket '{bucket}'.")
        return df
    except Exception as err:
        print(f"Error retrieving metadata file from bucket: {err}")
        raise

def put_metadata_file(
    metadata: pd.DataFrame,
    bucket: str, 
    metadata_filename: str,
    data_path: str,
) -> None:
    """
    Uploads a metadata pickle file to a GCP bucket.
    
    Args:
        metadata (pd.DataFrame): The DataFrame to upload.
        bucket (str): The name of the GCP bucket.
        metadata_filename (str): The target filename in the bucket.
        data_path (str): The folder path in the bucket where the file should be uploaded.
    """
    # Initialize the GCS client.
    storage_client = storage.Client()
    bucket_obj = storage_client.bucket(bucket)
    
    # Construct the object name.
    object_name = f"{data_path}/{metadata_filename}" if data_path else metadata_filename

    # Write the DataFrame to a BytesIO buffer in pickle format.
    pickle_buffer = io.BytesIO()
    metadata.to_pickle(pickle_buffer)
    pickle_buffer.seek(0)
    
    try:
        blob = bucket_obj.blob(object_name)
        blob.upload_from_file(pickle_buffer, content_type="application/octet-stream")
        print(f"Uploaded processed metadata pickle file to '{object_name}' in bucket '{bucket}'.")
    except Exception as err:
        print(f"Error uploading processed metadata pickle file: {err}")
        raise