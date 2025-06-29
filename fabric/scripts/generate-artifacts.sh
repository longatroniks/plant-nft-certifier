#!/bin/bash
set -e

# Get script's real location and set project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(realpath "$SCRIPT_DIR/..")"

# Paths (now relative to the root)
CRYPTO_CONFIG="$PROJECT_ROOT/config/crypto-config.yaml"
CONFIGTX="$PROJECT_ROOT/config/configtx.yaml"
CONFIG_PATH="$PROJECT_ROOT/config"
OUTPUT_DIR="$PROJECT_ROOT/channel-artifacts"
ORG_NAME="PlantOrgMSP"
CHANNEL_NAME="sensor-readings-channel"
GENESIS_PROFILE="PlantGenesis"
CHANNEL_PROFILE="SensorReadingsChannel"

# Ensure necessary tools are installed
for tool in cryptogen configtxgen; do
  if ! command -v $tool &> /dev/null; then
    echo "❌ $tool not found in PATH. Make sure Fabric binaries are installed and added to your PATH."
    exit 1
  fi
done

echo "🔧 Cleaning previous crypto and artifacts..."
rm -rf "$PROJECT_ROOT/crypto-config" "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

echo "🔐 Generating crypto material..."
cryptogen generate --config="$CRYPTO_CONFIG" --output="$PROJECT_ROOT/crypto-config"

echo "📦 Generating genesis block..."
configtxgen -profile "$GENESIS_PROFILE" \
  -channelID system-channel \
  -outputBlock "$OUTPUT_DIR/genesis.block" \
  -configPath "$CONFIG_PATH"

echo "📨 Generating channel creation transaction..."
configtxgen -profile "$CHANNEL_PROFILE" \
  -outputCreateChannelTx "$OUTPUT_DIR/${CHANNEL_NAME}.tx" \
  -channelID "$CHANNEL_NAME" \
  -configPath "$CONFIG_PATH"

echo "📡 Generating anchor peer update for $ORG_NAME..."
configtxgen -profile "$CHANNEL_PROFILE" \
  -outputAnchorPeersUpdate "$OUTPUT_DIR/${ORG_NAME}anchors.tx" \
  -channelID "$CHANNEL_NAME" \
  -asOrg "$ORG_NAME" \
  -configPath "$CONFIG_PATH"

echo "✅ All artifacts generated in $OUTPUT_DIR"
