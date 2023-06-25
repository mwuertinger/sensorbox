package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var validConfigStr = `---
http:
  listen: ":4000"
  readWriteTimeout: 10s
  shutdownTimeout: 20s
  cert: sensorbox-server.crt
  key: sensorbox-server.key
influx:
  server: http://localhost:8086
devices:
  1:
    location: "Foo"
    authToken: "ooF"
  2:
    location: "Bar"
    authToken: "raB"
temperatureAlertLocations:
- - Foo
  - Bar
`

var validConfig = Config{
	Http: HttpConfig{
		Listen:           ":4000",
		ReadWriteTimeout: 10 * time.Second,
		ShutdownTimeout:  20 * time.Second,
		CertFile:         "sensorbox-server.crt",
		KeyFile:          "sensorbox-server.key",
	},
	Influx: InfluxConfig{
		Server: "http://localhost:8086",
	},
	Devices: map[int]Device{
		1: {
			Location:  "Foo",
			AuthToken: "ooF",
		},
		2: {
			Location:  "Bar",
			AuthToken: "raB",
		},
	},
	TemperatureAlertLocations: [][2]string{{"Foo", "Bar"}},
}

func TestMarshal(t *testing.T) {
	out, err := yaml.Marshal(validConfig)
	if err != nil {
		t.Error(err)
	}
	log.Println(string(out))
}

func TestParseConfig(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "TestParseConfig")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(validConfigStr); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	path := f.Name()
	f.Close()

	config, err := parseConfig(path)
	if err != nil {
		t.Fatalf("parseConfig: %v", err)
	}

	assert.Equal(t, validConfig, *config)
}
