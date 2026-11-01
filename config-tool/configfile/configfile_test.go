package configfile

import (
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
	corrupted := append([]byte(nil), valid...)
	corrupted[8] ^= 0xff
	hugeLength := append([]byte(nil), valid...)
	hugeLength[3] = 0xff

	for name, in := range map[string][]byte{
		"empty":       nil,
		"short":       {1, 2, 3},
		"checksum":    corrupted,
		"huge length": hugeLength,
		"erased":      bytesOf(0xff, Size),
	} {
		if _, err := Decode(in); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func bytesOf(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}
