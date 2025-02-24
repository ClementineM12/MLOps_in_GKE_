# Use Python 3.12 slim image as the base image
FROM python:3.12-slim

# Update and upgrade OS packages, then install build essentials (if needed)
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends build-essential && \
    rm -rf /var/lib/apt/lists/*

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file first for caching purposes
COPY requirements.txt .

# Upgrade pip and install Python dependencies from requirements.txt
RUN pip install --upgrade pip && pip install -r requirements.txt

# Copy the rest of your application code into the container
COPY . .
