#include <Adafruit_BME280.h>
#include "Adafruit_seesaw.h"
#include <Adafruit_Sensor.h>
#include <CRC32.h>
#include <EEPROM.h>
#include <ESP8266WiFi.h>
#include <ESP8266httpUpdate.h>
#include <pb_encode.h>
#include <pb_decode.h>
#include <PubSubClient.h>
#include <SoftwareSerial.h>
#include <string.h>
#include <U8g2lib.h>
#include <Wire.h>

#include "sensorbox.pb.h"

#define LED_GREEN D6
#define LED_YELLOW D5
#define LED_RED D0
#define BTN_DISPLAY D3

ConfigPb configPb = ConfigPb_init_zero;

char hostname[64];

float pressure = 0, humidity = 0, temperature = 0;
uint16_t co2 = 0, soilMoisture = 0;

Adafruit_BME280 bme;
bool bmeInitialized = false;

Adafruit_seesaw ss;

SoftwareSerial co2Sensor(13, 15);

PubSubClient mqtt;
time_t lastUpdate = 0;

bool display = true;
unsigned long displayLastTrigger = 0;

U8G2_SSD1306_128X64_NONAME_F_HW_I2C u8g2(U8G2_R0, U8X8_PIN_NONE);

void onMqttMessage(char *topic, byte *payload, unsigned int length);

void fatalError(char const *msg) {
  Serial.println(msg);
  if (configPb.hasLeds) {
    while (true) {
      digitalWrite(LED_GREEN, HIGH);
      digitalWrite(LED_YELLOW, LOW);
      digitalWrite(LED_RED, LOW);
      delay(200);
      digitalWrite(LED_GREEN, LOW);
      digitalWrite(LED_YELLOW, HIGH);
      digitalWrite(LED_RED, LOW);
      delay(200);
      digitalWrite(LED_GREEN, LOW);
      digitalWrite(LED_YELLOW, LOW);
      digitalWrite(LED_RED, HIGH);
      delay(200);
    }
  } else {
    for (int i = 0; i < 3; i++) {
      digitalWrite(LED_BUILTIN, HIGH);
      delay(200);
      digitalWrite(LED_BUILTIN, LOW);
      delay(200);
    }
    delay(500);
  }
}

void setupConfig() {
  Serial.println("setupConfig()");

  const size_t EEPROM_SIZE = 1024;
  EEPROM.begin(EEPROM_SIZE);

  // data layout:
  // 4 bytes = message length
  // 4 bytes = CRC checksum
  // protocol buffers message
  uint32_t *pMessageLength = (uint32_t *) EEPROM.getDataPtr();
  uint32_t *pMessageChecksum = (uint32_t * )(EEPROM.getDataPtr() + 4);
  uint8_t *data = EEPROM.getDataPtr() + 8; // skip length and checksum field
  Serial.printf("setupConfig(): *pMessageLength=%d\n\r", *pMessageLength);

  uint32_t checksum = 42;
  if (*pMessageLength < EEPROM_SIZE) {
    Serial.printf("setupConfig(): calculating checksum\n\r");
    CRC32 crc;
    for (size_t i = 0; i < *pMessageLength; i++) {
      crc.update(data[i]);
    }
    checksum = crc.finalize();
  }
  if (checksum != *pMessageChecksum) {
    fatalError("setupConfig: Config checksum error!");
  }

  uint32_t message_length = *pMessageLength;
  Serial.printf("setupConfig: Read message_length=%d\r\n", message_length);
  pb_istream_t stream = pb_istream_from_buffer(data, message_length);
  bool status = pb_decode(&stream, ConfigPb_fields, &configPb);
  if (status) {
    Serial.println("setupConfig: Config decoded successfully");
  } else {
    fatalError("setupConfig: Config decoding failed!");
  }
}

