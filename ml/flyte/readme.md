## Build and Run the Workflow

This repository uses a GitHub Action to automatically build and push the Docker image used in the workflow. Follow the steps below to install dependencies and schedule your workflow on Flyte.

### 1. Build the Docker Image

When you push changes to the repository, the GitHub Action will automatically build your Docker image and push it to the configured registry. To manually trigger a build, simply commit and push your changes.

### 2. Install Python Dependencies

Before running the workflow locally or scheduling it on Flyte, install the required Python packages by running:

```bash
pip install -r requirements.txt
```

### 3. Schedule Your Workflow
Once your Docker image is built and dependencies are installed, schedule your workflow by running:

```bash
pyflyte run --remote workflow.py mlops_workflow
```