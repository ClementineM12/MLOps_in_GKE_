from flytekit import task, workflow, dynamic, PodTemplate, Resources
from kubernetes import client

import functions.load as load
import functions.process as process
import functions.feature_engineering as feature_engineering
import functions.model as model

from typing import Literal

retrieved_metadata_filename = "HAM10000_metadata.csv" 
bucket_name = "flyte-data-01" 
data_path = "data"
processed_data_path = "processed_data"
metadata_filename = "metadata.pkl"
images_dir = "images"
segmented_images_dir="segmented_images"

region="europe-west4"
project_id="mlops-development-project"
image_name="mlop-base"
image_tag="latest"
repository="flyte"

artifact_registry=f"{region}-docker.pkg.dev/{project_id}/{repository}"
base_image_ref = f"{artifact_registry}/{image_name}:{image_tag}"

def getPodTemplate(label: Literal["highmem", "highcpu"]):
    return PodTemplate(
    pod_spec=client.V1PodSpec(
        containers=[],
        affinity=client.V1Affinity(
            node_affinity=client.V1NodeAffinity(
                required_during_scheduling_ignored_during_execution=client.V1NodeSelector(
                    node_selector_terms=[
                        client.V1NodeSelectorTerm(
                            match_expressions=[
                                client.V1NodeSelectorRequirement(
                                    key="dedicated",
                                    operator="In",
                                    values=[label],
                                )
                            ]
                        )
                    ]
                )
            )
        ),
    )
)

def getPodSpec(label: Literal["highmem", "highcpu"], target: Literal["limits", "requests"]):
    mapResources = {
        "highmem": {
            "requests": Resources(cpu="2", mem="8Gi"),
            "limits": Resources(cpu="4", mem="32Gi")
        },
        "highcpu": {
            "requests": Resources(cpu="4", mem="8Gi"),
            "limits": Resources(cpu="16", mem="32Gi")
        }
    }
    return mapResources[label][target]


# TASK 1: Fetch dataset
@task(
    requests=getPodSpec("highmem", "requests"),
    limits=getPodSpec("highmem", "limits"),
    pod_template=getPodTemplate("highmem"),
    container_image=base_image_ref, 
    name="fetch_dataset",
)
def fetch_dataset_task() -> None:
    load.fetch_dataset(
        retrieved_metadata_filename=retrieved_metadata_filename, 
        metadata_filename=metadata_filename,
        bucket=bucket_name, 
        images_dir=images_dir, 
        data_path=data_path
    )
    return

# TASK 2: Process metadata and images
@task(
    requests=getPodSpec("highmem", "requests"),
    limits=getPodSpec("highmem", "limits"),
    pod_template=getPodTemplate("highmem"),
    container_image=base_image_ref, 
    name="process_metadata_and_images",
)
def process_task(sample: int) -> None:
    process.process_metadata(
        metadata_filename=metadata_filename,
        bucket=bucket_name,
        images_dir=images_dir,
        data_path=data_path,
        processed_data_path=processed_data_path
    )
    process.create_segmented_images(
        metadata_filename=metadata_filename,
        bucket=bucket_name,
        images_dir=segmented_images_dir,
        processed_data_path=processed_data_path,
        sample=sample
    )
    return

# TASK 3: Feature Engineering
@task(
    requests=getPodSpec("highmem", "requests"),
    limits=getPodSpec("highmem", "limits"),
    pod_template=getPodTemplate("highmem"),
    container_image=base_image_ref, 
    name="feature_engineering",
)
def feature_engineering_task() -> None:
    feature_engineering.feature_engineer(
        bucket=bucket_name,
        metadata_filename=metadata_filename,
        data_path=processed_data_path
    )
    return

# TASK 4: Train Random Forest
@task(
    requests=getPodSpec("highcpu", "requests"),
    limits=getPodSpec("highcpu", "limits"),
    pod_template=getPodTemplate("highcpu"),
    container_image=base_image_ref, 
    name="train_random_forest",
)
def train_random_forest_task() -> None:
    model.train_random_forest(
        bucket=bucket_name,
        metadata_filename=metadata_filename,
        data_path=processed_data_path,
    )

# TASK 4: Train CNN
@task(
    requests=getPodSpec("highcpu", "requests"),
    limits=getPodSpec("highcpu", "limits"),
    pod_template=getPodTemplate("highcpu"),
    container_image=base_image_ref, 
    name="train_cnn",
)
def train_cnn_task(sample: int, batch_size: int = 32, epochs: int = 10) -> None:
    model.train_cnn(
        bucket=bucket_name,
        metadata_filename=metadata_filename,
        data_path=processed_data_path,
        sample=sample,
        batch_size=batch_size,
        epochs=epochs,
    )

@dynamic
def branch_training_task(model_type: str, sample: int) -> None:
    # Now, model_type is resolved at runtime and you can use standard if/else.
    if model_type == "random_forest":
        train_random_forest_task()
    elif model_type == "cnn":
        train_cnn_task(sample=sample)
    else:
        raise ValueError(f"Unsupported model_type: {model_type}")

# WORKFLOW: Compose the tasks into a single workflow with branching
@workflow
def mlops_workflow(model_type: str = "cnn", sample: int = 1000) -> None:
    fetch_dataset = fetch_dataset_task()
    processing = process_task(sample=sample)
    feature_engineering = feature_engineering_task()
    
    # Use the model_type parameter to decide which training task to run.
    branch_training = branch_training_task(model_type=model_type, sample=sample)

    fetch_dataset >> processing
    processing >> feature_engineering
    feature_engineering >> branch_training
