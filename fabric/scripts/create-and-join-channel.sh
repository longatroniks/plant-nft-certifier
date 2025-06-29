#!/bin/bash

set -e

# Load peer + orderer environment variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/peer-env.sh"

# Channel config
CHANNEL_NAME="$CHANNEL_NAME"
OUTPUT_DIR="$SCRIPT_DIR/../channel-artifacts"
ORDERER_ADDRESS="$ORDERER_ADDRESS"
ORDERER_HOSTNAME="$ORDERER_TLS_HOSTNAME"
ANCHOR_UPDATE_TX="${OUTPUT_DIR}/PlantOrgMSPanchors.tx"
CHANNEL_TX="${OUTPUT_DIR}/${CHANNEL_NAME}.tx"
CHANNEL_BLOCK="${OUTPUT_DIR}/${CHANNEL_NAME}.block"

echo "📨 Creating channel..."
peer channel create \
  -o "$ORDERER_ADDRESS" \
  --ordererTLSHostnameOverride "$ORDERER_HOSTNAME" \
  -c "$CHANNEL_NAME" \
  -f "$CHANNEL_TX" \
  --outputBlock "$CHANNEL_BLOCK" \
  --tls \
  --cafile "$ORDERER_CA"

echo "✅ Channel created. Block saved to $CHANNEL_BLOCK"

echo "➕ Joining peer to the channel..."
peer channel join \
  -b "$CHANNEL_BLOCK"

echo "🔁 Updating anchor peers for PlantOrgMSP..."
peer channel update \
  -o "$ORDERER_ADDRESS" \
  --ordererTLSHostnameOverride "$ORDERER_HOSTNAME" \
  -c "$CHANNEL_NAME" \
  -f "$ANCHOR_UPDATE_TX" \
  --tls \
  --cafile "$ORDERER_CA"

echo "🎉 Peer joined channel '$CHANNEL_NAME' and anchor peers updated."
