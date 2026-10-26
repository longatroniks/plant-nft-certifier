import os
import time
from dotenv import load_dotenv
from paho.mqtt.client import Client

from mqtt.mqtt_handler import on_connect, on_message
from collector.data_collector import summarize_batch, get_last_message_time, take_batch
from storage.file_writer import save_batch_to_file
from storage.cid_recorder import record_ipfs_metadata
from ipfs.uploader import upload_to_ipfs
from ipfs.pinata import pin_pending, pinning_enabled

load_dotenv()

BROKER = os.getenv("BROKER", "mqtt-broker")
PORT = int(os.getenv("PORT", 1883))
TOPIC = os.getenv("TOPIC", "sensor/data")
# Seconds of silence after which the current lifecycle is considered finished.
# Must be well above the gateway's send interval and any expected Wi-Fi outage,
# otherwise one field session is split into several certificates.
INACTIVITY_TIMEOUT = int(os.getenv("INACTIVITY_TIMEOUT", 30))
IPFS_RETRY_INTERVAL = 30
PIN_INTERVAL = 60
DATA_DIR = os.getenv("DATA_DIR", "/app/data")

def finalize_batch(batch):
    print("[⏳] No new data — assuming plant was harvested.")
    summary = summarize_batch(batch)
    return save_batch_to_file(summary, batch)

def try_upload(batch_path):
    cid = upload_to_ipfs(batch_path)
    if cid:
        print(f"[NFT-READY] IPFS CID: {cid}")
        record_ipfs_metadata(cid, batch_path)
    return cid

def start_aggregator():
    client = Client(userdata={"topic": TOPIC})
    client.on_connect = on_connect
    client.on_message = on_message
    client.connect(BROKER, PORT, keepalive=60)
    client.loop_start()

    print(f"[...] Aggregator is running (inactivity timeout {INACTIVITY_TIMEOUT}s). Awaiting sensor data...")

    # Batches saved to disk whose IPFS upload failed; retried until they succeed
    # so a temporary IPFS outage does not lose a certificate.
    pending_uploads = []
    last_retry = 0
    last_pin = 0
    if not pinning_enabled():
        print("[!] PINATA_JWT not set: batches stay on the local IPFS node only")

    try:
        while True:
            time.sleep(1)
            last_time = get_last_message_time()
            if last_time and (time.time() - last_time > INACTIVITY_TIMEOUT):
                batch = take_batch()
                if batch:
                    batch_path = finalize_batch(batch)
                    if not try_upload(batch_path):
                        pending_uploads.append(batch_path)
                    last_pin = 0  # pin the new batch right away

            if pending_uploads and time.time() - last_retry > IPFS_RETRY_INTERVAL:
                last_retry = time.time()
                print(f"[↻] Retrying {len(pending_uploads)} pending IPFS upload(s)")
                pending_uploads = [p for p in pending_uploads if not try_upload(p)]

            if time.time() - last_pin > PIN_INTERVAL:
                last_pin = time.time()
                pin_pending(DATA_DIR)
    except KeyboardInterrupt:
        print("\n[✋] Aggregator manually stopped.")
        client.loop_stop()

if __name__ == "__main__":
    start_aggregator()
