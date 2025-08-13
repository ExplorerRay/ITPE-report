package config

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
)

type Config struct {
	PrometheusURL string
	MQTTURL       string
	MQTTTopic     string
	MQTTUsername  string
	MQTTPassword  string
}

func DefaultConfig() *Config {
	return &Config{
		PrometheusURL: "http://localhost:9090",
		MQTTURL:       "tcp://localhost:1883",
		MQTTTopic:     "your/topic",
		MQTTUsername:  "your-username",
		MQTTPassword:  "your-password",
	}
}

func ParseArgs() *Config {
	const appName = "itpe-report"
	app := kingpin.New(appName, "ITPE report tool - Used to fetch power info from existing exporter and generate report for perf & energy")

	config := DefaultConfig()

	app.Flag("prom-url", "Prometheus URL").StringVar(&config.PrometheusURL)
	app.Flag("mqtt-url", "MQTT broker URL").StringVar(&config.MQTTURL)
	app.Flag("mqtt-topic", "MQTT topic to subscribe to").StringVar(&config.MQTTTopic)
	app.Flag("mqtt-username", "MQTT username").StringVar(&config.MQTTUsername)
	app.Flag("mqtt-password", "MQTT password").StringVar(&config.MQTTPassword)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	return config
}
