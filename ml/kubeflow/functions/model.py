# Import necessities
import argparse
import numpy as np
import pandas as pd
import PIL.Image as Image
import matplotlib.pyplot as plt
import seaborn as sns
import math

from tensorflow import keras

from sklearn.model_selection import train_test_split
from sklearn.decomposition import PCA
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, confusion_matrix, classification_report
from collections import defaultdict

from io import BytesIO
from google.cloud import storage

import functions.glob as flyte_glob

# for gpu in tf.config.experimental.list_physical_devices("GPU"):
#     tf.config.experimental.set_memory_growth(gpu, True)

def train_random_forest(
    bucket: str,
    metadata_filename: str,
    data_path: str,
    seed=42,
):
    """
    Train Random Forest Model
    
    Args:
        bucket (str): The name of the GCP bucket.
        metadata_filename (str): The name we will set for the metadata CSV file.
        data_path (str): The folder path in the bucket from where all data should be retrieved.
        seed (int): The answer to the great question of “life, the universe and everything” is 42.
        
    Returns:
        None
    """  
    data = flyte_glob.get_metadata_file( # type: ignore
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=data_path, 
        target="pickle"
    )
    skin_features = data[['perimeter', 'non_zeros', 'circularity',
       'main_assymetry', 'secondary_assymetry', 'r_channel', 'g_channel',
       'b_channel', 'label']]
    
    skin_features = skin_features.sample(frac=1, random_state=seed)

    skin_features.head()

    skin_features['perimeter'] = skin_features['perimeter'].astype(float)
    skin_features['non_zeros'] = skin_features['perimeter'].astype(float)
    skin_features['circularity'] = skin_features['perimeter'].astype(float)
    skin_features['main_assymetry'] = skin_features['perimeter'].astype(float)
    skin_features['secondary_assymetry'] = skin_features['perimeter'].astype(float)
    
    # create a PCA object with n_components=1
    pca = PCA(n_components=1)

    for channel in ['r_channel', 'g_channel', 'b_channel']:
        # Apply PCA to each row in the channel column and save the result in the same place
        skin_features[channel] = skin_features[channel].apply(lambda x: pca.fit_transform(x).flatten()[0])

    skin_features.head()
    # Split into data and labels
    labels = skin_features['label']
    data = skin_features.drop('label', axis=1)
    # Split the dataset into training and testing sets
    X_train, X_test, y_train, y_test = train_test_split(data, labels, test_size=0.25, random_state=seed)
    
    RF = RandomForest(RandomForestClassifier(random_state=seed), X_train, y_train, X_test, y_test, 'RF')
    return


# Random Forest classifier
def RandomForest(
        model, 
        X_train, 
        y_train, 
        X_test, 
        y_test, 
        title,
    ):
    
    predictions = defaultdict(list)

    model = model
    hist = model.fit(X_train, y_train)
    y_pred=model.predict(X_test)
    predictions[title].append(y_pred)
    print('Model accuracy score with 10 decision-trees: {0:0.4f}%'. format(accuracy_score(y_test, y_pred) * 100))
    
    target_names = ['Melanoma', 'Non-melanoma']
    print(classification_report(y_test, y_pred, target_names=target_names, zero_division=0))
    
    cm = confusion_matrix(y_test, y_pred)
    cm_matrix = pd.DataFrame(data=cm, columns=['Actual Normal', 'Actual Pathologic'], 
                                 index=['Predicted Normal', 'Predicted Pathologic'])
    sns.heatmap(cm_matrix, annot=True, fmt='d', cmap='coolwarm')
    print('Confusion matrix:\n')
    return model


def train_cnn(
    bucket: str, 
    metadata_filename: str,
    images_dir: str,
    segmented_images_dir: str,
    data_path: str,
    processed_data_path: str,
    sample: int,
    batch_size: int,
    epochs: int,
    get_segmented: bool,
):
    # Get the datasets:
    training_set, validation_set = _get_datasets(
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=data_path, 
        processed_data_path=processed_data_path,
        images_dir=images_dir,
        segmented_images_dir=segmented_images_dir,
        test_size=0.2, 
        data_size=sample, 
        batch_size=batch_size,
        get_segmented=get_segmented,
    )
    
    # Compute steps_per_epoch based on the number of training samples.
    # With a sample size of `sample` and a test split of 20%, training samples = sample * 0.8.
    steps_per_epoch = math.ceil((sample * 0.8) / batch_size)
    
    # Get the model
    model = _get_model()
    # Initialize the optimizer
    optimizer = keras.optimizers.Adam(learning_rate=1e-4)
    optimizer.lr = optimizer.learning_rate
    
    model.compile(optimizer=optimizer, loss='binary_crossentropy', metrics=['accuracy'])
    
    history = model.fit(
        training_set,
        validation_data=validation_set,
        epochs=epochs,
        steps_per_epoch=steps_per_epoch,
    )
    return
    
def _get_model() -> keras.Model: 
    # Load a pre-trained VGG16 model and remove the top layers
    base_model = keras.applications.VGG16(
        weights='imagenet', 
        include_top=False, 
        input_shape=(224, 224, 3),
    )

    head_model = base_model.output
    head_model = keras.layers.Flatten(name="flatten")(head_model)
    head_model = keras.layers.Dense(256, activation='relu', kernel_regularizer=keras.regularizers.l2(0.01))(head_model)
    head_model = keras.layers.Dropout(0.5)(head_model)
    head_model = keras.layers.Dense(128, activation='relu', kernel_regularizer=keras.regularizers.l2(0.01))(head_model)
    head_model = keras.layers.Dense(1, activation='sigmoid')(head_model)

    # Create the model
    model = keras.Model(name="skin_cancer_detection", inputs=base_model.input, outputs=head_model)

    for layer in base_model.layers:
        layer.trainable = False

    return model 
    
