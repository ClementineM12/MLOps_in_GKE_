import os
import pandas as pd
from flytekit import task, workflow, parameter
from flytekit.types.file import FlyteFile

import functions as run  # Ensure that your functions folder is in the PYTHONPATH or installed as a package

# TASK 1: Fetch dataset
@task
def fetch_dataset_task() -> FlyteFile:
    dataset_path = run.fetch_dataset()
    return FlyteFile(dataset_path)

# TASK 2: Process metadata and images
@task
def process_task(dataset_file: FlyteFile) -> FlyteFile:
    dataset_path = dataset_file.path
    metadata_csv_path = os.path.join(dataset_path, "HAM10000_metadata.csv")
    
    skin_meta = pd.read_csv(metadata_csv_path)
    run.process(skin_meta)
    
    processed_csv = os.path.join(dataset_path, "processed_skin_meta.csv")
    skin_meta.to_csv(processed_csv, index=False)
    return FlyteFile(processed_csv)

# TASK 3: Feature Engineering
@task
def feature_engineering_task(processed_file: FlyteFile) -> FlyteFile:
    skin_meta = pd.read_csv(processed_file.path)
    skin_meta = run.feature_engineer(skin_meta)
    
    engineered_csv = os.path.join(os.path.dirname(processed_file.path), "engineered_skin_meta.csv")
    skin_meta.to_csv(engineered_csv, index=False)
    return FlyteFile(engineered_csv)

# TASK 4: Train Random Forest
@task
def train_random_forest_task(engineered_file: FlyteFile) -> None:
    data = pd.read_csv(engineered_file.path)
    run.train_random_forest(data)

# TASK 5: Train CNN
@task
def train_cnn_task(engineered_file: FlyteFile, sample: int = 400) -> None:
    data = pd.read_csv(engineered_file.path)
    run.train_cnn(data, sample=sample)

# WORKFLOW: Compose the tasks into a single workflow with branching
@workflow
def skin_cancer_workflow(model_type: str = parameter.String("random_forest")) -> None:
    dataset_file = fetch_dataset_task()
    processed_file = process_task(dataset_file=dataset_file)
    engineered_file = feature_engineering_task(processed_file=processed_file)
    
    # Use the model_type parameter to decide which training task to run.
    if model_type == "random_forest":
        train_random_forest_task(engineered_file=engineered_file)
    elif model_type == "cnn":
        train_cnn_task(engineered_file=engineered_file, sample=400)
    else:
        raise ValueError(f"Unsupported model_type: {model_type}")
