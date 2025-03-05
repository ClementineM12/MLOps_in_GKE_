import os
import io
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.utils import resample
import cv2
import numpy as np
from PIL import Image
from minio import Minio
from minio.error import S3Error
from io import BytesIO

# Creating dictionary for displaying more human-friendly labels
lesion_type_dict = {
    'nv': 'Melanocytic nevi',
    'mel': 'Melanoma',
    'bkl': 'Benign keratosis-like lesions',
    'bcc': 'Basal cell carcinoma',
    'akiec': 'Actinic keratoses',
    'vasc': 'Vascular lesions',
    'df': 'Dermatofibroma'
}
# Create MinIO client
minio_client = Minio(
    "minio.mlrun.svc.cluster.local:9000",
    access_key="mlrun",
    secret_key="mlrun1234",
    secure=False
)
    
def process_metadata(
    metadata_filename: str, 
    source_bucket: str, 
    processed_bucket: str,
    images_dir: str,
    processed_metadata_filename: str
) -> str:
    """
    Process the skin metadata and images stored in a source MinIO bucket,
    then upload the processed (segmented) images to another bucket.
    
    Args:
        skin_meta_filename (str): The CSV file name containing metadata.
        source_bucket (str): Name of the MinIO bucket with raw data.
        processed_bucket (str): Name of the MinIO bucket for processed images.
    """
    
    # Ensure that the processed bucket exists (create if necessary)
    try:
        bucket_exists = minio_client.bucket_exists(processed_bucket)
    except S3Error as err:
        print(f"Error checking bucket existence: {err}")
        bucket_exists = False
        
    if not bucket_exists:
        minio_client.make_bucket(processed_bucket)
        print(f"Created processed bucket: {processed_bucket}")
    else:
        print(f"Processed bucket '{processed_bucket}' already exists.")

    # Retrieve the metadata DataFrame from the bucket.
    skin_meta = get_metadata_file(source_bucket, metadata_filename)
    
    # (Optional) Retrieve raw images from source bucket.
    # Assume that raw images are stored under a prefix "images/" and that the metadata CSV
    # has an 'image_id' column which corresponds to the filename (without extension).
    def get_image_from_minio(image_id: str):
        object_name = f"{images_dir}/{image_id}.jpg"
        try:
            response = minio_client.get_object(source_bucket, object_name)
            image = Image.open(BytesIO(response.read()))
            return image
        except Exception as e:
            print(f"Error retrieving image '{object_name}': {e}")
            return None

    # Add a column with the downloaded PIL images
    skin_meta['image'] = skin_meta['image_id'].apply(get_image_from_minio)
    
    # Create additional columns for better understanding of features
    skin_meta['Cell type'] = skin_meta['dx'].map(lesion_type_dict.get)
    skin_meta['target'] = pd.Categorical(skin_meta['Cell type']).codes
    
    # Replace missing age with the mean
    skin_meta['age'].fillna(skin_meta['age'].mean(), inplace=True)
    
    # Create a binary label for benign (0) and malignant (1) lesions.
    skin_meta = skin_meta.assign(label=skin_meta.apply(lesion_type, axis=1))
    
    skin_meta['image'] = skin_meta['image_id'].apply(lambda img_id: convert_image_into_array(img_id, source_bucket=source_bucket, images_dir=images_dir))
    
    # Convert the processed metadata DataFrame to CSV (in memory)
    pickle_bytes = BytesIO()
    skin_meta.to_pickle(pickle_bytes)
    pickle_bytes.seek(0)
    pickle_length = pickle_bytes.getbuffer().nbytes
    
    try:
        minio_client.put_object(
            processed_bucket, 
            processed_metadata_filename, 
            pickle_bytes, 
            pickle_length, 
            content_type="application/octet-stream"
        )
        print(f"Uploaded processed metadata pickle file to '{processed_metadata_filename}' in bucket '{processed_bucket}'.")
    except S3Error as err:
        print(f"Error uploading processed metadata pickle file: {err}")
        raise
        
    
def create_segmented_images(
    processed_bucket: str, 
    processed_metadata_filename: str,
    n_samples: int = 2000, 
):
    skin_meta = get_metadata_file(processed_bucket, metadata_filename)
    print(f"Skin Image shape: {skin_meta['image'][0].shape}")
    
    balanced_data = create_balanced_dataset(skin_meta, n_samples)
    
    create_segmented_pictures(balanced_data['image'], balanced_data, processed_bucket)

