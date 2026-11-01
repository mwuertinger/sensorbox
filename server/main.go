package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/mwuertinger/sensorbox/server/pb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/proto"
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

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	handler := instrumentHandler(reg, mux)

	server := &http.Server{
		Addr:         config.Http.Listen,
		Handler:      handler,
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
		log.Printf("http: %v", err)
	}
	app.wg.Wait() // wait for all goroutines to finish
	if err := influxClient.Close(); err != nil {
		log.Printf("influx: %v", err)
	}
}

// maxRequestSize limits the size of request bodies. Requests sent by sensorboxes are ~100 bytes.
const maxRequestSize = 4096

// instrumentHandler wraps handler with Prometheus HTTP server metrics.
func instrumentHandler(reg prometheus.Registerer, handler http.Handler) http.Handler {
	const namespace = "sensorbox_server"
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Number of HTTP requests by status code and method.",
	}, []string{"code", "method"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request latency.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"code", "method"})
	requestSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_size_bytes",
		Help:      "HTTP request size.",
		Buckets:   prometheus.ExponentialBuckets(64, 4, 6),
	}, []string{"code", "method"})
	responseSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_response_size_bytes",
		Help:      "HTTP response size.",
		Buckets:   prometheus.ExponentialBuckets(64, 4, 6),
	}, []string{"code", "method"})
	reg.MustRegister(requests, duration, requestSize, responseSize)

	return promhttp.InstrumentHandlerCounter(requests,
		promhttp.InstrumentHandlerDuration(duration,
			promhttp.InstrumentHandlerRequestSize(requestSize,
				promhttp.InstrumentHandlerResponseSize(responseSize, handler))))
}

type application struct {
	config       *Config
	influxClient influxdb.Client
	wg           sync.WaitGroup // used to wait for pending goroutines
}

func (app *application) httpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestSize))
	if err != nil {
		log.Printf("read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var request pb.Request
	if err := proto.Unmarshal(buf, &request); err != nil {
		log.Printf("unmarshal: %v", err)
		w.WriteHeader(http.StatusBadRequest)
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

	if _, err := w.Write(data); err != nil {
		// Headers have already been sent at this point, nothing else to do.
		log.Printf("write response: %v", err)
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
	if subtle.ConstantTimeCompare([]byte(dev.AuthToken), []byte(request.AuthToken)) != 1 {
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

// defaultInfluxTimeout is used if no timeout is configured. Without a timeout, a hanging InfluxDB would cause
// goroutines to pile up.
const defaultInfluxTimeout = 10 * time.Second

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
	if config.Influx.Timeout == 0 {
		config.Influx.Timeout = defaultInfluxTimeout
	}
	return &config, nil
}
