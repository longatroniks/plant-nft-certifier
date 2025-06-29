# 🌱 Plant NFT Certifier – Quickstart Guide

## 🚧 Project Status
The current version of the project does not yet include automatic NFT minting. However, it successfully:
- Simulates sensor data using a dedicated container.
- Aggregates the data in a separate service.
- Sets up a fully functional Hyperledger Fabric network with chaincode deployed and ready for interaction.
- This lays the groundwork for the next phase of the project: implementing a Java-based “minter” service that will use the official Fabric SDK to mint NFTs from the aggregated data.
- While future updates may include migrating some Python components to Java for consistency, the current architecture remains in Python. This is intentional, as the Fabric network is now reliably operational thanks to the streamlined deployment scripts.

## 📁 Prerequisites

Install and configure the following:

### 🐳 Docker & Docker Compose

Install Docker:
```bash
sudo apt update
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

Install Docker Compose:
```bash
sudo apt install docker-compose
```

Add your user to the Docker group (optional, avoids using `sudo`):
```bash
sudo usermod -aG docker $USER
newgrp docker
```

---

### 🐹 Golang 1.20+

Install Go (example with 1.20):
```bash
wget https://go.dev/dl/go1.20.12.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.20.12.linux-amd64.tar.gz
```

Add Go to your shell:
```bash
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
```

---

### ⚙️ Fabric Binaries

Install the Fabric tools (peer CLI, configtxgen, etc.):
```bash
curl -sSL https://bit.ly/HyperledgerFabric-Installer | bash -s -- 2.5.0 1.5.0
export PATH=$PATH:$PWD/bin
```

---

### 📡 Mosquitto MQTT Broker

Ensure Mosquitto is installed and running on the host machine:

Install:
```bash
sudo apt install mosquitto mosquitto-clients
```

Start the broker:
```bash
sudo systemctl enable mosquitto
sudo systemctl start mosquitto
```

---

## 🏗 1. Generate Crypto Materials & Artifacts

```bash
bash scripts/generate-artifacts.sh
```

This creates:
- `crypto-config/`
- `channel-artifacts/genesis.block`
- `channel-artifacts/sensor-readings-channel.tx`
- `channel-artifacts/PlantOrgMSPanchors.tx`

---

## 🐳 2. Launch the Network

Run the docker-compose.yaml which is found in the root of the project.

```bash
docker-compose up -d
```

✅ Example terminal output:
```plaintext
Creating network "plantnet" with the default driver
Creating volume "plant-nft-certifier_peer0-data" with default driver
Building simulator
...
Creating orderer.plantnet.com        ... done
Creating ipfs                        ... done
Creating sensor-simulator            ... done
Creating peer0.plantorg.plantnet.com ... done
Creating cli                         ... done
Creating aggregator                  ... done
```

---

## 📡 3. Enter the CLI & Create Channel

```bash
docker exec -it cli bash
source scripts/peer-env.sh
```

```bash
bash scripts/create-and-join-channel.sh
```

✅ Example CLI output:
```plaintext
📨 Creating channel...
✅ Channel created. Block saved to /opt/gopath/src/.../sensor-readings-channel.block
➕ Joining peer to the channel...
Successfully submitted proposal to join channel
🔁 Updating anchor peers for PlantOrgMSP...
Successfully submitted channel update
🎉 Peer joined channel 'sensor-readings-channel' and anchor peers updated.
```

---

## 📦 4. Deploy the Chaincode

```bash
bash scripts/deploy-chaincode.sh
```

✅ Example output:
```plaintext
📦 Packaging chaincode with sequence 1...
📥 Installing chaincode...
✅ Found package ID: plantnft_1:1096cdbde3e171...
🛠 Approving chaincode for org...
🔍 Checking commit readiness...
{
  "approvals": {
    "PlantOrgMSP": true
  }
}
🧾 Committing chaincode to the channel...
🎉 Chaincode 'plantnft' is now deployed and committed to channel 'sensor-readings-channel' with sequence 1.
```

---

## 🧪 5. Mint a Test NFT

```bash
peer chaincode invoke   -o "$ORDERER_ADDRESS"   --ordererTLSHostnameOverride "$ORDERER_TLS_HOSTNAME"   --tls   --cafile "$ORDERER_CA"   -C "$CHANNEL_NAME"   -n "$CHAINCODE_NAME"   --peerAddresses "$CORE_PEER_ADDRESS"   --tlsRootCertFiles "$CORE_PEER_TLS_ROOTCERT_FILE"   -c '{"Args":["MintNFT", "testCID123", "testCID123", "1751155319", "{\"temperature\":{\"avg\":22.4,\"min\":21.3,\"max\":23.1}}"]}'
```

✅ Example result:
```plaintext
Chaincode invoke successful. result: status:200
```

---

## 🔍 6. Query the NFT

```bash
peer chaincode query   -C "$CHANNEL_NAME"   -n "$CHAINCODE_NAME"   -c '{"Args":["GetNFT", "testCID123"]}'
```

✅ Expected output:
```json
{
  "id": "testCID123",
  "cid": "testCID123",
  "timestamp": 1751155319,
  "summary": {
    "temperature": {
      "avg": 22.4,
      "min": 21.3,
      "max": 23.1
    }
  }
}
```

---

## ✅ At This Point…

- You have a local Fabric network running.
- Sensor data is simulated via MQTT.
- Aggregator collects and aggregates data.
- Chaincode is deployed and NFTs can be minted.

Let the plants prove their environment! 🌿
