package main

import (
	"encoding/binary"
	"flag"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"os"

	"github.com/mwuertinger/sensorbox/config-tool/pb"
)

func main() {
	devId := flag.Uint("dev_id", 0, "device ID")
	wlanSSID := flag.String("wlan_ssid", "", "WLAN SSID")
	wlanPasswd := flag.String("wlan_passwd", "", "WLAN password")
	mqttAddr := flag.String("mqtt_addr", "", "MQTT server address")
	mqttPort := flag.Uint("mqtt_port", 0, "MQTT server port")
	mqttUser := flag.String("mqtt_user", "", "MQTT user")
	mqttPasswd := flag.String("mqtt_passwd", "", "MQTT password")
	mqttPubKeyFile := flag.String("mqtt_pub_key_file", "", "Path to file containing MQTT pub key")
	flag.Parse()

	mqttPubKey, err := ioutil.ReadFile(*mqttPubKeyFile)
	if err != nil {
		log.Fatalf("unable to read mqtt_pub_key_file: %v", err)
	}

	config := &pb.ConfigPb{}
	config.DevId = uint32(*devId)
	config.WlanSsid = *wlanSSID
	config.WlanPassword = *wlanPasswd
	config.MqttAddr = *mqttAddr
	config.MqttPort = uint32(*mqttPort)
	config.MqttUser = *mqttUser
	config.MqttPasswd = *mqttPasswd
	config.MqttPubkey = string(mqttPubKey)

	buf, err := proto.Marshal(config)
	if err != nil {
		log.Fatalln("Failed to marshal config: ", err)
	}

	out := make([]byte, len(buf)+4)
	binary.LittleEndian.PutUint32(out, uint32(len(buf)))
	copy(out[4:], buf)

	os.Stdout.Write(out)
}
