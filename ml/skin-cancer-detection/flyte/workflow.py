import os
import pandas as pd
from flytekit import task, workflow
from flytekit.types.file import FlyteFile

import functions as run

# Assume your other imports (cv2, numpy, glob, etc.) and function definitions are available.

# TASK 1: Fetch dataset
@task
def fetch_dataset_task() -> FlyteFile:
    """
    Check if dataset exists in MinIO. If not, download from Kaggle,
    upload to MinIO, and then (optionally) transfer to GCS.
    Returns the local path to the dataset.
    """
    dataset_path = run.fetch_dataset()
    return FlyteFile(dataset_path)

# TASK 2: Process metadata and images
@task
def process_task(dataset_file: FlyteFile) -> FlyteFile:
    """
    Load the metadata CSV from the dataset folder, then run the process() function.
    Returns the path to the processed metadata CSV.
    """
    dataset_path = dataset_file.path
    metadata_csv_path = os.path.join(dataset_path, "HAM10000_metadata.csv")
    
    # Load metadata into a DataFrame
    skin_meta = pd.read_csv(metadata_csv_path)
    # Process the metadata (this function will merge image paths, perform plotting, etc.)
    run.process(skin_meta)
    
    # Optionally, write the processed DataFrame to a new CSV file.
    processed_csv = os.path.join(dataset_path, "processed_skin_meta.csv")
    skin_meta.to_csv(processed_csv, index=False)
    return FlyteFile(processed_csv)

# TASK 3: Feature Engineering
@task
def feature_engineering_task(processed_file: FlyteFile) -> FlyteFile:
    """
    Perform feature engineering on the processed metadata.
    Returns the path to the feature engineered CSV.
    """
    skin_meta = pd.read_csv(processed_file.path)
    skin_meta = run.feature_engineer(skin_meta)
    
    engineered_csv = os.path.join(os.path.dirname(processed_file.path), "engineered_skin_meta.csv")
    skin_meta.to_csv(engineered_csv, index=False)
    return FlyteFile(engineered_csv)

# TASK 4: Train Random Forest
@task
def train_random_forest_task(engineered_file: FlyteFile) -> None:
    """
    Train a Random Forest model on the engineered dataset.
    """
    data = pd.read_csv(engineered_file.path)
    run.train_random_forest(data)

# TASK 5: Train CNN
@task
def train_cnn_task(engineered_file: FlyteFile) -> None:
    """
    Train a CNN on a sample of the engineered dataset.
    """
    data = pd.read_csv(engineered_file.path)
    run.train_cnn(data, sample=400)

# WORKFLOW: Compose the tasks into a single workflow
@workflow
def skin_cancer_workflow() -> None:
    dataset_file = fetch_dataset_task()
    processed_file = process_task(dataset_file=dataset_file)
    engineered_file = feature_engineering_task(processed_file=processed_file)
    train_random_forest_task(engineered_file=engineered_file)
    train_cnn_task(engineered_file=engineered_file)

if __name__ == "__main__":
    # Run the workflow locally
    skin_cancer_workflow()
