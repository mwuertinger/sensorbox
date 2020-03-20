#ifndef CONFIG_H
#define CONFIG_H

class Config {
public:
    const char *ssid = "";
    const char *password = "";
    const uint8_t mqttServerIP[4] = {0, 0, 0, 0};
    const int mqttServerPort = 1883;
    const char *mqttClient = "";
    const char *mqttUser = "";
    const char *mqttPassword = "";
    const char *mqttServerPubKey = "";
};

const Config config;

#endif //CONFIG_H
