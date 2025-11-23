#include <WaspSensorEvent_v30.h>
#include <WaspXBee802.h>
#include <WaspWIFI_PRO.h>
#include <MQTTPacket.h>
#include <MQTTPublish.h>
#include <MQTTFormat.h>

// ──────────────────────────────────────────────
// XBee configuration (unchanged)
uint8_t channel = 0x0F;
uint8_t panID[2] = {0x12, 0x34};
uint8_t encryptionMode = 0;

// Timing constants
const int  RECEIVE_TIMEOUT = 9000;   // ms
const int  SEND_INTERVAL   = 10000;  // ms
unsigned long lastSendTime = 0;

// Network configuration
char ESSID[]      = "Karlo’s iPhone";
char PASSW[]      = "kkkkkkkk";
char RX_ADDRESS[] = "0013A200412A1342";

// MQTT Configuration
char HOST[]         = "172.20.10.3";
char REMOTE_PORT[]  = "1883";
char LOCAL_PORT[]   = "1883";
const char* MQTT_CLIENT_ID   = "waspmote-gateway";
const char* MQTT_CLIENT_ID_2 = "waspmote-endnode";
const char* MQTT_TOPIC_BASE  = "sensor/data";

// System variables
uint8_t  socket        = SOCKET1;
uint8_t  error;
bool     wifiInitialized = false;
uint16_t socket_handle = 0;

// Parsed-sensor storage
float  temperature, humidity, pressure;
double lat =  40.415875;
double lon =  -3.707790;

// Real physical-sensor storage
float  real_temperature, real_humidity, real_pressure;
double real_latitude  = 40.416775;
double real_longitude = -3.703790;
uint8_t status;

// ──────────────────────────────────────────────
// Helper: dump bytes as hex over USB
void dumpHex(const unsigned char* buf, int length) {
  for (int i = 0; i < length; ++i) {
    uint8_t b = buf[i];
    if (b < 0x10) USB.print('0');
    USB.print(b, HEX);
    USB.print(' ');
  }
  USB.println();
}

// ──────────────────────────────────────────────
// SETUP
void setup() {
  USB.ON();
  USB.println(F("\n\n========================"));
  USB.println(F("   MQTT Gateway Start   "));
  USB.println(F("========================\n"));
  initializeSystem();
}

// ──────────────────────────────────────────────
// INITIALISATION
void initializeSystem() {
  USB.print(F("XBee init ... "));
  xbee802.ON(SOCKET0);
  xbee802.setEncryptionMode(encryptionMode);
  xbee802.setPAN(panID);
  xbee802.setChannel(channel);
  xbee802.writeValues();
  USB.println(F("OK"));

  if (!wifiInitialized) {
    USB.print(F("WiFi init ... "));
    error = WIFI_PRO.ON(socket);
    if (error == 0) {
      WIFI_PRO.resetValues();
      WIFI_PRO.setESSID(ESSID);
      WIFI_PRO.setPassword(WPA2, PASSW);
      WIFI_PRO.softReset();
      wifiInitialized = true;
      USB.println(F("OK"));
      WIFI_PRO.ping(HOST);
      USB.print(F("Ping result : "));
      USB.println(error == 0 ? F("OK") : F("FAIL"));
    } else {
      USB.println(F("ERROR"));
    }
  }
}

// ──────────────────────────────────────────────
// PUBLISH HELPERS
bool publishToChannel(const char* clientId,
                      const char* topic,
                      float temp, float hum, float press,
                      float lat,  float lon)
{
  USB.println(F("\n========== MQTT Publish =========="));
  Utils.setLED(LED0, LED_ON);

  if (!WIFI_PRO.isConnected()) {
    USB.println(F("Status      : WiFi NOT connected"));
    Utils.setLED(LED0, LED_OFF);
    USB.println(F("==================================\n"));
    return false;
  }

  // Open TCP socket
  error = WIFI_PRO.setTCPclient(HOST, REMOTE_PORT);
  USB.print(F("TCP connect : "));
  USB.print(error == 0 ? F("OK") : F("FAIL"));
  USB.print(F("   (code="));
  USB.print(error);
  USB.println(F(")"));

  socket_handle = WIFI_PRO._socket_handle;

  // --- MQTT CONNECT packet ---
  MQTTPacket_connectData data = MQTTPacket_connectData_initializer;
  unsigned char buf[200];
  int buflen = sizeof(buf);

  data.clientID.cstring  = (char*)clientId;
  data.keepAliveInterval = 60;
  data.cleansession      = 1;

  int len = MQTTSerialize_connect(buf, buflen, &data);

  error = WIFI_PRO.send(socket_handle, buf, len);
  if (error != 0) {
    USB.println(F("Status      : MQTT CONNECT failed"));
    Utils.setLED(LED0, LED_OFF);
    USB.println(F("==================================\n"));
    return false;
  }

  // --- BUILD PAYLOAD ---
  char tempBuf [10], humBuf [10], pressBuf[12];
  char latBuf  [16], lonBuf  [16];

  dtostrf(temp,  6, 2, tempBuf);
  dtostrf(hum,   6, 1, humBuf);
  dtostrf(press, 8, 2, pressBuf);
  dtostrf(lat,  11, 6, latBuf);
  dtostrf(lon,  11, 6, lonBuf);

  char payload[200];
  snprintf(payload, sizeof(payload),
           "temp=%s&hum=%s&press=%s&lat=%s&lon=%s",
           tempBuf, humBuf, pressBuf, latBuf, lonBuf);

  USB.print(F("Client ID   : ")); USB.println(clientId);
  USB.print(F("Topic       : ")); USB.println(topic);
  USB.print(F("Payload     : ")); USB.println(payload);

  // --- PUBLISH + DISCONNECT ---
  MQTTString topicString = MQTTString_initializer;
  topicString.cstring    = (char*)topic;
  int payloadlen         = strlen(payload);

  len = MQTTSerialize_publish(buf, buflen,
                              0, 0, 0, 1,
                              topicString,
                              (unsigned char*)payload, payloadlen);
  len += MQTTSerialize_disconnect(buf + len, buflen - len);

  USB.println(F("=== MQTT PUBLISH + DISCONNECT packet (hex) ==="));

  error = WIFI_PRO.send(socket_handle, buf, len);
  WIFI_PRO.closeSocket(socket_handle);
  Utils.setLED(LED0, LED_OFF);

  USB.print(F("Publish     : "));
  USB.println(error == 0 ? F("SUCCESS") : F("FAIL"));
  USB.println(F("==================================\n"));

  if (error == 0) {
    Utils.blinkGreenLED();
    return true;
  } else {
    Utils.blinkRedLED();
    return false;
  }
}

