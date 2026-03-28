package timeseries

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/KaningNoppasin/embedded-system-lab-backend/app/mqtt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const (
	defaultInfluxURL    = "http://localhost:8086"
	defaultInfluxToken  = "embedded-lab-token"
	defaultInfluxOrg    = "embedded-lab"
	defaultInfluxBucket = "telemetry"
	writeTimeout        = 5 * time.Second
	healthTimeout       = 10 * time.Second
	maxHealthRetries    = 10
)

type InfluxWriter struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
}

func NewInfluxWriter() (*InfluxWriter, error) {
	url := getEnv("INFLUXDB_URL", defaultInfluxURL)
	token := getEnv("INFLUXDB_TOKEN", defaultInfluxToken)
	org := getEnv("INFLUXDB_ORG", defaultInfluxOrg)
	bucket := getEnv("INFLUXDB_BUCKET", defaultInfluxBucket)

	client := influxdb2.NewClient(url, token)

	if err := waitForHealthyClient(client); err != nil {
		client.Close()
		return nil, err
	}

	return &InfluxWriter{
		client:   client,
		writeAPI: client.WriteAPIBlocking(org, bucket),
	}, nil
}

func (w *InfluxWriter) StoreTemperature(topic string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	point := influxdb2.NewPoint(
		"temperature",
		map[string]string{"topic": topic},
		map[string]interface{}{"value": value},
		time.Now(),
	)

	return w.writeAPI.WritePoint(ctx, point)
}

func (w *InfluxWriter) StoreINA219(topic string, payload mqtt.INA219Payload) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	point := influxdb2.NewPoint(
		"ina219",
		map[string]string{"topic": topic},
		map[string]interface{}{
			"bus_voltage_v":    payload.BusVoltageV,
			"shunt_voltage_mv": payload.ShuntVoltageMv,
			"load_voltage_v":   payload.LoadVoltageV,
			"current_ma":       payload.CurrentMa,
		},
		time.Now(),
	)

	return w.writeAPI.WritePoint(ctx, point)
}

func (w *InfluxWriter) Close() {
	if w == nil || w.client == nil {
		return
	}

	w.client.Close()
}

func waitForHealthyClient(client influxdb2.Client) error {
	var lastErr error

	for range maxHealthRetries {
		ctx, cancel := context.WithTimeout(context.Background(), healthTimeout)
		_, err := client.Health(ctx)
		cancel()
		if err == nil {
			return nil
		}

		lastErr = err
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("influxdb health check failed after %d attempts: %w", maxHealthRetries, lastErr)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
