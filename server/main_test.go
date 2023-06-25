package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mwuertinger/sensorbox/server/pb"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
)

type InfluxMock struct {
	mu        sync.Mutex
	writes    []influxdb.BatchPoints
	writeChan chan any
}

func (m *InfluxMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeChan != nil {
		close(m.writeChan)
	}
	m.writeChan = make(chan any)
	m.writes = nil
}

func (m *InfluxMock) Ping(timeout time.Duration) (time.Duration, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return 1 * time.Second, "", nil
}

func (m *InfluxMock) Write(bp influxdb.BatchPoints) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writes = append(m.writes, bp)
	m.writeChan <- true
	return nil
}

func (m *InfluxMock) Query(q influxdb.Query) (*influxdb.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &influxdb.Response{}, nil
}

func (m *InfluxMock) QueryAsChunk(q influxdb.Query) (*influxdb.ChunkedResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &influxdb.ChunkedResponse{}, nil
}

func (m *InfluxMock) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *InfluxMock) WaitForWrite(timeout time.Duration) {
	select {
	case <-time.NewTimer(timeout).C:
		return
	case <-m.writeChan:
		return
	}
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

	influxMock := InfluxMock{}
	app := application{
		config:           &validConfig,
		influxClient:     &influxMock,
		lastNotification: make(map[string]time.Time),
		temperatures:     make(map[string]float32),
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			influxMock.Reset()
			_, err := app.handleRequest(&d.request)
			if d.expectedError {
				assert.Error(t, err, "error expected")
			} else {
				assert.NoErrorf(t, err, "success expected")

				influxMock.WaitForWrite(time.Second)

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
		})
	}
}

type ntfyMock struct {
	msgs []string
}

func (n *ntfyMock) SendNotification(msg string) error {
	n.msgs = append(n.msgs, msg)
	return nil
}

func (n *ntfyMock) Shutdown() {
}

func TestTemperatureAlert(t *testing.T) {
	ntfyMock := ntfyMock{}
	app := application{
		config:           &validConfig,
		influxClient:     &InfluxMock{},
		ntfyClient:       &ntfyMock,
		temperatures:     make(map[string]float32),
		lastNotification: make(map[string]time.Time),
	}

	app.temperatureAlert(Device{Location: "Foo"}, &pb.Measurement{Temperature: 17})
	assert.Emptyf(t, ntfyMock.msgs, "no notifications expected")
	app.temperatureAlert(Device{Location: "Bar"}, &pb.Measurement{Temperature: 18})
	assert.Len(t, ntfyMock.msgs, 1)
	app.temperatureAlert(Device{Location: "Bar"}, &pb.Measurement{Temperature: 19})
	assert.Len(t, ntfyMock.msgs, 1)
	app.temperatureAlert(Device{Location: "Bar"}, &pb.Measurement{Temperature: 16})
	assert.Len(t, ntfyMock.msgs, 1)
}
