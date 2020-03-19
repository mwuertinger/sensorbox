package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/tarm/serial"
)

var cmdRead = []byte{0xFF, 0x01, 0x86, 0x00, 0x00, 0x00, 0x00, 0x00, 0x79}
var cmdCali = []byte{0xFF, 0x01, 0x87, 0x00, 0x00, 0x00, 0x00, 0x00, 0x78}

func main() {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal("open port: ", err)
	}
	defer s.Close()

	for {
		val, err := getCO2(s)
		if err != nil {
			log.Print("getCO2: ", err)
		}
		log.Printf("CO2: %d", val)
		time.Sleep(60 * time.Second)
	}

	//	if err := calibrate(s); err != nil {
	//		log.Fatal(err)
	//	}
}

func getCO2(s *serial.Port) (val int, err error) {
	log.Print("sending command")
	_, err = s.Write(cmdRead)
	if err != nil {
		return 0, fmt.Errorf("write cmd: %v", err)
	}

	log.Println("reading response")
	var buf [9]byte
	_, err = io.ReadFull(s, buf[:])
	if err != nil {
		return 0, fmt.Errorf("read result: %v", err)
	}
	log.Println("response: ", hex(buf))

	var checksum byte
	for i := 1; i < 8; i++ {
		checksum += buf[i]
	}
	checksum = 0xFF - checksum
	checksum += 1

	if buf[8] != checksum {
		return 0, fmt.Errorf("checksum mismatch: %x != %x", checksum, buf[8])
	}

	return 256*int(buf[2]) + int(buf[3]), nil
}

func hex(buf [9]byte) string {
	var str bytes.Buffer
	for _, b := range buf {
		fmt.Fprintf(&str, "%02x, ", b)
	}
	return str.String()
}

func calibrate(s *serial.Port) error {
	log.Println("sending calibration command")
	_, err := s.Write(cmdCali)
	return err
}