void setupLeds() {
  if (!configPb.hasLeds) {
    return;
  }
  pinMode(LED_GREEN, OUTPUT);
  pinMode(LED_YELLOW, OUTPUT);
  pinMode(LED_RED, OUTPUT);
  digitalWrite(LED_GREEN, LOW);
  digitalWrite(LED_YELLOW, LOW);
  digitalWrite(LED_RED, LOW);
}

void setupDisplay() {
  if (!configPb.hasDisplay) {
    return;
  }

  u8g2.begin();
  u8g2.clearBuffer();
  u8g2.setFontRefHeightExtendedText();
  u8g2.setDrawColor(1);
  u8g2.setFontPosTop();
  u8g2.setFontDirection(0);
}

void updateDisplay() {
  if (!configPb.hasDisplay) {
    return;
  }

  u8g2.clearBuffer();
  if (!display) {
    u8g2.sendBuffer();
    return;
  }

  u8g2.setFont(u8g2_font_courB18_tf);

  char str[128];

  if (co2 != 0) {
    snprintf(str, 128, "%dppm", co2);
  } else {
    snprintf(str, 128, "---");
  }
  u8g2.drawStr(0, 0, str);
  snprintf(str, 128, "%.1fC", temperature - 273.15);
  u8g2.drawStr(0, 23, str);
  snprintf(str, 128, "%.1f%%\n", humidity);
  u8g2.drawStr(0, 46, str);

  u8g2.sendBuffer();
}

void updateLeds() {
  if (!configPb.hasLeds) {
    return;
  }
  if (!display) {
    digitalWrite(LED_GREEN, LOW);
    digitalWrite(LED_YELLOW, LOW);
    digitalWrite(LED_RED, LOW);
  } else if (co2 == 0) {
    digitalWrite(LED_GREEN, LOW);
    digitalWrite(LED_YELLOW, LOW);
    digitalWrite(LED_RED, LOW);
  } else if (co2 < 1000) {
    digitalWrite(LED_GREEN, HIGH);
    digitalWrite(LED_YELLOW, LOW);
    digitalWrite(LED_RED, LOW);
  } else if (co2 < 2000) {
    digitalWrite(LED_GREEN, LOW);
    digitalWrite(LED_YELLOW, HIGH);
    digitalWrite(LED_RED, LOW);
  } else {
    digitalWrite(LED_GREEN, LOW);
    digitalWrite(LED_YELLOW, LOW);
    digitalWrite(LED_RED, HIGH);
  }
}

IRAM_ATTR void onDisplay() {
  // debounce by ignoring interrupts for 100ms
  if (millis() - displayLastTrigger < 100) {
    return;
  }
  display = !display;
  displayLastTrigger = millis();
  Serial.printf("display = %d\n\r", display);

  updateDisplay();
  updateLeds();
}

void setupButton() {
  pinMode(BTN_DISPLAY, INPUT_PULLUP);
  attachInterrupt(digitalPinToInterrupt(BTN_DISPLAY), onDisplay, RISING);
}

