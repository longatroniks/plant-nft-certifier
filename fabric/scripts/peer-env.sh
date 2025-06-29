#!/bin/bash

export ORDERER_ADDRESS=orderer.plantnet.com:7050
export ORDERER_TLS_HOSTNAME=orderer.plantnet.com
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/plantnet.com/orderers/orderer.plantnet.com/tls/ca.crt

export CORE_PEER_LOCALMSPID=PlantOrgMSP
export CORE_PEER_ADDRESS=peer0.plantorg.plantnet.com:7051
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/plantorg.plantnet.com/users/Admin@plantorg.plantnet.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/plantorg.plantnet.com/peers/peer0.plantorg.plantnet.com/tls/ca.crt

export CHANNEL_NAME=sensor-readings-channel
export CHAINCODE_NAME=plantnft
export CHAINCODE_VERSION=1.0
export CHAINCODE_LABEL=${CHAINCODE_NAME}_1
export CHAINCODE_PATH=/opt/gopath/src/github.com/hyperledger/fabric/chaincode/$CHAINCODE_NAME
export PACKAGE_FILE=$CHAINCODE_PATH/${CHAINCODE_LABEL}.tar.gz

export FABRIC_CFG_PATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/config
export CIDS_DIR=/data/cids
export DATA_DIR=/data
