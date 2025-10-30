FROM python:3.9-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Install Python dependencies
COPY ./python/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Install Google AI Gemini dependencies
RUN pip install --no-cache-dir google-generativeai pillow

# Copy requirements file
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY . .

# Create necessary directories
RUN mkdir -p /app/models /app/cache

# Expose ports (will be overridden by docker-compose)
EXPOSE 5000

# Set Python to run in unbuffered mode
ENV PYTHONUNBUFFERED=1

# Default command (will be overridden by docker-compose)
CMD ["python", "run_services.py"]