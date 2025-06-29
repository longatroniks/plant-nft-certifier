#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/peer-env.sh"

# Config
CHAINCODE_NAME="$CHAINCODE_NAME"
CHAINCODE_VERSION="1.0"
CHAINCODE_LABEL="${CHAINCODE_NAME}_1"
CHAINCODE_PATH="/opt/gopath/src/github.com/hyperledger/fabric/chaincode/$CHAINCODE_NAME"
CHANNEL_NAME="$CHANNEL_NAME"
PACKAGE_FILE="$CHAINCODE_PATH/${CHAINCODE_LABEL}.tar.gz"

# 🔢 Detect current committed sequence
CURRENT_SEQUENCE=$(peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" 2>/dev/null \
  | grep -A1 "Name: ${CHAINCODE_NAME}," \
  | grep "Sequence:" \
  | awk '{print $2}')

if [[ -z "$CURRENT_SEQUENCE" ]]; then
  SEQUENCE=1
else
  SEQUENCE=$((CURRENT_SEQUENCE + 1))
fi

echo "📦 Packaging chaincode with sequence $SEQUENCE..."
peer lifecycle chaincode package "$PACKAGE_FILE" \
  --path "$CHAINCODE_PATH" \
  --lang golang \
  --label "$CHAINCODE_LABEL"

echo "📥 Installing chaincode..."
peer lifecycle chaincode install "$PACKAGE_FILE"

echo "🔍 Querying installed chaincodes..."
peer lifecycle chaincode queryinstalled

PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep "$CHAINCODE_LABEL" | sed -E 's/^Package ID: ([^,]+),.*/\1/')

if [[ -z "$PACKAGE_ID" ]]; then
  echo "❌ Could not find package ID after install."
  exit 1
fi

echo "✅ Found package ID: $PACKAGE_ID"

echo "🛠 Approving chaincode for org..."
peer lifecycle chaincode approveformyorg \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --package-id "$PACKAGE_ID" \
  --sequence "$SEQUENCE" \
  --tls \
  --cafile "$ORDERER_CA" \
  --orderer "$ORDERER_ADDRESS" \
  --ordererTLSHostnameOverride "$ORDERER_TLS_HOSTNAME"

echo "🔍 Checking commit readiness..."
peer lifecycle chaincode checkcommitreadiness \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$SEQUENCE" \
  --output json \
  --tls \
  --cafile "$ORDERER_CA"

echo "🧾 Committing chaincode to the channel..."
peer lifecycle chaincode commit \
  -o "$ORDERER_ADDRESS" \
  --ordererTLSHostnameOverride "$ORDERER_TLS_HOSTNAME" \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$SEQUENCE" \
  --tls \
  --cafile "$ORDERER_CA" \
  --peerAddresses "$CORE_PEER_ADDRESS" \
  --tlsRootCertFiles "$CORE_PEER_TLS_ROOTCERT_FILE"

echo "🎉 Chaincode '$CHAINCODE_NAME' is now deployed and committed to channel '$CHANNEL_NAME' with sequence $SEQUENCE."