void setupWiFi() {
  Serial.print("Connecting to WiFi");

  if (configPb.hasDisplay) {
    u8g2.setFont(u8g2_font_t0_11_tf);
    u8g2.clearBuffer();
    u8g2.drawStr(0, 0, hostname);
    u8g2.drawStr(0, 12, "Connecting to WiFi...");
    u8g2.sendBuffer();
  }

//    IPAddress clientIp(192, 168, 178, 60);
//    IPAddress gateway(192, 168, 178, 1);
//    IPAddress subnet(255, 255, 255, 0);
//    IPAddress primaryDns(192, 168, 178, 1);
//    WiFi.config(clientIp, gateway, subnet, primaryDns, primaryDns);
  WiFi.setAutoReconnect(true);
  WiFi.hostname(hostname);
  WiFi.begin(configPb.wlan_ssid, configPb.wlan_password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println(WiFi.localIP());
  if (configPb.hasDisplay) {
    u8g2.clearBuffer();
    u8g2.drawStr(0, 0, hostname);
    u8g2.drawStr(0, 12, WiFi.localIP().toString().c_str());
    u8g2.sendBuffer();
  }
}

void setupMqtt() {
  BearSSL::PublicKey *pubKey = new BearSSL::PublicKey();
  if (!pubKey->parse((configPb.mqtt_pubkey))) {
    fatalError("Parsing pub key failed!");
  }
  WiFiClientSecure *client = new WiFiClientSecure();
  client->setKnownKey(pubKey);
  mqtt.setClient(*client);
  mqtt.setServer(configPb.mqtt_addr, configPb.mqtt_port);
  mqtt.setCallback(onMqttMessage);
}

void onOtaStart() {
  Serial.println("OTA begin");
  if (configPb.hasDisplay) {
    u8g2.clearBuffer();
    u8g2.setFont(u8g2_font_t0_15_tf);
    u8g2.drawStr(4, 0, "Updating...");
    u8g2.drawFrame(4, 20, 120, 20);
    u8g2.sendBuffer();
  }
}

void onOtaEnd() {
  Serial.println("OTA end");
}

void onOtaProgress(int cur, int total) {
  Serial.printf("OTA progress: %d/%d\r\n", cur, total);
  if (configPb.hasDisplay) {
    u8g2.drawBox(4, 20, int(float(cur) / float(total) * 120.0), 20);
    u8g2.sendBuffer();
  }
}

void onOtaError(int err) {
  Serial.printf("OTA error: %d\r\n", err);

  if (configPb.hasDisplay) {
    char str[128];
    snprintf(str, 128, "Error: %d", err);
    u8g2.drawStr(0, 50, str);
    u8g2.sendBuffer();
  }
}

void setupOta() {
  ESPhttpUpdate.setLedPin(LED_BUILTIN, LOW);
  ESPhttpUpdate.onStart(onOtaStart);
  ESPhttpUpdate.onEnd(onOtaEnd);
  ESPhttpUpdate.onProgress(onOtaProgress);
  ESPhttpUpdate.onError(onOtaError);
}

void setupSensors() {
  if (configPb.hasSensorBME280) {
    // default i2c address 0x76
    if (bme.begin(0x76)) {
      bmeInitialized = true;
    } else {
      Serial.println("BME setup failed!");
    }
  }

  if (configPb.hasSensorCo2) {
    co2Sensor.begin(9600);
  }

  if (configPb.hasSensorSeesaw) {
    ss.begin(0x36);
  }
}

void calibrateCo2() {
  if (!configPb.hasSensorCo2) {
    Serial.printf("No CO2 sensor configured\r\n");
    return;
  }
  Serial.printf("Calibrating CO2 sensor\r\n");
  uint8_t cmdCalibrate[] = {0xFF, 0x01, 0x87, 0x00, 0x00, 0x00, 0x00, 0x00, 0x78};
  size_t res = co2Sensor.write(cmdCalibrate, 9);
  if (res != 9) {
    Serial.printf("CO2 sensor: writing failed: %d\r\n", res);
    return;
  }
  co2Sensor.flush();
}

void setup() {
  Serial.begin(9600);
  setupLeds();
  setupConfig();
  snprintf(hostname, 64, "sensorbox%02d", configPb.devId);
  setupButton();
  setupDisplay();
  setupWiFi();
  setupMqtt();
  setupOta();
  setupSensors();

  pinMode(D5, OUTPUT);
  digitalWrite(D5, LOW);
  int voltageRaw = analogRead(A0);
  float voltage = float(voltageRaw) / 1023.0 * 4.2086;
  digitalWrite(D5, HIGH);
  Serial.printf("voltageRaw = %d, voltage = %f\r\n", voltageRaw, voltage);
}

void mqttReconnect() {
  while (!mqtt.connected()) {
    Serial.print("Connecting to MQTT server...");
    char mqttClient[64];
    snprintf(mqttClient, 64, "sensorbox%02d", configPb.devId);
    if (mqtt.connect(mqttClient, configPb.mqtt_user, configPb.mqtt_passwd)) {
      mqtt.subscribe("sensorbox");
      mqtt.subscribe(hostname);
      Serial.println(" done");
    } else {
      Serial.printf("failed (state=%d), try again in 5 seconds\n", mqtt.state());
      delay(5000);
    }
  }
}

void sensorUpdate() {
  time_t now = time(nullptr);
  if (now - lastUpdate < 60) { // update every 60s
    return;
  }
  lastUpdate = now;

  unsigned long uptime = millis() / 1000; // system uptime in seconds

  if (bmeInitialized) {
    pressure = bme.readPressure() / 100.0;
    humidity = bme.readHumidity();
    temperature = bme.readTemperature() + 273.15; // convert to Kelvin
  }

  if (configPb.hasSensorCo2) {
    uint8_t cmdRead[] = {0xFF, 0x01, 0x86, 0x00, 0x00, 0x00, 0x00, 0x00, 0x79};
    size_t res = co2Sensor.write(cmdRead, 9);
    if (res != 9) {
      Serial.printf("CO2 sensor: writing failed: %d\r\n", res);
      return;
    }
    co2Sensor.flush();

    uint8_t buf[9];
    int n = 0;
    // try to read from serial port up to 10000 times
    // usually data arrives after ~1300 iterations
    for (int i = 0; i < 10000 && n < 9; i++) {
      int val = co2Sensor.read();
      if (val < 0) {
        continue;
      }
      buf[n] = val;
      n++;
    }

    // verify checksum
    uint8_t checksum = 0;
    for (uint8_t i = 1; i < 8; i++) {
      checksum += buf[i];
    }
    checksum = 0xFF - checksum;
    checksum += 1;
    if (buf[8] != checksum) {
      Serial.println("CO2 sensor: checksum error!");
    } else {
      co2 = 256 * buf[2] + buf[3];
      if (co2 < 200 || co2 > 10000) {
        Serial.printf("CO2 sensor: discarding inplausible value: %d\r\n", co2);
        co2 = 0;
      }
    }
  }

  soilMoisture = 0;
  if (configPb.hasSensorSeesaw) {
    soilMoisture = ss.touchRead(0);
  }

  char payload[128];
  snprintf(payload,
           128,
           "%d,%lld,%ld,%f,%f,%f,%d,%d",
           configPb.devId,
           now,
           uptime,
           pressure,
           humidity,
           temperature,
           co2,
           soilMoisture);
  Serial.printf("Publishing: %s\r\n", payload);
  if (!mqtt.publish("sensorbox/measurements", payload)) {
    Serial.println("Publishing failed!");
  }

  updateDisplay();
  updateLeds();
}

void loop() {
  mqttReconnect();
  mqtt.loop();
  sensorUpdate();

  if (configPb.deepSleepMicroSeconds > 0) {
    Serial.printf("Sleeping for %lld us...\r\n", configPb.deepSleepMicroSeconds);
    bme.sleep();
    ESP.deepSleep(configPb.deepSleepMicroSeconds);
  }
}

void onMqttMessage(char *topic, byte *payload, unsigned int length) {
  char *str = (char *) malloc(length + 1);
  memcpy(str, payload, length);
  str[length] = 0;
  Serial.printf("MQTT message (%s): %s\r\n", topic, str);

  if (strcmp(topic, hostname) == 0 && strncmp(str, "ota", length) == 0) {
    WiFiClient client;
    ESPhttpUpdate.update(client, String("hal"), 10000, String("/sensorbox-esp12e.bin"));
  }
  if (strcmp(topic, hostname) == 0 && strncmp(str, "calibrate_co2", length) == 0) {
    calibrateCo2();
  }
  if (strcmp(topic, hostname) == 0 && strncmp(str, "display_on", length) == 0) {
    display = true;
    updateDisplay();
    updateLeds();
  }
  if (strcmp(topic, hostname) == 0 && strncmp(str, "display_off", length) == 0) {
    display = false;
    updateDisplay();
    updateLeds();
  }

  free(str);
}
