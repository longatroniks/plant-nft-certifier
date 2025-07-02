#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/peer-env.sh"

# ✅ Validate required environment variables
REQUIRED_VARS=(
  "CHAINCODE_NAME"
  "CHANNEL_NAME"
  "ORDERER_CA"
  "ORDERER_ADDRESS"
  "ORDERER_TLS_HOSTNAME"
  "CORE_PEER_ADDRESS"
  "CORE_PEER_TLS_ROOTCERT_FILE"
)

for var in "${REQUIRED_VARS[@]}"; do
  if [[ -z "${!var}" ]]; then
    echo "❌ Environment variable '$var' is not set. Exiting."
    exit 1
  fi
done

# Config
CHAINCODE_VERSION="1.0"
CHAINCODE_PATH="/opt/gopath/src/github.com/hyperledger/fabric/chaincode/$CHAINCODE_NAME"

# 🔢 Determine current sequence and increment
CURRENT_SEQUENCE=$(peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" 2>/dev/null \
  | awk -v cc="$CHAINCODE_NAME" '
    $0 ~ "Name: "cc"," { found=1 }
    found && $0 ~ /Sequence:/ { print $2; exit }
  ')

if ! [[ "$CURRENT_SEQUENCE" =~ ^[0-9]+$ ]]; then
  echo "⚠️ Warning: No committed version of chaincode '$CHAINCODE_NAME' found. Starting from sequence 1."
  SEQUENCE=1
else
  SEQUENCE=$((CURRENT_SEQUENCE + 1))
fi

# 🏷 Use label with version suffix based on sequence
CHAINCODE_LABEL="${CHAINCODE_NAME}_v${SEQUENCE}"
PACKAGE_FILE="$CHAINCODE_PATH/${CHAINCODE_LABEL}.tar.gz"

# 🔥 Remove old package
echo "🧹 Removing old package if exists..."
rm -f "$PACKAGE_FILE"

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

echo "🎉 Chaincode '$CHAINCODE_NAME' is now deployed and committed with sequence $SEQUENCE on channel '$CHANNEL_NAME'."
