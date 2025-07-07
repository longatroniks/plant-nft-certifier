# ./simulator/Dockerfile
FROM python:3.10-slim
ENV PYTHONUNBUFFERED=1
WORKDIR /app

# install netcat so our `until nc -z` guard works
RUN apt-get update \
 && apt-get install -y netcat-openbsd \
 && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY mqtt_producer.py sensor_log.txt ./

# wait for mqtt-broker:1883 to be up, then launch
CMD ["sh", "-c", "\
    echo '[…] waiting for mqtt-broker…'; \
    until nc -z mqtt-broker 1883; do sleep 1; done; \
    echo '[✔] mqtt-broker is up, starting simulator'; \
    python3 -u mqtt_producer.py \
"]
