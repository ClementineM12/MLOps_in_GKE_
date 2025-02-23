# Use a lightweight Python image.
FROM python:3.12-slim

# Set working directory.
WORKDIR /app

# (Optional) Copy requirements file if you have one.
# COPY requirements.txt .
# RUN pip install --no-cache-dir -r requirements.txt

# Install Flytekit.
RUN pip install flytekit

# Copy your workflow code into the image.
COPY hello_workflow.py .

# Default command to run your workflow (for local testing).
CMD ["python", "hello_workflow.py"]
