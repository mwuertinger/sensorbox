#include <Wire.h>
#include <Adafruit_Sensor.h>
#include <Adafruit_BME280.h>
#include <math.h>
#include <ESP8266WiFi.h>
#include <ESP8266WebServer.h>
#include <string.h>
#include "credentials.h"

// Turn on for debugging
#define DEBUG_OUTPUT

// I2C addresse
#define ADDRESS_BME 0x76

Adafruit_BME280 bme;
ESP8266WebServer server(80);

void handleMetrics() {
  float t = bme.readTemperature();
  float p = bme.readPressure() / 100.0f;
  float h = bme.readHumidity();

  char metrics[1024];
  snprintf(metrics, 1024, "temperature %f\npressure %f\nhumidity %f\n", t, p, h);
  server.send(200, "text/plain", metrics);
}

void setup() {
  unsigned status = bme.begin(ADDRESS_BME);

#ifdef DEBUG_OUTPUT
  Serial.begin(9600);
  while(!Serial);    // time to get serial running

  if (!status) {
    Serial.println("Could not find a valid BME280 sensor, check wiring, address, sensor ID!");
    Serial.print("SensorID was: 0x"); Serial.println(bme.sensorID(),16);
    Serial.print("        ID of 0xFF probably means a bad address, a BMP 180 or BMP 085\n");
    Serial.print("   ID of 0x56-0x58 represents a BMP 280,\n");
    Serial.print("        ID of 0x60 represents a BME 280.\n");
    Serial.print("        ID of 0x61 represents a BME 680.\n");
    while (1) delay(10);
  }
#endif

  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("");
  Serial.println("WiFi connected: "); 
  Serial.print(WiFi.localIP());
  Serial.println();

  server.on("/metrics", handleMetrics);
  server.begin();
}

void loop() {
  server.handleClient();
  Serial.print(".");
}
