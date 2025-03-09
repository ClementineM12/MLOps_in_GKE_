FROM python:3.12-slim

# Install OS-level dependencies, including libglib2.0-0 which provides libgthread-2.0.so.0
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements file first for caching purposes
COPY requirements.txt .

# Upgrade pip and install Python dependencies
RUN pip install --upgrade pip && pip install -r requirements.txt

# Copy the rest of the source code
COPY . .