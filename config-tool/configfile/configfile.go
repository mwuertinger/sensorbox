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
	"encoding/binary"
	"fmt"
	"hash/crc32"

	"google.golang.org/protobuf/proto"

	"github.com/mwuertinger/sensorbox/config-tool/pb"
)

// Size is the size of the config area in flash.
const Size = 1024

const headerSize = 8

// Encode serializes config into the on-flash format.
func Encode(config *pb.ConfigPb) ([]byte, error) {
	data, err := proto.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
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
