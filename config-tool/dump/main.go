package main

import (
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"os"

	"github.com/mwuertinger/sensorbox/config/pb"
)

func main() {
	in, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	length := binary.LittleEndian.Uint32(in[0:4])
	config := &pb.ConfigPb{}
	if err := proto.Unmarshal(in[4:length+4], config); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}

	log.Println(config)
}
