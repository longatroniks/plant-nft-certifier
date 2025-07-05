import os
import json
import time

def save_batch_to_file(summary, raw_readings):
    import datetime

    batch_dir = os.path.join(os.getenv("DATA_DIR", "/app/data"), "batches")
    timestamp = int(time.time())
    filename = f"batch_{timestamp}.json"
    path = os.path.join(batch_dir, filename)
    os.makedirs(batch_dir, exist_ok=True)

    enriched = {
        "sensor_readings": raw_readings,
        "aggregated_summary": summary["summary"],
        "aggregated_at": datetime.datetime.utcnow().isoformat() + "Z",
        "device_id": "sensor-001"  # Optional: could also be dynamic
    }

    with open(path, "w") as f:
        json.dump(enriched, f, indent=2)

    print(f"[✓] Saved full batch to {path}")
    return path
