package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultBrokerURL   = "tcp://localhost:1883"
	defaultTopic       = "esp32/temperature"
	defaultINA219Topic = "esp32/ina219"
	connectTimeout     = 10 * time.Second
)

type Subscriber struct {
	client mqtt.Client
}

type INA219Payload struct {
	BusVoltageV    float64 `json:"bus_voltage_v"`
	ShuntVoltageMv float64 `json:"shunt_voltage_mv"`
	LoadVoltageV   float64 `json:"load_voltage_v"`
	CurrentMa      float64 `json:"current_ma"`
}

type TelemetryStore interface {
	StoreTemperature(topic string, value float64) error
	StoreINA219(topic string, payload INA219Payload) error
}

func NewTemperatureSubscriber(store TelemetryStore) (*Subscriber, error) {
	brokerURL := getEnv("MQTT_BROKER_URL", defaultBrokerURL)
	temperatureTopic := getEnv("MQTT_TOPIC", defaultTopic)
	ina219Topic := getEnv("MQTT_INA219_TOPIC", defaultINA219Topic)
	clientID := getEnv("MQTT_CLIENT_ID", fmt.Sprintf("embedded-lab-api-%d", time.Now().UnixNano()))

	options := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	options.OnConnect = func(client mqtt.Client) {
		token := client.Subscribe(temperatureTopic, 0, func(_ mqtt.Client, msg mqtt.Message) {
			payload := strings.TrimSpace(string(msg.Payload()))
			temperature, err := strconv.ParseFloat(payload, 64)
			if err != nil {
				log.Printf("mqtt message parse failed for topic %s: payload=%q err=%v", msg.Topic(), payload, err)
				return
			}

			if err := store.StoreTemperature(msg.Topic(), temperature); err != nil {
				log.Printf("influxdb write failed for topic %s value=%f: %v", msg.Topic(), temperature, err)
				return
			}

			log.Printf("mqtt temperature received: topic=%s value=%f", msg.Topic(), temperature)
		})

		if token.WaitTimeout(connectTimeout) && token.Error() != nil {
			log.Printf("mqtt subscribe failed for topic %s: %v", temperatureTopic, token.Error())
			return
		}

		log.Printf("mqtt subscribed to topic %s", temperatureTopic)

		token = client.Subscribe(ina219Topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
			payload := strings.TrimSpace(string(msg.Payload()))

			var ina219Payload INA219Payload
			if err := json.Unmarshal([]byte(payload), &ina219Payload); err != nil {
				log.Printf("mqtt ina219 parse failed for topic %s: payload=%q err=%v", msg.Topic(), payload, err)
				return
			}

			if err := store.StoreINA219(msg.Topic(), ina219Payload); err != nil {
				log.Printf("influxdb ina219 write failed for topic %s payload=%q err=%v", msg.Topic(), payload, err)
				return
			}

			log.Printf(
				"mqtt ina219 received: topic=%s bus_voltage_v=%f shunt_voltage_mv=%f load_voltage_v=%f current_ma=%f",
				msg.Topic(),
				ina219Payload.BusVoltageV,
				ina219Payload.ShuntVoltageMv,
				ina219Payload.LoadVoltageV,
				ina219Payload.CurrentMa,
			)
		})

		if token.WaitTimeout(connectTimeout) && token.Error() != nil {
			log.Printf("mqtt subscribe failed for topic %s: %v", ina219Topic, token.Error())
			return
		}

		log.Printf("mqtt subscribed to topic %s", ina219Topic)
	}

	options.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	})

	client := mqtt.NewClient(options)
	token := client.Connect()
	if !token.WaitTimeout(connectTimeout) {
		return nil, fmt.Errorf("mqtt connect timeout after %s", connectTimeout)
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	log.Printf("mqtt connected to broker %s", brokerURL)
	return &Subscriber{client: client}, nil
}

func (s *Subscriber) Close() {
	if s == nil || s.client == nil || !s.client.IsConnected() {
		return
	}

	s.client.Disconnect(250)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
