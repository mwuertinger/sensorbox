#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <Wire.h>
#include <Adafruit_Sensor.h>
#include <Adafruit_BME280.h>
#include <string.h>
#include "config.h"

Adafruit_BME280 bme;
bool bmeInitialized = false;
WiFiClientSecure client;
PubSubClient mqtt(client);
BearSSL::PublicKey mqttServerPubKey;
time_t lastUpdate = 0;

void onMqttMessage(char* topic, byte* payload, unsigned int length);

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
    if (!mqttServerPubKey.parse((config.mqttServerPubKey))) {
        Serial.println("Parsing pub key failed");
        while(true);
    }
    client.setKnownKey(&mqttServerPubKey);
    mqtt.setClient(client);
    mqtt.setServer(config.mqttServerIP, config.mqttServerPort);
    mqtt.setCallback(onMqttMessage);
    mqtt.subscribe("sensorbox");
    mqtt.subscribe(config.mqttClient);
}

void setupSensors() {
    // default i2c address 0x76
    if (bme.begin(0x76)) {
        bmeInitialized = true;
    } else {
        Serial.println("BME setup failed!");
    }
}

void setup() {
    Serial.begin(9600);
    setupWiFi();
    setupNtp();
    setupMqtt();
    setupSensors();
}

void mqttReconnect() {
    while (!mqtt.connected()) {
        Serial.print("Connecting to MQTT server...");
        if (mqtt.connect(config.mqttClient, config.mqttUser, config.mqttPassword)) {
            Serial.println(" done");
        } else {
            Serial.printf("failed (state=%d), try again in 5 seconds\n", mqtt.state());
            delay(5000);
        }
    }
}

void sensorUpdate() {
    time_t now = time(nullptr);
    if (now - lastUpdate < 60) {
        return;
    }

    if (bmeInitialized) {
        float pressure = bme.readPressure() / 100.0;
        float humidity = bme.readHumidity();
        float temperature = bme.readTemperature();

        char payload[128];
        snprintf(payload, 128, "{\"c\": \"%s\", \"T\": %d, \"p\": %f, \"h\": %f, \"t\": %f}", config.mqttClient, now, pressure, humidity,
                 temperature);
        mqtt.publish("sensorbox/measurements", payload);
    }

    lastUpdate = now;
}

void loop() {
    mqttReconnect();
    mqtt.loop();
    sensorUpdate();
}

void onMqttMessage(char* topic, byte* payload, unsigned int length) {
    char *str = (char*) malloc(length+1);
    memcpy(str, payload, length);
    str[length] = 0;
    Serial.printf("MQTT message (%s): %s\n", topic, str);
    free(str);
}
