# Sensorbox
[![CI](https://github.com/mwuertinger/sensorbox/actions/workflows/ci.yml/badge.svg)](https://github.com/mwuertinger/sensorbox/actions/workflows/ci.yml)

ESP8266 powered CO2, temperature, humidity and pressure sensor.

Displays the measurements on an OLED display and shows CO2 levels with green, yellow and red LEDs.
Measurements are sent via HTTPS to the [server](server/) which writes them to InfluxDB.
Alerting (temperature and battery) is handled by [temperature-alert](https://github.com/mwuertinger/temperature-alert).

## Software
The firmware is written in C++ and built using platformio. The server and config-tool are written in Go.

Generating protobuf code (required before building anything):
```
pip install nanopb==0.4.9.1 grpcio-tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
./protoc.sh
```

Building the firmware:
```
cd firmware
platformio run
```

Installing to a board connected via USB:
```
platformio run --target upload
```

### Configuration
The configuration is stored in flash and created with the [config-tool](config-tool/):
```
cd config-tool
go run ./generate -out config.bin -dev_id 1 -wlan_ssid ... -wlan_passwd ... -auth_token ... \
  -server_uri https://192.168.178.2:4000/sensorbox -server_fingerprint <sha1 of server cert> -bme280
esptool.py --port /dev/ttyUSB0 write_flash 0x03fb000 config.bin
```

The SHA1 fingerprint of the server certificate can be obtained with:
```
openssl x509 -in sensorbox-server.crt -noout -fingerprint -sha1
```
Devices with a fingerprint reject any other certificate, so reflash their config when the server certificate changes.

## Hardware
The PCB was designed with KiCad. The following parts are required:

### Parts
- 1x D1 Mini NodeMCU with ESP8266-12F
- 1x MH-Z14A NDIR CO2 Sensor
- 1x GY-BME280 Temperature/Humidity/Pressure Sensor
- 1x 128x64 Pixel 0.96in OLED Display with I2C Interface
- 3x LED
- 3x 2N7002 MOSFETs
- 3x 330Ω
- 2x 4.7kΩ
- 1x Button

### Pictures
#### Rendered PCB
![Rendered PCB](/doc/img/pcb-render.webp)

#### PCB Front and Back
![PCB Front and Back](/doc/img/pcb-front-and-back.webp)

#### Assembled PCB
![Assembled PCB](/doc/img/pcb-assembled.webp)
