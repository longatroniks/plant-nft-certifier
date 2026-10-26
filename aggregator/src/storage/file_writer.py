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

    device_ids = sorted({r["device_id"] for r in raw_readings if "device_id" in r})

    enriched = {
        "sensor_readings": raw_readings,
        "aggregated_summary": summary["summary"],
        "start_timestamp": summary["start_timestamp"],
        "end_timestamp": summary["end_timestamp"],
        "reading_count": summary["reading_count"],
        "aggregated_at": datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z"),
        "device_ids": device_ids or ["sensor-001"]
    }

    with open(path, "w") as f:
        # allow_nan=False: fail loudly instead of writing invalid JSON
        json.dump(enriched, f, indent=2, allow_nan=False)

    print(f"[✓] Saved full batch to {path}")
    return path
