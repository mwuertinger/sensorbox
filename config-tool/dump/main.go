// dump prints a config read from flash, e.g. with:
//
//	esptool.py --port /dev/ttyUSB0 read_flash 0x03fb000 0x400 config.bin
package main

import (
	"log"
	"os"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/mwuertinger/sensorbox/config-tool/configfile"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <config.bin>", os.Args[0])
	}
	in, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("read file: %v", err)
	}
	config, err := configfile.Decode(in)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.WriteString(prototext.Format(config) + "\n")
}
