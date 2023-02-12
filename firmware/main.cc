#include <Adafruit_BME280.h>
#include "Adafruit_seesaw.h"
#include <Adafruit_Sensor.h>
#include <CRC32.h>
#include <EEPROM.h>
#include <ESP8266WiFi.h>
#include <ESP8266httpUpdate.h>
#include <pb_encode.h>
#include <pb_decode.h>
#include <SoftwareSerial.h>
#include <string.h>
#include <U8g2lib.h>
#include <Wire.h>

#include "src/sensorbox.pb.h"

#define LED_GREEN D6
#define LED_YELLOW D5
#define LED_RED D0
#define BTN_DISPLAY D3

ConfigPb configPb = ConfigPb_init_zero;

char hostname[64];

float pressure = 0, humidity = 0, temperature = 0, batteryVoltage = 0;
uint16_t co2 = 0, soilMoisture = 0;

Adafruit_BME280 bme;
bool bmeInitialized = false;

// 0: off, 1: display+leds, 2: leds only
uint8_t display = 1;

Adafruit_seesaw ss;

SoftwareSerial co2Sensor(13, 15);

time_t lastUpdate = 0;

unsigned long displayLastTrigger = 0;

U8G2_SSD1306_128X64_NONAME_F_HW_I2C u8g2(U8G2_R0, U8X8_PIN_NONE);

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
  uint32_t *pMessageChecksum = (uint32_t *) (EEPROM.getDataPtr() + 4);
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
  Serial.println("setupLeds");
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
  if (display != 1) {
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
  Serial.println("updateLeds");
  if (display == 0) {
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
  display = (display + 1) % 3;
  displayLastTrigger = millis();
  Serial.printf("display = %d\n\r", display);

  updateDisplay();
  updateLeds();
}

void setupButton() {
  if (configPb.hasDisplay) {
    pinMode(BTN_DISPLAY, INPUT_PULLUP);
    attachInterrupt(digitalPinToInterrupt(BTN_DISPLAY), onDisplay, RISING);
  }
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
  setupConfig();
  setupLeds();
  snprintf(hostname, 64, "sensorbox%02d", configPb.devId);
  setupButton();
  setupDisplay();
  setupWiFi();
  setupOta();
  setupSensors();
}

size_t buildRequest(uint8_t *buf, size_t length) {
  Request request = Request_init_zero;
  request.devId = configPb.devId;
  strncpy(request.authToken, configPb.authToken, sizeof(request.authToken));
  request.has_measurement = true;
  request.measurement.pressure = pressure;
  request.measurement.humidity = humidity;
  request.measurement.temperature = temperature;
  request.measurement.co2 = co2;
  request.measurement.soilMoisture = soilMoisture;
  request.measurement.batteryVoltage = batteryVoltage;

  pb_ostream_t stream = pb_ostream_from_buffer(buf, length);
  if (!pb_encode(&stream, Request_fields, &request)) {
    Serial.printf("pb_encode failed: %s\r\n", stream.errmsg);
  }
  Serial.printf("pb_encode: length=%d\r\n", stream.bytes_written);
  return stream.bytes_written;
}

void sendRequest(uint8_t *data, size_t length) {
  BearSSL::WiFiClientSecure client;
  client.setFingerprint(configPb.serverFingerprintSha1);
  // client.setSSLVersion(BR_TLS12, BR_TLS12);
  client.setInsecure();

  HTTPClient http;
  if (!http.begin(client,  "https://192.168.178.2:4000/sensorbox")) {
    Serial.println("http.begin failed");
    return;
  }
  int httpCode = http.POST(data, length);
  if (httpCode != HTTP_CODE_OK) {
    Serial.printf("http.POST: %d\r\n", httpCode);
  } else {
    auto response = http.getString();
    Serial.printf("response size: %d\r\n", response.length());
  }

  http.end();
  client.stop();
}

void updateCo2() {
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
      Serial.printf("CO2 sensor: discarding implausible value: %d\r\n", co2);
      co2 = 0;
    }
  }
}

void sensorUpdate() {
  time_t now = time(nullptr);
  if (now - lastUpdate < 60) { // update every 60s
    return;
  }
  lastUpdate = now;

  if (bmeInitialized) {
    pressure = bme.readPressure() / 100.0;
    humidity = bme.readHumidity();
    temperature = bme.readTemperature() + 273.15; // convert to Kelvin
  }

  co2 = 0;
  if (configPb.hasSensorCo2) {
    updateCo2();
  }

  soilMoisture = 0;
  if (configPb.hasSensorSeesaw) {
    soilMoisture = ss.touchRead(0);
  }

  if (configPb.hasSensorBatteryVoltage) {
    pinMode(D5, OUTPUT);
    digitalWrite(D5, LOW);
    int voltageRaw = analogRead(A0);
    batteryVoltage = float(voltageRaw) / 1023.0 * 4.2086;
    digitalWrite(D5, HIGH);
  }

  auto buf = new uint8_t[128];
  auto length = buildRequest(buf, 128);
  sendRequest(buf, length);
  delete[] buf;

  updateDisplay();
  updateLeds();
}

void loop() {
  sensorUpdate();

  if (configPb.deepSleepMicroSeconds > 0) {
    Serial.printf("Sleeping for %lld us...\r\n", configPb.deepSleepMicroSeconds);
    bme.setSampling(Adafruit_BME280::MODE_SLEEP);
    ESP.deepSleep(configPb.deepSleepMicroSeconds);
  }
}
