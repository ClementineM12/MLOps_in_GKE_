import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.utils import resample
import cv2
import numpy as np
from PIL import Image
from io import BytesIO

from google.cloud import storage
import functions.glob as flyte_glob

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
    

def process_metadata(
    metadata_filename: str, 
    bucket: str,
    images_dir: str,
    data_path: str,
    processed_data_path: str
) -> None:
    # Retrieve the metadata DataFrame from the bucket.
    skin_meta = flyte_glob.get_metadata_file(bucket=bucket, metadata_filename=metadata_filename, data_path=data_path)
    
    # (Optional) Retrieve raw images from source bucket.
    # Assume that raw images are stored under a prefix "images/" and that the metadata CSV
    # has an 'image_id' column which corresponds to the filename (without extension).
    def get_image_from_gcs(image_id: str):
        storage_client = storage.Client()

        object_name = f"{data_path}/{images_dir}/{image_id}.jpg"
        try:
            src_bucket = storage_client.bucket(bucket)
            blob = src_bucket.blob(object_name)
            data = blob.download_as_bytes()
            image = Image.open(BytesIO(data))
            return image
        except Exception as e:
            print(f"Error retrieving image '{object_name}': {e}")
            return None

    # Add a column with the downloaded PIL images
    skin_meta['image'] = skin_meta['image_id'].apply(get_image_from_gcs)
    
    # Create additional columns for better understanding of features
    skin_meta['Cell type'] = skin_meta['dx'].map(lesion_type_dict.get)
    skin_meta['target'] = pd.Categorical(skin_meta['Cell type']).codes
    
    # Replace missing age with the mean
    skin_meta['age'].fillna(skin_meta['age'].mean(), inplace=True)
    
    # Create a binary label for benign (0) and malignant (1) lesions.
    skin_meta = skin_meta.assign(label=skin_meta.apply(lesion_type, axis=1))
    
    skin_meta['image'] = skin_meta['image_id'].apply(
        lambda img_id: convert_image_into_array(img_id, bucket=bucket, images_dir=images_dir)
    )

    flyte_glob.put_metadata_file(
        metadata=skin_meta, 
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=processed_data_path,
    )
        
    
def create_segmented_images(
    metadata_filename: str,
    bucket: str, 
    processed_data_path: str,
    images_dir: str,
    sample: int, 
) -> None:
    metadata = flyte_glob.get_metadata_file(
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=processed_data_path, 
        target="pickle"
    )
    print(f"Skin Image shape: {metadata['image'][0].shape}")
    
    balanced_data = create_balanced_dataset(metadata=metadata, sample=sample)
    
    skin_data = create_segmented_pictures(
        images_array=balanced_data['image'], 
        metadata=balanced_data, 
        bucket=bucket,
        data_path=processed_data_path,
        images_dir=images_dir)
    flyte_glob.put_metadata_file(
        metadata=skin_data, 
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=processed_data_path
    )


def convert_image_into_array(
    image_id: str,
    bucket_name: str,
    images_dir: str, 
    data_path: str,
) -> np.ndarray:
    """
    Retrieves an image from a GCS bucket and converts it into a NumPy array.
    
    Args:
        image_id (str): The unique identifier of the image (without extension).
        bucket_name (str): The GCS bucket name.
        images_dir (str): The directory (prefix) where images are stored in the bucket.
        data_path (str): The base path in the bucket where the images are stored.
        
    Returns:
        np.ndarray: The image as a NumPy array, or None if retrieval fails.
    """
    # Build the full object name for the image.
    object_name = f"{data_path}/{images_dir}/{image_id}.jpg"
    try:
        storage_client = storage.Client()
        bucket = storage_client.bucket(bucket_name)
        blob = bucket.blob(object_name)
        data = blob.download_as_bytes()
        image = Image.open(BytesIO(data))
        return np.asarray(image)
    except Exception as e:
        print(f"Error retrieving image '{object_name}': {e}")
        return None

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
    metadata: pd.DataFrame, 
    sample: int, 
    seed: int =42
) -> pd.DataFrame:
    skin_0 = metadata[metadata['label'] == 0]
    skin_1 = metadata[metadata['label'] == 1]
    
    skin_0_balanced = resample(skin_0, replace=True, n_samples=sample, random_state=seed)
    skin_1_balanced = resample(skin_1, replace=True, n_samples=sample, random_state=seed)
    
    skin_data = pd.concat([skin_0_balanced, skin_1_balanced]).reset_index()

    return skin_data

def create_segmented_pictures(
    images_array,
    metadata: pd.DataFrame,  
    bucket: str,
    data_path: str,
    images_dir: str
) -> pd.DataFrame:
    """
    Process the images in images_array to create segmented images,
    upload them to a GCS bucket under a given data_path, and update the
    skin_data DataFrame with the processed images as NumPy arrays.
    
    Args:
        images_array: A list/array of raw images as NumPy arrays.
        skin_data (pd.DataFrame): DataFrame containing image metadata with an 'image_id' column.
        bucket (str): The name of the GCS bucket.
        data_path (str): The folder path in the bucket where segmented images will be stored.
        
    Returns:
        pd.DataFrame: Updated DataFrame with a new column 'segmented_image' containing the processed image arrays.
    """
    kernel = np.ones((8, 8), np.uint8)
    metadata['segmented_image'] = ''

    # Initialize the GCS client and get the bucket.
    storage_client = storage.Client()
    bucket_obj = storage_client.bucket(bucket)
    
    for i in range(len(images_array)):
        image_np = images_array[i]
        if image_np is None:
            print(f"Skipping index {i}: image not found.")
            continue

        # Process the image: convert to grayscale, blur, threshold, crop, and invert.
        image_gr = cv2.cvtColor(image_np, cv2.COLOR_RGB2GRAY)
        image_blur = cv2.medianBlur(image_gr, 5)
        _, image_result = cv2.threshold(image_blur, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
        image_cropped = image_result[60:400, 50:550]
        image_cropped = cv2.bitwise_not(image_cropped)
        
        # Convert the segmented image to a PIL Image.
        image_segmented = Image.fromarray(image_cropped)
        buffer = BytesIO()
        image_segmented.save(buffer, format="JPEG")
        buffer.seek(0)

        # Use the image_id from the DataFrame to construct the object name.
        image_id = metadata['image_id'].iloc[i]
        object_name = f"{data_path}/{images_dir}/{image_id}.jpg"
        
        try:
            # Upload the file to the GCS bucket.
            blob = bucket_obj.blob(object_name)
            blob.upload_from_file(buffer, content_type="image/jpeg")
            print(f"Uploaded segmented image to '{object_name}'")
        except Exception as err:
            print(f"Error uploading segmented image '{object_name}': {err}")
            raise

        # Update the DataFrame with the processed image array.
        metadata.at[i, 'segmented_image'] = image_cropped.astype(np.uint8)
        
    return metadata
