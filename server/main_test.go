package main

import (
	"os"
	"testing"
	"time"

	"github.com/mwuertinger/sensorbox/server/pb"

	influxdb "github.com/influxdata/influxdb1-client/v2"
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

type InfluxMock struct {
	writes []influxdb.BatchPoints
}

func (m *InfluxMock) Reset() {
	m.writes = nil
}

func (m *InfluxMock) Ping(timeout time.Duration) (time.Duration, string, error) {
	return 1 * time.Second, "", nil
}

func (m *InfluxMock) Write(bp influxdb.BatchPoints) error {
	m.writes = append(m.writes, bp)
	return nil
}

func (m *InfluxMock) Query(q influxdb.Query) (*influxdb.Response, error) {
	return &influxdb.Response{}, nil
}

func (m *InfluxMock) QueryAsChunk(q influxdb.Query) (*influxdb.ChunkedResponse, error) {
	return &influxdb.ChunkedResponse{}, nil
}

func (m *InfluxMock) Close() error {
	return nil
}

func TestRequest(t *testing.T) {
	data := []struct {
		request       pb.Request
		expectedError bool
	}{
		{
			request: pb.Request{
				DevId:       1,
				AuthToken:   "ooF",
				Measurement: &pb.Measurement{Co2: 1, Humidity: 2, Pressure: 3, SoilMoisture: 4, Temperature: 5},
			},
			expectedError: false,
		},
		{
			request: pb.Request{
				DevId:     1,
				AuthToken: "foo",
			},
			expectedError: true,
		},
		{
			request: pb.Request{
				DevId:     2,
				AuthToken: "ooF",
			},
			expectedError: true,
		},
		{
			request: pb.Request{
				DevId:     100,
				AuthToken: "",
			},
			expectedError: true,
		},
	}

	var influxMock InfluxMock
	app := app{
		config:       &validConfig,
		influxClient: &influxMock,
	}

	for _, d := range data {
		influxMock.Reset()
		_, err := app.handleRequest(&d.request)
		if d.expectedError {
			assert.Error(t, err, "error expected")
		} else {
			assert.NoErrorf(t, err, "success expected")
			assert.Len(t, influxMock.writes, 1, "1 write expected")
			write := influxMock.writes[0]
			points := write.Points()
			assert.Len(t, points, 1, "1 write expected")
			fields, err := points[0].Fields()
			assert.NoError(t, err)

			assert.Equal(t, float64(d.request.Measurement.Temperature)-273.15, fields["temperature"])
			assert.Equal(t, float64(d.request.Measurement.Pressure), fields["pressure"])
			assert.Equal(t, int64(d.request.Measurement.Co2), fields["co2"])
			assert.Equal(t, float64(d.request.Measurement.SoilMoisture), fields["soil_moisture"])
			assert.Equal(t, float64(d.request.Measurement.Humidity), fields["humidity"])
		}
	}
}