def convert_image_into_array(
    image_id: str,
    source_bucket: str,
    images_dir: str, 
) -> np.ndarray:
    """
    Retrieves an image from a MinIO bucket and converts it into a NumPy array.
    
    Args:
        image_id (str): The unique identifier of the image (without extension).
        images_dir (str): The directory (prefix) where images are stored in the bucket.
        
    Returns:
        np.ndarray: The image as a NumPy array, or None if retrieval fails.
    """
    object_name = f"{images_dir}/{image_id}.jpg"  # adjust if your path is different
    try:
        response = minio_client.get_object(source_bucket, object_name)
        image = Image.open(BytesIO(response.read()))
        return np.asarray(image)
    except Exception as e:
        print(f"Error retrieving image '{object_name}': {e}")
        return None
    
def get_metadata_file(
    bucket: str, 
    metadata_filename: str,
) -> pd.DataFrame:
    """
    Retrieve the metadata CSV file from the source bucket.
    """
    try:
        # Get the metadata file as an object from the source bucket.
        response = minio_client.get_object(bucket, metadata_filename)
        # Read all bytes from the response.
        data = response.read()
        # Use StringIO to convert bytes to a file-like object for pandas.
        df = pd.read_csv(io.StringIO(data.decode('utf-8')))
        print(f"Retrieved metadata file '{metadata_filename}' from bucket '{bucket}'.")
        return df
    except S3Error as err:
        print(f"Error retrieving metadata file from bucket: {err}")
        raise
            
def plot_1(skin_meta: pd.DataFrame):
    sns.set_style('darkgrid', {"grid.color": ".8", "grid.linestyle": ":"})
    fig, ax = plt.subplots(figsize=(12, 8))
    ax = sns.countplot(x='Cell type', data=skin_meta, palette='coolwarm')
    for container in ax.containers:
        ax.bar_label(container)
    plt.title('Counts per Cell Type of Skin Lesions')
    plt.ylabel('Counts')
    plt.xlabel('Skin lesion type')
    plt.xticks(rotation=35)
    plt.show()

def plot_2(skin_meta: pd.DataFrame):
    sns.set_style('darkgrid', {"grid.color": ".8", "grid.linestyle": ":"})
    fig, ax = plt.subplots(figsize=(12, 8))
    ax = sns.histplot(data=skin_meta, x='age', hue='Cell type', multiple='stack', palette='coolwarm')
    plt.title('Skin lesion type occurrence by Age')
    plt.ylabel('Counts')
    plt.xlabel('Age')
    plt.show()

def lesion_type(row):
    if row['dx'] in ['bkl', 'df', 'nv', 'vasc']:
        return 0  # benign
    elif row['dx'] in ['akiec', 'mel', 'bcc']:
        return 1  # malignant

def create_balanced_dataset(
    skin_meta: pd.DataFrame, 
    n_samples: int, 
    seed: int =42
) -> pd.DataFrame:
    skin_0 = skin_meta[skin_meta['label'] == 0]
    skin_1 = skin_meta[skin_meta['label'] == 1]
    
    skin_0_balanced = resample(skin_0, replace=True, n_samples=n_samples, random_state=seed)
    skin_1_balanced = resample(skin_1, replace=True, n_samples=n_samples, random_state=seed)
    
    skin_data = pd.concat([skin_0_balanced, skin_1_balanced]).reset_index()

    return skin_data

def create_segmented_pictures(
    images_array,
    skin_data: pd.DataFrame,  
    processed_bucket: str,
) -> pd.DataFrame:
    kernel = np.ones((8, 8), np.uint8)
    skin_data['segmented_image'] = ''
    
    for i in range(len(images_array)):
        image_np = images_array[i]
        if image_np is None:
            print(f"Skipping index {i}: image not found.")
            continue

        image_gr = cv2.cvtColor(image_np, cv2.COLOR_RGB2GRAY)
        image_blur = cv2.medianBlur(image_gr, 5)
        _, image_result = cv2.threshold(image_blur, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
        image_cropped = image_result[60:400, 50:550]
        image_cropped = cv2.bitwise_not(image_cropped)
        
        image_segmented = Image.fromarray(image_cropped)
        buffer = BytesIO()
        image_segmented.save(buffer, format="JPEG")
        buffer.seek(0)

        image_id = skin_data['image_id'].iloc[i]
        object_name = f"segmented/{image_id}.jpg"
        try:
            file_size = buffer.getbuffer().nbytes
            minio_client.put_object(processed_bucket, object_name, buffer, file_size, content_type="image/jpeg")
            print(f"Uploaded segmented image: {object_name}")
        except S3Error as err:
            print(f"Error uploading segmented image '{object_name}': {err}")
            raise

        skin_data.at[i, 'segmented_image'] = image_cropped
        
    return skin_data
