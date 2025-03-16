## Build and Run the Workflow

### Build the Docker Image

Manually trigger a build.

### Install Python Dependencies

Before running the workflow, install the required Python packages by running:

```bash
pip install -r requirements.txt
```

### Schedule Your Workflow
Once the Docker image is built and dependencies are installed, schedule your workflow by running:

```bash
pyflyte run --remote workflow.py mlops_workflow
```