def _get_datasets(
    bucket: str,
    metadata_filename: str,
    data_path: str,
    processed_data_path: str,
    images_dir: str,
    segmented_images_dir: str, 
    test_size: float, 
    data_size: int,
    batch_size: str,
    get_segmented: bool,
    seed: int = 42,
    is_evaluation: bool = False,
):
    # Build the dataset going through the classes directories and collecting the images
    metadata = flyte_glob.get_metadata_file(
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=processed_data_path,
        target="pickle"
    )

    data = metadata[:data_size]

    labels = data['label']

    path = processed_data_path if get_segmented else data_path
    dir = segmented_images_dir if get_segmented else images_dir
    images = load_images_from_gcs(
        dataset=data, 
        bucket=bucket, 
        data_path=path, 
        images_dir=dir,
    )
    # Check if its an evaluation, if so, use the entire data
    if is_evaluation:
        return images, labels

    # Split the dataset into training and validation sets
    x_train, x_test, y_train, y_test = train_test_split(
        images, labels, test_size=test_size, stratify=labels, random_state=seed,
    )

    # Construct the training image generator for data augmentation:
    image_data_generator = keras.preprocessing.image.ImageDataGenerator(
        rotation_range=20,
        zoom_range=0.15,
        width_shift_range=0.2,
        height_shift_range=0.2,
        shear_range=0.15,
        horizontal_flip=True,
        fill_mode="nearest",
    )

    return (
        image_data_generator.flow(x_train, y_train, batch_size=batch_size),
        (x_test, y_test),
    )

def load_images_from_gcs(
    dataset: pd.DataFrame, 
    bucket: str,
    data_path: str,
    images_dir: str,
    target_size=(224, 224),
) -> np.ndarray:
    """
    Retrieve images from a GCS bucket (replacing MinIO) and return them as a NumPy array.
    
    Args:
        dataset (pd.DataFrame): DataFrame containing an 'image_id' column.
        bucket (str): The name of the GCS bucket.
        data_path (str): The base path in the bucket where images are stored.
        target_size (tuple): The target size to which each image will be resized.
        
    Returns:
        np.ndarray: A stacked array of all images.
    """
    images = []
    storage_client = storage.Client()
    bucket_obj = storage_client.bucket(bucket)
    
    for i in range(len(dataset)):
        image_id = dataset.iloc[i]['image_id']
        # Construct the object path. Here we assume images are stored under "segmented" inside data_path.
        object_path = f"{data_path}/{images_dir}/{image_id}.jpg"
        try:
            blob = bucket_obj.blob(object_path)
            img_bytes = blob.download_as_bytes()
            img = Image.open(BytesIO(img_bytes)).convert("RGB")
            img = img.resize(target_size)
            img_array = np.array(img)
            img_array = np.expand_dims(img_array, axis=0)
            images.append(img_array)
        except Exception as err:
            print(f"Error loading image '{object_path}': {err}")
            continue

    return np.vstack(images)


# Function to plot loss during training
def plot_loss(model_fitting):
    """Plots the training and validation loss values of a deep learning model over epochs"""
    plt.plot(model_fitting.history['loss'], label='train') 
    if 'val_loss' in model_fitting.history:
        plt.plot(model_fitting.history['val_loss'], label='test')
    plt.title('Model Loss')
    plt.xlabel('epochs')
    plt.ylabel('loss values')
    plt.legend(loc='upper right')
    plt.show()
    
# Function to plot accuracy measurements during training-testing
def plot_accuracy(model_fitting):
    """Plots the training and validation accuracy values of a deep learning model over epochs"""
    plt.plot(model_fitting.history['accuracy'], label='train')
    if 'val_accuracy' in model_fitting.history:
        plt.plot(model_fitting.history['val_accuracy'], label='test')
    plt.title('Model Accuracy')
    plt.xlabel('epochs')
    plt.ylabel('accuracy')
    plt.legend(loc='upper right')
    plt.show()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Train a model")
    parser.add_argument("--model", choices=["rf", "cnn"], default="rf", help="Model type to train")
    parser.add_argument("--bucket", type=str, required=True, help="GCS bucket name")
    parser.add_argument("--metadata_filename", type=str, required=True, help="Metadata filename")
    parser.add_argument("--data_path", type=str, required=True, help="Data path")
    # Arguments specific to CNN
    parser.add_argument("--processed_data_path", type=str, default="", help="Processed data path")
    parser.add_argument("--images_dir", type=str, default="", help="Images directory")
    parser.add_argument("--segmented_images_dir", type=str, default="", help="Segmented images directory")
    parser.add_argument("--sample", type=int, default=400, help="Sample size")
    parser.add_argument("--batch_size", type=int, default=32, help="Batch size")
    parser.add_argument("--epochs", type=int, default=10, help="Number of epochs")
    parser.add_argument("--get_segmented", type=lambda x: x.lower() == "true", default=False, help="Whether to train on the segmented images")
    # Argument specific to Random Forest
    parser.add_argument("--seed", type=int, default=42, help="Seed for the random forest model")
    
    args = parser.parse_args()
    
    if args.model == "cnn":
        train_cnn(
            bucket=args.bucket,
            metadata_filename=args.metadata_filename,
            images_dir=args.images_dir,
            segmented_images_dir=args.segmented_images_dir,
            data_path=args.data_path,
            processed_data_path=args.processed_data_path,
            sample=args.sample,
            batch_size=args.batch_size,
            epochs=args.epochs,
            get_segmented=args.get_segmented,
        )
    else:
        train_random_forest(
            bucket=args.bucket,
            metadata_filename=args.metadata_filename,
            data_path=args.data_path,
            seed=args.seed,
        )