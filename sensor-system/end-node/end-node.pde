ee#include <WaspSensorEvent_v30.h>
#include <WaspXBee802.h>

// ──────────────────────────────────────────────
// XBee Configuration
uint8_t channel        = 0x0F;
uint8_t panID[2]       = {0x12, 0x34};
uint8_t encryptionMode = 0;

// Sensor storage
float temperature, humidity, pressure;

// Communication
char    RX_ADDRESS[] = "0013A200417EE503";   // gateway’s 64-bit address
uint8_t error;

// ──────────────────────────────────────────────
// HELPERS
bool readAndValidateSensors(float& t, float& h, float& p)
{
  t = Events.getTemperature();
  h = Events.getHumidity();
  p = Events.getPressure();
  return !(isnan(t) || isnan(h) || isnan(p));
}

void printSensorReadings(float t, float h, float p)
{
  // Format strings
  char tStr[10], hStr[10], pStr[12];
  dtostrf(t, 6, 2, tStr);
  dtostrf(h, 6, 1, hStr);
  dtostrf(p, 8, 2, pStr);

  // Pretty table
  USB.println(F("\n====== Sensor Readings ======"));
  USB.println(F("+--------------+-----------+"));
  USB.println(F("| Sensor       | Value     |"));
  USB.println(F("+--------------+-----------+"));
  USB.print (F("| Temperature  | ")); USB.print(tStr); USB.println(F(" °C |"));
  USB.print (F("| Humidity     | ")); USB.print(hStr); USB.println(F(" %  |"));
  USB.print (F("| Pressure     | ")); USB.print(pStr); USB.println(F(" Pa |"));
  USB.println(F("+--------------+-----------+\n"));
}

void printTxStatus(const char* payload, uint8_t err)
{
  USB.println(F("\n========== XBee Transmit =========="));
  USB.print  (F("Payload     : ")); USB.println(payload);
  USB.print  (F("Status      : "));
  USB.println(err == 0 ? F("SUCCESS") : F("FAIL"));
  if (err != 0) {
    USB.print(F("Error code  : "));
    USB.println(err, DEC);
  }
  USB.println(F("===================================\n"));
}

// ──────────────────────────────────────────────
// SETUP
void setup()
{
  USB.ON();
  USB.println(F("\n========================"));
  USB.println(F("   Sensor Node Start    "));
  USB.println(F("========================"));

  // Initialise XBee
  xbee802.ON(SOCKET0);
  xbee802.setEncryptionMode(encryptionMode);
  xbee802.setPAN(panID);
  xbee802.setChannel(channel);
  xbee802.writeValues();
  xbee802.OFF();

  USB.println(F("Setup complete, entering loop...\n"));
}

// ──────────────────────────────────────────────
// LOOP
void loop()
{
  // 1) READ
  if (!readAndValidateSensors(temperature, humidity, pressure)) {
    USB.println(F("[WARN] Sensor error — skipping this cycle"));
    delay(5000);
    return;
  }
  printSensorReadings(temperature, humidity, pressure);

  // 2) FORMAT PAYLOAD
  char tStr[10], hStr[10], pStr[15], buf[80];
  dtostrf(temperature, 6, 2, tStr);
  dtostrf(humidity,    6, 1, hStr);
  dtostrf(pressure,    8, 2, pStr);
  snprintf(buf, sizeof(buf), "0 T:%s,H:%s,P:%s", tStr, hStr, pStr);

  // 3) SEND
  xbee802.ON(SOCKET0);
  error = xbee802.send(RX_ADDRESS, (uint8_t*)buf, strlen(buf));
  xbee802.OFF();
  printTxStatus(buf, error);

  // 4) WAIT
  delay(10000);
}

