package fabric

const (
	mspID         = "PlantOrgMSP"
	basePath      = "/opt/fabric/crypto-config/peerOrganizations/plantorg.plantnet.com"
	certPath      = basePath + "/users/User1@plantorg.plantnet.com/msp/signcerts/User1@plantorg.plantnet.com-cert.pem"
	keyDir        = basePath + "/users/User1@plantorg.plantnet.com/msp/keystore"
	tlsCertPath   = basePath + "/peers/peer0.plantorg.plantnet.com/tls/ca.crt"
	peerEndpoint  = "peer0.plantorg.plantnet.com:7051"
	gatewayPeer   = "peer0.plantorg.plantnet.com"
	channelName   = "sensor-readings-channel"
	chaincodeName = "plantnft"
)