// ──────────────────────────────────────────────
// TOP-LEVEL PUBLISH CALL
bool publishSensorData() {
  readSensorData();  // refresh real_*

  bool ok  = publishToChannel(MQTT_CLIENT_ID_2, MQTT_TOPIC_BASE,
                              real_temperature, real_humidity, real_pressure,
                              real_latitude,   real_longitude);

  ok &= publishToChannel(MQTT_CLIENT_ID, MQTT_TOPIC_BASE,
                         temperature, humidity, pressure,
                         lat, lon);

  return ok;
}

// ──────────────────────────────────────────────
// MAIN LOOP
void loop() {
  static bool dataReady = false;

  if (receiveAndParseData()) {
    dataReady = true;
  }

  if (dataReady && (millis() - lastSendTime >= SEND_INTERVAL)) {
    if (publishSensorData()) {
      dataReady    = false;
      lastSendTime = millis();
    }
  }

  Events.attachInt();
  PWR.clearInterruptionPin();
}

// ──────────────────────────────────────────────
// XBEE RECEIVE + PARSE
bool receiveAndParseData() {
  Utils.setLED(LED0, LED_ON);
  error = xbee802.receivePacketTimeout(RECEIVE_TIMEOUT);
  Utils.setLED(LED0, LED_OFF);

  if (error != 0) {
    USB.println(F("XBee        : No packet"));
    Utils.blinkRedLED();
    return false;
  }

  char copyData[200];
  strncpy(copyData, (char*)xbee802._payload, sizeof(copyData) - 1);
  copyData[sizeof(copyData) - 1] = '\0';

  USB.print(F("\n[XBee] Raw  : "));
  USB.println(copyData);

  // Skip initial "0 "
  char *ptr = strchr(copyData, ' ');
  if (ptr) { ptr++; } else { return false; }

  char *token = strtok(ptr, ",");
  while (token) {
    char key[4], value[20];
    if (sscanf(token, "%[^:]:%s", key, value) == 2) {
      float parsed_value = atof(value);

      if (strcmp(key, "T") == 0)        temperature = parsed_value;
      else if (strcmp(key, "H") == 0)   humidity    = parsed_value;
      else if (strcmp(key, "P") == 0)   pressure    = parsed_value;
    }
    token = strtok(NULL, ",");
  }

  printSensorData();          // pretty table
  Utils.blinkGreenLED();
  return true;
}

// ──────────────────────────────────────────────
// READ PHYSICAL SENSORS + PRETTY PRINT
void readSensorData() {
  real_temperature = Events.getTemperature();
  real_humidity    = Events.getHumidity();
  real_pressure    = Events.getPressure();

  char tStr[10], hStr[10], pStr[12];
  dtostrf(real_temperature, 6, 2, tStr);
  dtostrf(real_humidity,    6, 1, hStr);
  dtostrf(real_pressure,    8, 2, pStr);

  USB.println(F("\n====== Real Physical Sensor Readings ======"));
  USB.println(F("+--------------+-----------+"));
  USB.println(F("| Sensor       | Value     |"));
  USB.println(F("+--------------+-----------+"));
  USB.print (F("| Temperature  | ")); USB.print(tStr); USB.println(F(" °C |"));
  USB.print (F("| Humidity     | ")); USB.print(hStr); USB.println(F(" %  |"));
  USB.print (F("| Pressure     | ")); USB.print(pStr); USB.println(F(" Pa |"));
  USB.println(F("+--------------+-----------+\n"));
}

// ──────────────────────────────────────────────
// PRINT PARSED (XBEE) DATA
void printSensorData() {
  char tStr[10], hStr[10], pStr[15];
  dtostrf(temperature, 6, 2, tStr);
  dtostrf(humidity,    6, 1, hStr);
  dtostrf(pressure,    8, 2, pStr);

  USB.println(F("\n========== Parsed XBee Data =========="));
  USB.println(F("+--------------+-----------+"));
  USB.println(F("| Sensor       | Value     |"));
  USB.println(F("+--------------+-----------+"));
  USB.print (F("| Temperature  | ")); USB.print(tStr); USB.println(F(" °C |"));
  USB.print (F("| Humidity     | ")); USB.print(hStr); USB.println(F(" %  |"));
  USB.print (F("| Pressure     | ")); USB.print(pStr); USB.println(F(" Pa |"));
  USB.println(F("+--------------+-----------+\n"));
}

