# Config Tool

Read the config from a device and print it:
```
esptool.py --port /dev/ttyUSB0 read_flash 0x03fb000 0x400 config.bin
go run ./dump config.bin
```

Generate a config (see `go run ./generate -help` for all flags) and write it to a device:
```
go run ./generate -out config.bin -dev_id 1 ...
esptool.py --port /dev/ttyUSB0 write_flash 0x03fb000 config.bin
```
