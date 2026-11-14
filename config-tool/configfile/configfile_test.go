package configfile

import (
	"bytes"
	"crypto/sha1"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/mwuertinger/sensorbox/config-tool/pb"
)

func TestRoundTrip(t *testing.T) {
	config := &pb.ConfigPb{DevId: 7, WlanSsid: "ssid", AuthToken: "token", HasSensorCo2: true}
	buf, err := Encode(config)
	if err != nil {
		t.Fatal(err)
	}
	if len(buf) != Size {
		t.Fatalf("len = %d, want %d", len(buf), Size)
	}
	decoded, err := Decode(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(config, decoded) {
		t.Fatalf("got %v, want %v", decoded, config)
	}
}

func TestDecodeInvalid(t *testing.T) {
	valid, err := Encode(&pb.ConfigPb{DevId: 1})
	if err != nil {
		t.Fatal(err)
	}
	corrupted := slices.Clone(valid)
	corrupted[8] ^= 0xff
	hugeLength := slices.Clone(valid)
	hugeLength[3] = 0xff

	for name, in := range map[string][]byte{
		"empty":       nil,
		"short":       {1, 2, 3},
		"checksum":    corrupted,
		"huge length": hugeLength,
		"erased":      bytes.Repeat([]byte{0xff}, Size),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(in); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestEncodeInvalid(t *testing.T) {
	long := strings.Repeat("x", maxStringLength+1)
	for name, config := range map[string]*pb.ConfigPb{
		"long ssid":         {WlanSsid: long},
		"long password":     {WlanPassword: long},
		"long server uri":   {ServerUri: long},
		"long auth token":   {AuthToken: long},
		"short fingerprint": {ServerFingerprintSha1: []byte{1, 2, 3}},
		"too large":         {MqttPubkey: strings.Repeat("x", Size)},
		"empty":             {},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Encode(config); err == nil {
				t.Error("expected error")
			}
		})
	}
	if _, err := Encode(&pb.ConfigPb{AuthToken: long[1:], ServerFingerprintSha1: make([]byte, sha1.Size)}); err != nil {
		t.Errorf("max length: %v", err)
	}
}
