# Use Python 3.12 slim image as the base image
FROM python:3.12-slim

# Install OS-level dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy requirements file first for caching purposes
COPY requirements.txt .

# Upgrade pip and install Python dependencies
RUN pip install --upgrade pip && pip install -r requirements.txt

# Copy the rest of your application code into the container
COPY . .

# Set environment variable for GCP credentials (if needed)
ENV GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json

# Set the entrypoint to run your Flyte workflow
# Adjust the command if your main file is named differently
ENTRYPOINT ["python", "workflow.py"]
