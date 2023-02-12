load(
    "@platformio_rules//platformio:platformio.bzl",
    "platformio_library",
    "platformio_project",
)

platformio_project(
    name = "sensorbox",
    src = "main.cc",
    board = "esp12e",
    framework = "arduino",
    platform = "espressif8266",
    deps = [
        "@Adafruit_BME280",
#        "@adafruit_seesaw",
#        "@nanopb",
        "@u8g2",
    ],
)
