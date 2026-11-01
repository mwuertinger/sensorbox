package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/mwuertinger/sensorbox/server/pb"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <config>", os.Args[0])
	}
	var err error
	config, err := parseConfig(os.Args[1])
	if err != nil {
		log.Fatalf("loading config failed: %v", err)
	}

	influxClient, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:    config.Influx.Server,
		Timeout: config.Influx.Timeout,
	})
	if err != nil {
		log.Fatalf("influx initialization: %v", err.Error())
	}

	mux := http.NewServeMux()
	var app = application{
		config:       config,
		influxClient: influxClient,
	}
	mux.HandleFunc("/sensorbox", app.httpHandler)

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "sensorbox_server",
	})
	if err != nil {
		log.Fatalf("failed to create Prometheus exporter: %v", err)
	}
	mux.Handle("/metrics", pe)

	och := &ochttp.Handler{
		Handler: mux,
	}
	if err := view.Register(ochttp.ServerRequestCountView, ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView, ochttp.ServerLatencyView); err != nil {
		log.Fatalf("failed to register server views for HTTP metrics: %v", err)
	}

	server := &http.Server{
		Addr:         config.Http.Listen,
		Handler:      och,
		ReadTimeout:  config.Http.ReadWriteTimeout,
		WriteTimeout: config.Http.ReadWriteTimeout,
	}
	go func() {
		if err := server.ListenAndServeTLS(config.Http.CertFile, config.Http.KeyFile); err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	sig := <-sigc
	log.Printf("received %v -> shutting down...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), config.Http.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// Handlers may still be running and calling wg.Add, so Wait would be unsafe.
		log.Printf("http: %v", err)
	} else {
		app.wg.Wait() // wait for pending influx writes
	}
	if err := influxClient.Close(); err != nil {
		log.Printf("influx: %v", err)
	}
}

type application struct {
	config       *Config
	influxClient influxdb.Client
	wg           sync.WaitGroup // used to wait for pending goroutines
}

func (app *application) httpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("read body: %v", err)
		w.WriteHeader(500)
		return
	}
	var request pb.Request
	if err := proto.Unmarshal(buf, &request); err != nil {
		log.Printf("unmarshal: %v", err)
		w.WriteHeader(500)
		return
	}

	log.Printf("request from dev %d (%s)", request.DevId, r.RemoteAddr)

	response, err := app.handleRequest(&request)
	if err == unauthorizedErr {
		w.WriteHeader(401)
		return
	}
	if err != nil {
		log.Printf("handleRequest: %v", err)
		w.WriteHeader(500)
		return
	}

	data, err := proto.Marshal(response)
	if err != nil {
		log.Printf("marshal: %v", err)
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Printf("write response: %v", err)
		w.WriteHeader(500)
		return
	}
}

var unauthorizedErr = errors.New("unauthorized")

func (app *application) handleRequest(request *pb.Request) (*pb.Response, error) {
	devId := int(request.DevId)
	dev, ok := app.config.Devices[devId]
	if !ok {
		log.Printf("unknown device: %d", request.DevId)
		return nil, unauthorizedErr
	}
	if dev.AuthToken != request.AuthToken {
		log.Printf("wrong auth token for device: %d", request.DevId)
		return nil, unauthorizedErr
	}

	if request.Measurement != nil {
		app.processMeasurements(request, dev, devId)
	} else {
		log.Printf("device %d did not send measurements", request.DevId)
	}

	return &pb.Response{}, nil
}

func (app *application) processMeasurements(request *pb.Request, dev Device, devId int) {
	// Add must happen before the goroutine is started, otherwise Wait() in main() may miss it.
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := app.writeToInflux(dev.Location, request.Measurement); err != nil {
			log.Printf("write to influx (dev %d): %v", devId, err)
		}
	}()
}

// writeToInflux writes measurements to InfluxDB
func (app *application) writeToInflux(location string, m *pb.Measurement) error {
	// Create a new point batch
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  "sensors",
		Precision: "s",
	})
	if err != nil {
		return err
	}
	tags := map[string]string{"location": location}
	fields := map[string]interface{}{}
	if m.Pressure > 0 {
		fields["pressure"] = m.Pressure
	}
	if m.Humidity > 0 {
		fields["humidity"] = m.Humidity
	}
	if m.Temperature > 0 {
		fields["temperature"] = m.Temperature - 273.15 // convert to °C
	}
	if m.Co2 > 0 {
		fields["co2"] = m.Co2
	}
	if m.SoilMoisture > 0 {
		fields["soil_moisture"] = m.SoilMoisture
	}
	if m.BatteryVoltage > 0 {
		fields["battery_voltage"] = m.BatteryVoltage
	}

	log.Printf("writing to influx: %+v", fields)

	pt, err := influxdb.NewPoint("measurements", tags, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := app.influxClient.Write(bp); err != nil {
		return err
	}
	return nil
}

// Config represents a config file
type Config struct {
	Http    HttpConfig     `yaml:"http"`
	Influx  InfluxConfig   `yaml:"influx"`
	Devices map[int]Device `yaml:"devices"`
}

type HttpConfig struct {
	Listen           string        `yaml:"listen"`
	ReadWriteTimeout time.Duration `yaml:"readWriteTimeout"`
	ShutdownTimeout  time.Duration `yaml:"shutdownTimeout"`
	CertFile         string        `yaml:"cert"`
	KeyFile          string        `yaml:"key"`
}

type InfluxConfig struct {
	Server  string        `yaml:"server"`
	Token   string        `yaml:"token"`
	Timeout time.Duration `yaml:"timeout"`
}

type Device struct {
	Location  string `yaml:"location"`
	AuthToken string `yaml:"authToken"`
}

// parseConfig reads config file at path and returns the content or an error
func parseConfig(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
