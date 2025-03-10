# Import necessities
import numpy as np
import pandas as pd
import PIL.Image as Image
import matplotlib.pyplot as plt
import seaborn as sns
import math

from minio import Minio
from minio.error import S3Error

import io
from io import BytesIO

import mlrun 
from mlrun.frameworks.tf_keras import apply_mlrun
from mlrun.artifacts import PlotArtifact
from mlrun.execution import MLClientCtx

from tensorflow import keras

from sklearn.model_selection import train_test_split
from sklearn.decomposition import PCA
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, confusion_matrix, classification_report
from collections import defaultdict

# for gpu in tf.config.experimental.list_physical_devices("GPU"):
#     tf.config.experimental.set_memory_growth(gpu, True)

minio_client = Minio(
    "minio.mlrun.svc.cluster.local:9000",
    access_key="mlrun",
    secret_key="mlrun1234",
    secure=False
)

def train_random_forest(
    processed_bucket: str,
    processed_metadata_filename: str,
    seed=42,
):
    
    data = get_metadata_file(processed_bucket, processed_metadata_filename, "pickle")
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
    
    # mlrun.apply_mlrun(model=RF, model_name="random-forest", X_test=x_test, y_test=Y_test)


# Random Forest classifier
def RandomForest(
    model, 
    X_train, 
    y_train, 
    X_test, 
    y_test, 
    title
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
    context: MLClientCtx,
    processed_bucket: str, 
    processed_metadata_filename: str,
    sample: int,
    batch_size: int,
    epochs: int,
):
    # Get the datasets:
    training_set, validation_set = _get_datasets(
        processed_bucket, processed_metadata_filename, 0.2, sample, batch_size
    )
    
    # Compute steps_per_epoch based on the number of training samples.
    # With a sample size of `sample` and a test split of 20%, training samples = sample * 0.8.
    steps_per_epoch = math.ceil((sample * 0.8) / batch_size)
    
    # Get the model
    model = _get_model()
    # Initialize the optimizer
    optimizer = keras.optimizers.Adam(learning_rate=1e-4)
    optimizer.lr = optimizer.learning_rate
    
    # Apply MLRun's interface for tf.keras
    # apply_mlrun(model=model, model_name="skin_cancer_detection", context=context)
    
    model.compile(optimizer=optimizer, loss='binary_crossentropy', metrics=['accuracy'])
    
    history = model.fit(
        training_set,
        validation_data=validation_set,
        epochs=epochs,
        steps_per_epoch=steps_per_epoch,
    )
    
    # context.log_artifact(
    #     PlotArtifact("Plot loss", body=plot_loss(history)),
    # )
    
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
    processed_bucket: str,
    processed_metadata_filename: str,
    test_size: float, 
    data_size: int,
    batch_size: str,
    seed: int = 42,
    is_evaluation: bool = False,
):
    # Build the dataset going through the classes directories and collecting the images
    metadata = get_metadata_file(processed_bucket, processed_metadata_filename, "pickle")

    data = metadata[:data_size]

    labels = data['label']
    images = load_images_from_minio(data)
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

def evaluate_cnn(
    context: mlrun.MLClientCtx, 
    model_path: str, 
    processed_bucket: str, 
    processed_metadata_filename: str,
    batch_size: int,
):
    # Get the dataset
    x, y = _get_datasets(
        processed_bucket, processed_metadata_filename, 0.2, is_evaluation=True
    )

    # Apply MLRun's interface for tf.keras and load the model
    model_handler = mlrun_tf_keras.apply_mlrun(
        model_path=model_path,
        context=context,
    )

    # Evaluate
    model_handler.model.evaluate(x=x, y=y, batch_size=batch_size)


def get_metadata_file(
    bucket: str, 
    metadata_filename: str,
    target: str = "csv",
) -> pd.DataFrame:
    """
    Retrieve the metadata CSV file from the source bucket.
    """
    try:
        # Get the metadata file as an object from the source bucket.
        response = minio_client.get_object(bucket, metadata_filename)
        # Read all bytes from the response.
        data = response.read()

        # Process file based on the target type.
        if target == "csv":
            # Convert bytes to a file-like object and decode.
            df = pd.read_csv(io.StringIO(data.decode('utf-8')))
        elif target == "pickle":
            # Use BytesIO to load the pickled DataFrame.
            df = pd.read_pickle(io.BytesIO(data))
        else:
            raise ValueError(f"Unsupported target type: {target}")

        print(f"Retrieved metadata file '{metadata_filename}' from bucket '{bucket}'.")
        return df
    except S3Error as err:
        print(f"Error retrieving metadata file from bucket: {err}")
        raise

def put_metadata_file(
    metadata: pd.DataFrame,
    bucket: str, 
    metadata_filename: str,
    target: str = "pickle",
):
    pickle_bytes = BytesIO()
    metadata.to_pickle(pickle_bytes)
    pickle_bytes.seek(0)
    pickle_length = pickle_bytes.getbuffer().nbytes
    
    try:
        minio_client.put_object(
            bucket, 
            metadata_filename, 
            pickle_bytes, 
            pickle_length, 
            content_type="application/octet-stream"
        )
        print(f"Uploaded processed metadata pickle file to '{metadata_filename}' in bucket '{bucket}'.")
    except S3Error as err:
        print(f"Error uploading processed metadata pickle file: {err}")
        raise

def load_images_from_minio(
    dataset, 
    bucket_name: str = "processed-data", 
    target_size=(224, 224),
):
    images = []
    for i in range(len(dataset)):
        # Retrieve the image_id from the dataset.
        image_id = dataset.iloc[i]['image_id']
        # Construct the correct object path.
        object_path = f"segmented/{image_id}.jpg"
        
        # Retrieve the object from Minio using the constructed path.
        response = minio_client.get_object(bucket_name, object_path)
        try:
            # Read the image bytes and load into PIL.
            img_bytes = response.read()
            img = Image.open(BytesIO(img_bytes)).convert("RGB")
            img = img.resize(target_size)
            # Convert to numpy array and add to list.
            img_array = np.array(img)
            img_array = np.expand_dims(img_array, axis=0)   
            images.append(img_array)
        finally:
            response.close()
            response.release_conn()
    
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