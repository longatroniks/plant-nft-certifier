import os
import time
import json
import paho.mqtt.client as mqtt
from dotenv import load_dotenv

load_dotenv()

BROKER = os.environ.get("BROKER", "mqtt-broker")
PORT = int(os.environ.get("PORT", 1883))
TOPIC = os.environ.get("TOPIC", "sensor/data")
DELAY = float(os.environ.get("DELAY", 2))

def parse_sensor_blocks(file_path):
    with open(file_path, "r") as f:
        lines = f.readlines()

    readings = []
    current_block = []
    for line in lines:
        line = line.strip()
        if line == "-----------------------------":
            if current_block:
                readings.append(current_block)
                current_block = []
        elif line:
            current_block.append(line)

    if current_block:
        readings.append(current_block)

    return readings

def block_to_dict(block):
    data = {}
    for line in block:
        if line.startswith("Temperature:"):
            data["temperature"] = float(line.split(":")[1].split()[0])
        elif line.startswith("Humidity:"):
            data["humidity"] = float(line.split(":")[1].split()[0])
        elif line.startswith("Pressure:"):
            data["pressure"] = float(line.split(":")[1].split()[0])
        elif line.startswith("Timestamp:"):
            data["timestamp"] = int(line.split(":")[1].split()[0])
    return data

def simulate_sensor():
    client = mqtt.Client()
    client.connect(BROKER, PORT)

    readings = parse_sensor_blocks("olive_tree_data_11_07_2025.txt")

    while True:
        for block in readings:
            data_dict = block_to_dict(block)
            payload = json.dumps(data_dict)
            client.publish(TOPIC, payload)
            print(f"Published: {payload}")
            time.sleep(DELAY)

if __name__ == "__main__":
    simulate_sensor()
