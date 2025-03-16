FROM mlrun/mlrun

# Copy requirements file first for caching purposes
COPY requirements.txt .

# Upgrade pip and install Python dependencies
RUN pip install --upgrade pip && pip install -r requirements.txt

# Copy the rest of the source code
COPY . .