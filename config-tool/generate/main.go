// generate writes a config blob that can be flashed to a sensorbox, e.g. with:
//
//	esptool.py --port /dev/ttyUSB0 write_flash 0x03fb000 config.bin
package main

import (
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/mwuertinger/sensorbox/config-tool/configfile"
	"github.com/mwuertinger/sensorbox/config-tool/pb"
)

func main() {
	out := flag.String("out", "", "output file (required)")
	devId := flag.Uint("dev_id", 0, "device ID")
	wlanSSID := flag.String("wlan_ssid", "", "WLAN SSID")
	wlanPasswd := flag.String("wlan_passwd", "", "WLAN password")
	serverUri := flag.String("server_uri", "", "server URI, e.g. https://192.168.178.2:4000/sensorbox")
	authToken := flag.String("auth_token", "", "auth token, must match the server config")
	fingerprint := flag.String("server_fingerprint", "", "SHA1 fingerprint of the server certificate in hex (colons allowed); empty disables certificate verification")
	hasCo2 := flag.Bool("co2", false, "device has a MH-Z14A CO2 sensor")
	hasBME280 := flag.Bool("bme280", false, "device has a BME280 sensor")
	hasSeesaw := flag.Bool("seesaw", false, "device has a seesaw soil moisture sensor")
	hasDisplay := flag.Bool("display", false, "device has a display")
	hasLeds := flag.Bool("leds", false, "device has LEDs")
	hasBattery := flag.Bool("battery", false, "device measures battery voltage")
	deepSleep := flag.Uint64("deep_sleep_us", 0, "deep sleep duration between measurements in microseconds, 0 disables deep sleep")
	flag.Parse()

	if *out == "" {
		flag.Usage()
		os.Exit(2)
	}

	config := &pb.ConfigPb{
		DevId:                   uint32(*devId),
		WlanSsid:                *wlanSSID,
		WlanPassword:            *wlanPasswd,
		ServerUri:               *serverUri,
		AuthToken:               *authToken,
		HasSensorCo2:            *hasCo2,
		HasSensorBME280:         *hasBME280,
		HasSensorSeesaw:         *hasSeesaw,
		HasDisplay:              *hasDisplay,
		HasLeds:                 *hasLeds,
		HasSensorBatteryVoltage: *hasBattery,
		DeepSleepMicroSeconds:   *deepSleep,
	}
	if *fingerprint != "" {
		fp, err := hex.DecodeString(strings.ReplaceAll(*fingerprint, ":", ""))
		if err != nil || len(fp) != 20 {
			log.Fatalf("invalid server_fingerprint: must be 20 bytes in hex")
		}
		config.ServerFingerprintSha1 = fp
	}

	buf, err := configfile.Encode(config)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, buf, 0o600); err != nil {
		log.Fatal(err)
	}
}
