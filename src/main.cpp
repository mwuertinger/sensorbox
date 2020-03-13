#include <Wire.h>
#include <Adafruit_Sensor.h>
#include <Adafruit_BME280.h>
#include <math.h>
#include <ESP8266WiFi.h>
#include <IRremoteESP8266.h>
#include <IRsend.h>
#include <string.h>
#include "credentials.h"

// Turn on for debugging
#define DEBUG_OUTPUT

// I2C addresse
#define ADDRESS_BME 0x76

Adafruit_BME280 bme;
WiFiServer server(80);
IRsend irsend(15);

void setup() {
  unsigned status = bme.begin(ADDRESS_BME);

#ifdef DEBUG_OUTPUT
  irsend.begin();
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
  Serial.println("WiFi connected"); 
  
  server.begin();
  Serial.println("Server started");
  Serial.print("Use this URL to connect: ");
  Serial.print("http://");
  Serial.print(WiFi.localIP());
  Serial.println("/"); 
}

void loop() {
  WiFiClient client = server.available();
  if (!client) {
    return;
  }
  Serial.println("new client");
  while(!client.available()){
    delay(1);
  } 
  String request = client.readStringUntil('\r');
  Serial.println(request);
  client.flush();

  char *method = strtok((char*) request.c_str(), " ");
  char *path = strtok(NULL, " ");

  client.println("HTTP/1.1 200 OK");
  client.println("Content-Type: text/plain");
  client.println("Connection: close");
  client.println("");

  if (strcmp(method, "POST") == 0 && strcmp(path, "/stereo") == 0) {
    irsend.sendAiwaRCT501(0x7F);
    Serial.println("Toggle stereo");
  }
  
  float t = bme.readTemperature();
  float p = bme.readPressure() / 100.0f;
  float h = bme.readHumidity();

  client.printf("%.1f,%.1f,%.1f", t, p, h);
  client.println("");
  client.stop();

#ifdef DEBUG_OUTPUT
    Serial.print("t=");
    Serial.print(t);
    Serial.print(", p=");
    Serial.print(p);
    Serial.print(", h=");
    Serial.print(h);
    Serial.println();
#endif
}
