# Use a lightweight Python base image
FROM python:3.10-slim

# Set working directory inside the container
WORKDIR /app

# Install OS dependencies
RUN apt-get update && apt-get install -y \
    curl \
    iputils-ping \
    netcat-openbsd \
    git \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy the Python requirements and install them
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the source code
COPY src/ ./src/

# Set environment variables
ENV PYTHONUNBUFFERED=1

# This sets the src/ directory as the root for module resolution — so imports like from mqtt.mqtt_handler import on_connect work properly.
ENV PYTHONPATH=/app/src

# Default command (entrypoint)
CMD ["python3", "-m", "src.aggregator"]
