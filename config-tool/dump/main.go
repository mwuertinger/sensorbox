package main

import (
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"

	"github.com/mwuertinger/sensorbox/config-tool/pb"
)

func main() {
	in, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	length := binary.LittleEndian.Uint32(in[0:4])

	checksum := binary.LittleEndian.Uint32(in[4:8])
	calculatedChecksum := crc32.ChecksumIEEE(in[8 : length+8])
	if checksum != calculatedChecksum {
		log.Fatalf("Checksum mismatch: provided=%d, calculated=%d", checksum, calculatedChecksum)
	}

	config := &pb.ConfigPb{}
	if err := proto.Unmarshal(in[8:length+8], config); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}

	log.Println(config)
}
