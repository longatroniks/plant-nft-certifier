import os
import json
import time

def record_ipfs_metadata(cid, summary_path):
    cid_dir = os.path.join(os.getenv("DATA_DIR", "/app/data"), "cids")
    os.makedirs(cid_dir, exist_ok=True)
    timestamp = int(time.time())
    meta_filename = f"cid_{timestamp}.json"
    meta_path = os.path.join(cid_dir, meta_filename)

    metadata = {
        "cid": cid,
        "summary_path": os.path.relpath(summary_path, os.getenv("DATA_DIR", "/app/data"))
    }

    with open(meta_path, "w") as f:
        json.dump(metadata, f, indent=2)

    print(f"[+] Saved CID metadata to {meta_path}")
