package mqttclient

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mqttClient mqtt.Client

func Init(mqtt_url string, topic string, username string, passwd string) bool {
	// Initialize the MQTT client with the provided URL
	// This function is a placeholder and should be implemented
	// to connect to the MQTT broker at the specified URL.

	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqtt_url)
	opts.SetUsername(username)
	opts.SetPassword(passwd)

	mqttClient := mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Error connecting to MQTT broker: %v\n", token.Error())
		return false
	}
	return true
}

func subCallback(c mqtt.Client, msg mqtt.Message) {
	// Handle incoming messages
	fmt.Printf("Received message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
}

func Subscribe(topic string) bool {
	if mqttClient == nil {
		fmt.Println("MQTT client is not initialized")
		return false
	}

	if token := mqttClient.Subscribe(topic, 0, subCallback); token.Wait() && token.Error() != nil {
		fmt.Printf("Error subscribing to topic %s: %v\n", topic, token.Error())
		return false
	}
	return true
}
