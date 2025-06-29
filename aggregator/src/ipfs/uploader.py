import os
import json
import requests

IPFS_API = os.getenv("IPFS_API", "http://127.0.0.1:5001")

def upload_to_ipfs(filepath):
    try:
        with open(filepath, 'rb') as file:
            files = {'file': file}
            response = requests.post(f"{IPFS_API}/api/v0/add", files=files)

        if response.status_code == 200:
            result = response.json()
            cid = result["Hash"]
            print(f"[✓] Uploaded to IPFS: {cid}")
            return cid
        else:
            print(f"[!] IPFS upload failed: {response.text}")
            return None

    except Exception as e:
        print(f"[!] Error uploading to IPFS: {e}")
        return None
