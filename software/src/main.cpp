#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <Wire.h>
#include <Adafruit_Sensor.h>
#include <Adafruit_BME280.h>
#include <string.h>
#include <SoftwareSerial.h>
#include <U8g2lib.h>
#include "config.h"

#define LED_GREEN D6
#define LED_YELLOW D5
#define LED_RED D0
#define BTN_DARK D3

float pressure = 0, humidity = 0, temperature = 0;
uint16_t co2 = 0;

Adafruit_BME280 bme;
bool bmeInitialized = false;

SoftwareSerial co2Sensor(13, 15);

PubSubClient mqtt;
time_t lastUpdate = 0;

bool dark = false;
unsigned long darkLastTrigger = 0;

U8G2_SSD1306_128X64_NONAME_F_HW_I2C u8g2(U8G2_R0, U8X8_PIN_NONE);

void onMqttMessage(char *topic, byte *payload, unsigned int length);

void setupLeds() {
    pinMode(LED_GREEN, OUTPUT);
    pinMode(LED_YELLOW, OUTPUT);
    pinMode(LED_RED, OUTPUT);
    digitalWrite(LED_GREEN, LOW);
    digitalWrite(LED_YELLOW, LOW);
    digitalWrite(LED_RED, LOW);
}

void setupDisplay() {
    u8g2.begin();
    u8g2.clearBuffer();
    u8g2.setFontRefHeightExtendedText();
    u8g2.setDrawColor(1);
    u8g2.setFontPosTop();
    u8g2.setFontDirection(0);
}

void updateDisplay() {
    u8g2.clearBuffer();
    if (dark) {
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
    if (dark) {
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

ICACHE_RAM_ATTR void onDark() {
	// debounce by ignoring interrupts for 100ms
	if (millis() - darkLastTrigger < 100) {
		return;
	}
	dark = !dark;
	darkLastTrigger = millis();
	Serial.printf("dark = %d\n\r", dark);

	updateDisplay();
	updateLeds();
}

void setupButton() {
    pinMode(BTN_DARK, INPUT_PULLUP);
    attachInterrupt(digitalPinToInterrupt(BTN_DARK), onDark, FALLING);
}

void setupWiFi() {
    Serial.print("Connecting to WiFi");
    WiFi.setAutoReconnect(true);
    WiFi.begin(config.ssid, config.password);
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
    }
    Serial.println(WiFi.localIP());
}

void setupNtp() {
    configTime(0, 0, "pool.ntp.org", "time.nist.gov");
    Serial.print("Waiting for NTP time sync: ");
    time_t now = time(nullptr);
    while (now < 24 * 3600) {
        delay(500);
        Serial.print(".");
        now = time(nullptr);
    }
    Serial.println(now);
}

void setupMqtt() {
    BearSSL::PublicKey *pubKey = new BearSSL::PublicKey();
    if (!pubKey->parse((config.mqttServerPubKey))) {
        Serial.println("Parsing pub key failed");
        while (true);
    }
    WiFiClientSecure *client = new WiFiClientSecure();
    client->setKnownKey(pubKey);
    mqtt.setClient(*client);
    mqtt.setServer(config.mqttServerIP, config.mqttServerPort);
    mqtt.setCallback(onMqttMessage);
    mqtt.subscribe("sensorbox");
    char mqttClient[64];
    snprintf(mqttClient, 64, "sensorbox%02d", config.devId);
    mqtt.subscribe(mqttClient);
}

void setupSensors() {
    // default i2c address 0x76
    if (bme.begin(0x76)) {
        bmeInitialized = true;
    } else {
        Serial.println("BME setup failed!");
    }

    co2Sensor.begin(9600);
}

void setup() {
    Serial.begin(9600);
    setupLeds();
    setupButton();
    setupDisplay();
    setupWiFi();
    setupNtp();
    setupMqtt();
    setupSensors();
}

void mqttReconnect() {
    while (!mqtt.connected()) {
        Serial.print("Connecting to MQTT server...");
        char mqttClient[64];
        snprintf(mqttClient, 64, "sensorbox%02d", config.devId);
        if (mqtt.connect(mqttClient, config.mqttUser, config.mqttPassword)) {
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

    // CO2 sensor needs 300s to warm up
    if (uptime > 300) {
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

    char payload[128];
    snprintf(payload, 128, "%d,%ld,%ld,%f,%f,%f,%d", config.devId, now, uptime, pressure, humidity, temperature, co2);
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
}

void onMqttMessage(char *topic, byte *payload, unsigned int length) {
    char *str = (char *) malloc(length + 1);
    memcpy(str, payload, length);
    str[length] = 0;
    Serial.printf("MQTT message (%s): %s\n", topic, str);
    free(str);
}
