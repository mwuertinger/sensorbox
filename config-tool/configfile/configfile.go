// Package configfile implements the on-flash config format read by the firmware (see setupConfig() in
// firmware/src/main.cpp):
//
//	4 bytes  message length (little endian)
//	4 bytes  CRC32 (IEEE) of the message (little endian)
//	n bytes  ConfigPb protocol buffers message
//
// The whole blob is padded to Size bytes, which is the size of the EEPROM area used by the firmware.
package configfile

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash/crc32"

	"google.golang.org/protobuf/proto"

	"github.com/mwuertinger/sensorbox/config-tool/pb"
)

// Size is the size of the config area in flash.
const Size = 1024

const headerSize = 8

// maxStringLength is the nanopb max_length of the string fields in sensorbox.proto (except mqtt_pubkey, whose 1023
// exceeds Size anyway).
// Longer values pass the checksum but fail to decode on the device.
const maxStringLength = 63

// Encode serializes config into the on-flash format.
func Encode(config *pb.ConfigPb) ([]byte, error) {
	if err := validate(config); err != nil {
		return nil, err
	}
	data, err := proto.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty config is rejected by the firmware")
	}
	if len(data) > Size-headerSize {
		return nil, fmt.Errorf("config too large: %d bytes, at most %d bytes allowed", len(data), Size-headerSize)
	}
	out := make([]byte, Size)
	binary.LittleEndian.PutUint32(out[0:], uint32(len(data)))
	binary.LittleEndian.PutUint32(out[4:], crc32.ChecksumIEEE(data))
	copy(out[headerSize:], data)
	return out, nil
}

// Decode parses a config blob in the on-flash format.
func Decode(in []byte) (*pb.ConfigPb, error) {
	if len(in) < headerSize {
		return nil, fmt.Errorf("input too short: %d bytes", len(in))
	}
	length := binary.LittleEndian.Uint32(in[0:4])
	if uint64(length) > uint64(len(in)-headerSize) {
		return nil, fmt.Errorf("invalid message length %d for %d byte input", length, len(in))
	}
	data := in[headerSize : headerSize+int(length)]
	checksum := binary.LittleEndian.Uint32(in[4:8])
	if calculated := crc32.ChecksumIEEE(data); checksum != calculated {
		return nil, fmt.Errorf("checksum mismatch: provided=%d, calculated=%d", checksum, calculated)
	}
	config := &pb.ConfigPb{}
	if err := proto.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return config, nil
}

// validate checks the nanopb size limits from sensorbox.proto.
func validate(c *pb.ConfigPb) error {
	for name, s := range map[string]string{
		"wlan_ssid":     c.WlanSsid,
		"wlan_password": c.WlanPassword,
		"mqtt_addr":     c.MqttAddr,
		"mqtt_user":     c.MqttUser,
		"mqtt_passwd":   c.MqttPasswd,
		"serverUri":     c.ServerUri,
		"authToken":     c.AuthToken,
	} {
		if len(s) > maxStringLength {
			return fmt.Errorf("%s too long: %d bytes, at most %d bytes allowed", name, len(s), maxStringLength)
		}
	}
	if n := len(c.ServerFingerprintSha1); n != 0 && n != sha1.Size {
		return fmt.Errorf("serverFingerprintSha1 must be %d bytes, got %d", sha1.Size, n)
	}
	return nil
}
