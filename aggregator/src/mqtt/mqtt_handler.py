import time
import json
from urllib.parse import parse_qs

def on_connect(client, userdata, flags, rc):
    print(f"[MQTT] Connected with result code {rc}")
    client.subscribe(userdata["topic"])

def on_message(client, userdata, msg):
    from collector.data_collector import readings, update_last_message_time

    payload = msg.payload.decode()
    data = None

    # Try JSON first
    try:
        data = json.loads(payload)
    except json.JSONDecodeError:
        # Fallback to URL-encoded parsing
        params = parse_qs(payload, strict_parsing=False)
        # parse_qs gives lists, e.g. {'temp': ['23.45'], ...}
        try:
            data = {
                "temperature": float(params.get("temp", [None])[0]),
                "humidity":    float(params.get("hum",  [None])[0]),
                "pressure":    float(params.get("press",[None])[0]),
                # "latitude":   float(params.get("lat", [None])[0]),
                # "longitude":  float(params.get("lon", [None])[0]),
                "timestamp":   time.time()
            }
        except (TypeError, ValueError):
            data = None

    if data and all(k in data for k in ("temperature", "humidity", "pressure", "timestamp")):
        readings.append(data)
        update_last_message_time()
        print(f"[+] Collected reading #{len(readings)} → {data}")
    else:
        print(f"[!] Skipping unrecognized payload: {payload}")
