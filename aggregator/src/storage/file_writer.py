import os
import json
import time

def save_batch_to_file(summary):
    batch_dir = os.path.join(os.getenv("DATA_DIR", "/app/data"), "batches")
    timestamp = int(time.time())
    filename = f"batch_{timestamp}.json"
    path = os.path.join(batch_dir, filename)
    os.makedirs(batch_dir, exist_ok=True)

    with open(path, "w") as f:
        json.dump(summary, f, indent=2)

    print(f"[✓] Saved lifecycle batch to {path}")
    return path
