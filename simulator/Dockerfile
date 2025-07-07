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

# wait for the broker, then launch producer
CMD sh -c "\
    echo '[…] waiting for $BROKER:$PORT…'; \
    until nc -z \"$BROKER\" \"$PORT\"; do sleep 1; done; \
    echo '[✔] $BROKER is up, starting simulator'; \
    python3 -u mqtt_producer.py \
"
