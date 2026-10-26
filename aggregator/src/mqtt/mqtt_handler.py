import math
import time
import json
from urllib.parse import parse_qs

REQUIRED_FIELDS = ("temperature", "humidity", "pressure", "timestamp")

def on_connect(client, userdata, flags, rc):
    print(f"[MQTT] Connected with result code {rc}")
    client.subscribe(userdata["topic"])

def parse_payload(payload):
    # Try JSON first
    try:
        return json.loads(payload)
    except json.JSONDecodeError:
        pass

    # Fallback to URL-encoded parsing (Waspmote gateway format)
    # parse_qs gives lists, e.g. {'temp': ['23.45'], ...}
    params = parse_qs(payload, strict_parsing=False)
    try:
        data = {
            "temperature": float(params.get("temp", [None])[0]),
            "humidity":    float(params.get("hum",  [None])[0]),
            "pressure":    float(params.get("press",[None])[0]),
            "timestamp":   time.time()
        }
    except (TypeError, ValueError):
        return None

    if "lat" in params and "lon" in params:
        try:
            data["latitude"] = float(params["lat"][0])
            data["longitude"] = float(params["lon"][0])
        except ValueError:
            pass
    if "dev" in params:
        data["device_id"] = params["dev"][0]
    return data

def is_valid(data):
    if not isinstance(data, dict) or not all(k in data for k in REQUIRED_FIELDS):
        return False
    # float() accepts "nan"/"inf"; those would make the batch file invalid JSON
    # and break minting, so reject them here.
    try:
        return all(math.isfinite(float(data[k])) for k in REQUIRED_FIELDS)
    except (TypeError, ValueError):
        return False

def on_message(client, userdata, msg):
    from collector.data_collector import add_reading

    payload = msg.payload.decode(errors="replace")
    data = parse_payload(payload)

    if is_valid(data):
        count = add_reading(data)
        print(f"[+] Collected reading #{count} → {data}")
    else:
        print(f"[!] Skipping unrecognized payload: {payload}")
