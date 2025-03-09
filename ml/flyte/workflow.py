from flytekit import task, workflow, parameter

import functions.load as load
import functions.process as process
import functions.feature_engineering as feature_engineering
import functions.model as model

retrieved_metadata_filename = "HAM10000_metadata.csv" 
bucket_name = "flyte-data" 
data_path = "data"
processed_data_path = "processed_data"
metadata_filename = "metadata.pkl"
images_dir = "images"
segmented_images_dir="segmented_images"

# TASK 1: Fetch dataset
@task
def fetch_dataset_task() -> None:
    load.fetch_dataset(
        metadata_filename=retrieved_metadata_filename, 
        metadata_filename=metadata_filename, 
        bucket=bucket_name, 
        images_dir=images_dir, 
        data_path=data_path
    )

# TASK 2: Process metadata and images
@task
def process_task(n_samples: str) -> None:
    process.process_metadata(
        metadata_filename=metadata_filename,
        bucket=bucket_name,
        images_dir=images_dir,
        data_path=data_path,
        processed_data_path=processed_data_path
    )
    process.create_segmented_images(
        bucket=bucket_name,
        metadata_filename=metadata_filename,
        images_dir=segmented_images_dir,
        processed_data_path=processed_data_path,
        n_samples=n_samples
    )

# TASK 3: Feature Engineering
@task
def feature_engineering_task() -> None:
    feature_engineering.feature_engineer(
        bucket=bucket_name,
        metadata_filename=metadata_filename,
        data_path=processed_data_path
    )

# TASK 4: Train Random Forest
@task
def train_random_forest_task() -> None:
    model.train_random_forest(
        bucket=bucket_name,
        metadata_filename=metadata_filename,
        data_path=processed_data_path,
    )

# TASK 5: Train CNN
@task
def train_cnn_task(sample: int, batch_size: int = 32, epochs: int = 10) -> None:
    model.train_cnn(
        sample=sample,
        batch_size=batch_size,
        epochs=epochs,
    )

# WORKFLOW: Compose the tasks into a single workflow with branching
@workflow
def skin_cancer_workflow(model_type: str = parameter.String("random_forest")) -> None:
    fetch_dataset_task()
    process_task(n_samples=2000)
    feature_engineering_task()
    
    # Use the model_type parameter to decide which training task to run.
    if model_type == "random_forest":
        train_random_forest_task()
    elif model_type == "cnn":
        train_cnn_task(sample=1000)
    else:
        raise ValueError(f"Unsupported model_type: {model_type}")
