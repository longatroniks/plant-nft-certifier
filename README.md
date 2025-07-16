## Table of Contents

- [Overview](#overview)  
- [Architecture](#architecture)  
- [Prerequisites](#prerequisites)  
- [Getting Started](#getting-started)  
  - [1. Clone the repository](#1-clone-the-repository)  
  - [2. Configuration](#2-configuration)  
  - [3. Generate Fabric artifacts](#3-generate-fabric-artifacts)  
  - [4. Build and start services](#4-build-and-start-services)  
  - [5. Set up Fabric channel & chaincode](#5-set-up-fabric-channel-chaincode)  
  - [6. (Optional) Run sensor simulator](#6-optional-run-sensor-simulator)  
- [Directory Structure](#directory-structure)  
- [Components](#components)  
  - [Aggregator](#aggregator)  
  - [Fabric Network](#fabric-network)  
  - [Minter](#minter)  
  - [Simulator](#simulator)  
- [Usage](#usage)  
- [Troubleshooting](#troubleshooting)  
- [License](#license)  

---

## Overview

This project implements:

1. **Aggregator** – collects MQTT sensor readings, batches them on inactivity, uploads batch JSON to IPFS, records the resulting CIDs both locally and in `/data/cids` & `/data/batches`.  
2. **Fabric Network** – runs a local Hyperledger Fabric 2.5 network with chaincode (`plantnft`) that lets you register IPFS CIDs as NFT assets on-chain.  
3. **Minter** – watches the `/data/cids` folder for new batch CIDs and automatically invokes the Fabric gateway to mint NFTs for each CID, generating a QR code linking back to the IPFS batch.  
4. **Simulator** – optional sensor data producer that publishes fake measurements on the same MQTT topic for local testing.

---

## Architecture

```bash
    [Sensor Simulator]
        ↓ 
        ↓ sends to MQTT topic “sensor/data”
        ↓ 
    [Aggregator] ────> [IPFS] ────> saves data to /data/cids & /data/batches
        ↓ 
        ↓ watched by
        ↓ 
    [Minter] ────> [Fabric Gateway → Chaincode]
        ↓ 
        ↓ 
        ↓
    [QR code]

```

---

## Prerequisites

**Recommended OS & Editor**  
- **Ubuntu 22.04 LTS** (official repos) – this project was developed and tested here.  
- **Visual Studio Code** – optional but recommended. Suggested extensions:  
  - **Python** (ms-python.python)  
  - **Go** (golang.go)  
  - **Docker** (ms-azuretools.vscode-docker)  
  - **YAML** (redhat.vscode-yaml)  

**System Dependencies**  
```bash
# Update package index
sudo apt update

# Basic tools
sudo apt install -y \
  git \
  ca-certificates \
  curl \
  gnupg \
  lsb-release

# Docker setup
sudo mkdir -p /etc/apt/keyrings
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

sudo apt update
sudo apt install -y \
  docker-ce \
  docker-ce-cli \
  containerd.io \
  docker-buildx-plugin \
  docker-compose-plugin

# Verify Docker
sudo docker run hello-world

# Allow non-root Docker usage
sudo usermod -aG docker $USER
newgrp docker
docker run hello-world

# Python 3.11 setup
sudo apt install -y software-properties-common
sudo add-apt-repository ppa:deadsnakes/ppa -y
sudo apt update
sudo apt install -y \
  python3.11 \
  python3.11-venv \
  python3.11-dev
python3.11 --version

# Go 1.23.10 installation
GO_VERSION=1.23.10
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version

# Hyperledger Fabric CLI & samples
curl -sSL https://bit.ly/2ysbOFE | bash -s
echo 'export PATH=$PATH:$HOME/fabric-samples/bin' >> ~/.bashrc
source ~/.bashrc
```

## Getting Started

1. Configuration  
Copy and fill in environment templates:

```bash
cp .env.example .env
```

Adjust any broker addresses, ports, or IPFS endpoints as needed.

2. Generate Fabric artifacts

```bash
cd fabric/scripts
chmod +x generate-artifacts.sh
bash generate-artifacts.sh
cd ../..
```

3. Build and start services

```bash
docker compose down --remove-orphans -v


# For a clean slate
docker system prune -af --volumes

docker compose build [simulator, aggregator,...etc.]

# Test env (Only option if you don't have a Waspmote)
docker compose --env-file .env.sim --profile sim up

# In docker-compose.yaml, configure the IP that can access the mosquitto running on your local machine
docker compose up
```

4. Set up Fabric channel & chaincode

```bash
docker exec -it cli bash

# Load peer environment
source scripts/peer-env.sh

# Create and join channel
scripts/create-and-join-channel.sh

# Package, install, approve & commit
scripts/deploy-chaincode.sh

exit
```

## Directory Structure

```bash
.
├── aggregator
│   ├── Dockerfile
│   ├── requirements.txt
│   └── src
│       ├── mqtt
│       ├── collector
│       ├── ipfs
│       └── storage
├── fabric
│   ├── chaincode/plantnft
│   ├── config          # crypto-config.yaml, configtx.yaml, core.yaml
│   └── scripts         # generate-artifacts.sh, create-and-join-channel.sh, deploy-chaincode.sh
├── minter
│   ├── Dockerfile
│   └── src             # watcher, qr, parser, fabric client
└── simulator
    ├── Dockerfile
    ├── mqtt_producer.py
    └── sensor_log.txt
```

## Components

### Aggregator
Language: Python

Dependencies: paho-mqtt, requests, python-dotenv

Workflow:

- Listen on MQTT sensor/data  
- Buffer readings until 10 s inactivity  
- Save JSON batch to ./data/batches  
- Upload to IPFS  
- Record CID in ./data/cids & .ipfs_metadata.json  

Runs as the aggregator Docker service.

### Fabric Network
Version: Hyperledger Fabric 2.5.x

Chaincode: Go module in fabric/chaincode/plantnft

Key scripts:

- generate-artifacts.sh (crypto & genesis/channel tx)  
- create-and-join-channel.sh  
- deploy-chaincode.sh  

Accessible via the CLI container.

### Minter
Language: Go 1.23.10

Dependencies: Fabric Gateway SDK v1.7.1, go-qrcode, fsnotify

Workflow:

- Watch /data/cids  
- Extract batch CID  
- Invoke Fabric gateway to mint NFT  
- Generate QR code PNG  

Runs in the minter container.

### Simulator
Language: Python

Dependencies: paho-mqtt, python-dotenv

Publishes fake sensor data to sensor/data for testing.

## Usage

Deploy all services (see Getting Started).

Monitor logs:

- Aggregator: docker logs -f aggregator  
- Minter: docker logs -f minter  
- Peer: docker logs -f peer0.plantorg.plantnet.com  

Inspect IPFS: batch JSON under /data/batches/… and CID under /data/cids/….

View NFTs: query via the CLI or Fabric Explorer.

Scan QR codes in minter/qrs/… to access IPFS URLs.

## Troubleshooting

- **cryptogen/configtxgen not found**: Ensure Fabric binaries are in your $PATH.  
- **IPFS upload fails**: Verify IPFS_API_URL in aggregator/.env and IPFS daemon status.  
- **MQTT connection refused**: Check MQTT broker settings in aggregator/.env.  
- **Chaincode errors**: Check CLI logs and rerun deploy-chaincode.sh with correct env vars.  
