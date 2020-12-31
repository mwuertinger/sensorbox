# Sensorbox
ESP8266 powered CO2, temperature, humidity and pressure sensor.

Displays the measurements on an OLED display and shows CO2 levels with green, yellow and red LEDs.
In addition measurements are sent via MQTT and collected with mqtt2influx.

## Software
The software is written in C++ and built using platformio.

Building:
```
platformio run
```

Installing to a board connected via USB:
```
platformio run --target upload
```

### Configuration
Configuration must happen in src/config.h before building.

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
![Rendered PCB](/doc/img/pcb-render.webp)
![PCB Front and Back](/doc/img/pcb-front-and-back.webp)
![Assembled PCB](/doc/img/pcb-assembled.webp)
