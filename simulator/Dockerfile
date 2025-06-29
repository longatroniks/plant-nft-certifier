FROM python:3.10-slim

ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY mqtt_producer.py sensor_log.txt ./

CMD ["python3", "-u", "mqtt_producer.py"]
