import os
import time
from dotenv import load_dotenv
from paho.mqtt.client import Client

from mqtt.mqtt_handler import on_connect, on_message
from collector.data_collector import summarize_batch, readings, get_last_message_time, reset
from storage.file_writer import save_batch_to_file
from storage.cid_recorder import record_ipfs_metadata
from ipfs.uploader import upload_to_ipfs

load_dotenv()

BROKER = os.getenv("BROKER", "mqtt-broker")
PORT = int(os.getenv("PORT", 1883))
TOPIC = os.getenv("TOPIC", "sensor/data")
INACTIVITY_TIMEOUT = 15

def start_aggregator():
    client = Client(userdata={"topic": TOPIC})
    client.on_connect = on_connect
    client.on_message = on_message
    client.connect(BROKER, PORT, keepalive=60)
    client.loop_start()

    print("[...] Aggregator is running. Awaiting sensor data...")

    try:
        while True:
            time.sleep(1)
            last_time = get_last_message_time()
            if last_time and (time.time() - last_time > INACTIVITY_TIMEOUT):
                if readings:
                    print("[⏳] No new data — assuming plant was harvested.")
                    summary = summarize_batch(readings)
                    batch_path = save_batch_to_file(summary, readings)
                    cid = upload_to_ipfs(batch_path)
                    if cid:
                        print(f"[NFT-READY] IPFS CID: {cid}")
                        record_ipfs_metadata(cid, batch_path)
                    reset()
    except KeyboardInterrupt:
        print("\n[✋] Aggregator manually stopped.")
        client.loop_stop()

if __name__ == "__main__":
    start_aggregator()
