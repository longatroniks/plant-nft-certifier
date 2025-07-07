FROM python:3.10-slim
ENV PYTHONUNBUFFERED=1
WORKDIR /app

# install netcat (for the same Compose-level guard), plus some useful tools
RUN apt-get update \
 && apt-get install -y \
      curl \
      iputils-ping \
      netcat-openbsd \
      git \
      build-essential \
 && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY src/ ./src/
ENV PYTHONPATH=/app/src

# wait for the broker, then launch aggregator
CMD sh -c "\
    echo '[…] waiting for $BROKER:$PORT…'; \
    until nc -z \"$BROKER\" \"$PORT\"; do sleep 1; done; \
    echo '[✔] $BROKER is up, starting aggregator'; \
    python3 -m src.aggregator \
"
