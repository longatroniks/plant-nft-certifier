# ./aggregator/Dockerfile

FROM python:3.10-slim

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

ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app/src

# Your Compose wrapper will wait on mqtt-broker for you,
# then kick off the aggregator module.
CMD ["sh", "-c", "\
    echo '[…] waiting for mqtt-broker…'; \
    until nc -z mqtt-broker 1883; do sleep 1; done; \
    echo '[✔] mqtt-broker is up, starting aggregator'; \
    python3 -m src.aggregator \
"]
