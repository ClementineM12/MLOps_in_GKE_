import mlrun
import kfp
from kfp import dsl
from typing import Literal
import os

retrieved_metadata_filename = "HAM10000_metadata.csv" 
bucket_name = "mlop-train-data-01" 
data_path = "data"
processed_data_path = "processed_data"
metadata_filename = "metadata.pkl"
images_dir = "images"
segmented_images_dir="segmented_images"

region="europe-west4"
project_id="mlops-development-project"
image_name="mlop-base"
image_tag="latest"
repository="mlrun"

artifact_registry=f"{region}-docker.pkg.dev/{project_id}/{repository}"
base_image_ref = f"{artifact_registry}/{image_name}:{image_tag}"

mem_label = {"dedicated": "highmem"}
cpu_label = {"dedicated": "highcpu"}

is_first_run = False

# Set our project's name:
project_name = "mlop"
project_dir = os.path.abspath('./')

# Workflow
@kfp.dsl.pipeline(name="Skin Cancer Detection Pipeline")
def kfpipeline(
    model_type: Literal["rf", "cnn"],
    sample: int =400,
    batch_size: int = 32,
    epochs: int = 10,
    get_segmented: bool = False,
):     

    # Step 1: Load Data into GCS
    load_op = dsl.ContainerOp(
        name="Load Data into GCS",
        command=["python", "functions/load.py"],
        arguments=[
            "--retrieved_metadata_filename", retrieved_metadata_filename,
            "--metadata_filename", metadata_filename,
            "--bucket", bucket_name,
            "--images_dir", images_dir,
            "--data_path", data_path
        ]
    )
    load_op.add_node_selector_constraint("node-type", mem_label)

    # Step 2: Preprocess Data (process metadata)
    preprocess_op = dsl.ContainerOp(
        name="Preprocess Data",
        command=["python", "functions/process.py"],
        arguments=[
            "--action", "process_metadata",
            "--metadata_filename", metadata_filename,
            "--bucket", bucket_name,
            "--images_dir", images_dir,
            "--data_path", data_path,
            "--processed_data_path", processed_data_path
        ]
    )
    preprocess_op.add_node_selector_constraint("node-type", mem_label)
    preprocess_op.after(load_op)

    # Step 3: Create Segmented Images
    segmented_op = dsl.ContainerOp(
        name="Preprocess Images",
        command=["python", "functions/process.py"],
        arguments=[
            "--action", "create_segmented_images",
            "--metadata_filename", metadata_filename,
            "--bucket", bucket_name,
            "--images_dir", segmented_images_dir,
            "--processed_data_path", processed_data_path,
            "--sample", str(sample)
        ]
    )
    segmented_op.add_node_selector_constraint("node-type", mem_label)
    segmented_op.after(preprocess_op)

    # Step 4: Feature Engineering
    feature_op = dsl.ContainerOp(
        name="Feature Engineering",
        command=["python", "functions/feature_engineering.py"],
        arguments=[
            "--bucket", bucket_name,
            "--metadata_filename", metadata_filename,
            "--data_path", processed_data_path
        ]
    )
    feature_op.add_node_selector_constraint("node-type", mem_label)
    feature_op.after(segmented_op)

    # Conditional step for Random Forest training
    with dsl.Condition(model_type == "rf"):
        rf_op = dsl.ContainerOp(
            name="Train Random Forest Model",
            command=["python", "functions/model.py"],
            arguments=[
                "--model", "rf",
                "--bucket", bucket_name,
                "--metadata_filename", metadata_filename,
                "--data_path", processed_data_path,
                "--seed", "42"  # You can parameterize this value as needed.
            ]
        )
        rf_op.add_node_selector_constraint("node-type", cpu_label)
        rf_op.after(feature_op)

    # Conditional step for CNN training
    with dsl.Condition(model_type == "cnn"):
        cnn_op = dsl.ContainerOp(
            name="Train CNN",
            command=["python", "functions/model.py"],
            arguments=[
                "--model", "cnn",
                "--bucket", bucket_name,
                "--metadata_filename", metadata_filename,
                "--data_path", data_path,
                "--processed_data_path", processed_data_path,
                "--images_dir", images_dir,
                "--segmented_images_dir", segmented_images_dir,
                "--sample", str(sample),
                "--batch_size", str(batch_size),
                "--epochs", str(epochs),
                "--get_segmented", str(get_segmented)
            ]
        )
        cnn_op.add_node_selector_constraint("node-type", cpu_label)
        cnn_op.after(feature_op)