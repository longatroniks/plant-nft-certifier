import os
import json
import glob
import requests

PINATA_JWT = os.getenv("PINATA_JWT", "")
PINATA_API = os.getenv("PINATA_API", "https://api.pinata.cloud")

def pinning_enabled():
    return bool(PINATA_JWT)

def pin_file_to_pinata(filepath, expected_cid):
    """Upload a batch file to Pinata so it stays reachable when the local node is offline.

    cidVersion 0 with Pinata's default chunking reproduces the CID the local
    node computed, so the CID already on the ledger resolves through any
    public gateway.
    """
    try:
        with open(filepath, "rb") as f:
            response = requests.post(
                f"{PINATA_API}/pinning/pinFileToIPFS",
                headers={"Authorization": f"Bearer {PINATA_JWT}"},
                files={"file": (os.path.basename(filepath), f)},
                data={
                    "pinataOptions": json.dumps({"cidVersion": 0}),
                    "pinataMetadata": json.dumps({"name": os.path.basename(filepath)}),
                },
                timeout=60,
            )
    except Exception as e:
        print(f"[!] Pinata upload failed: {e}")
        return False

    if response.status_code != 200:
        print(f"[!] Pinata upload failed ({response.status_code}): {response.text}")
        return False

    cid = response.json().get("IpfsHash")
    if cid != expected_cid:
        print(f"[!] Pinata CID {cid} differs from local CID {expected_cid}")
        return False
    print(f"[📌] Pinned to Pinata: {cid}")
    return True

def pin_pending(data_dir):
    """Pin every recorded CID that has no pin marker yet.

    Markers live on disk, so batches certified while offline in the field
    are pinned automatically once the gateway laptop has internet again.
    """
    if not pinning_enabled():
        return
    pinned_dir = os.path.join(data_dir, "pinned")
    os.makedirs(pinned_dir, exist_ok=True)

    for meta_path in sorted(glob.glob(os.path.join(data_dir, "cids", "*.json"))):
        try:
            with open(meta_path) as f:
                meta = json.load(f)
        except (OSError, json.JSONDecodeError):
            continue
        cid = meta.get("cid")
        if not cid or os.path.exists(os.path.join(pinned_dir, cid)):
            continue
        batch_path = os.path.join(data_dir, meta["summary_path"])
        if pin_file_to_pinata(batch_path, cid):
            open(os.path.join(pinned_dir, cid), "w").close()
