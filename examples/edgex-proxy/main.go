package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// https://github.com/eclipse/paho.mqtt.golang/blob/master/cmd/stdoutsub/main.go
func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	// $iothub/methods/POST/Values/?$rid=1
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	topic := message.Topic()
	rid := topic[len(topic)-1 : len(topic)]
	fmt.Println("rid: " + rid)

	m, err := regexp.Compile("POST/(.*?)/")
	if err != nil {
		fmt.Println(err.Error())
	}
	methodName := m.FindString(topic)
	methodName = strings.Replace(methodName, "POST", "", 1)
	methodName = strings.Replace(methodName, "/", "", -1)
	fmt.Println("methodName: " + methodName)
	status := executePutCommand(methodName, message.Payload())

	//pubClient,err := newMQTTClient(func(client MQTT.Client){})
	//if err!= nil{
	//	fmt.Printf("Fail to create MQTT client %s\n", err.Error())
	//	return
	//}

	// $iothub/methods/res/{status}/?$rid={request id}
	pubTopic := "$iothub/methods/res/200/?$rid=" + rid
	token := client.Publish(pubTopic, byte(*qos), false, fmt.Sprintf("{\"status\":\"%s\"}", status))
	if token.WaitTimeout(time.Second*time.Duration(10)) && token.Error() != nil {
		fmt.Println("Error publish: " + err.Error())
	} else {
		fmt.Println("Published: " + pubTopic)
	}
}

var server, topic, clientid, username, password *string
var qos *int
var client MQTT.Client

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	server = flag.String("server", "tls://EssaiIOT.azure-devices.net:8883", "The full url of the MQTT server to connect to ex: tcp://127.0.0.1:1883")
	topic = flag.String("topic", "$iothub/methods/POST/#", "Topic to subscribe to")
	qos = flag.Int("qos", 0, "The QoS to subscribe to messages at")
	clientid = flag.String("clientid", "Coriance_Device", "A clientid for the connection")
	username = flag.String("username", "EssaiIOT.azure-devices.net/Coriance_Device/api-version=2019-03-30", "A username to authenticate to the MQTT server")
	password = flag.String("password", "", "Password to match username")
	flag.Parse()

	var err error
	client, err = newMQTTClient(func(client MQTT.Client) {
		token := client.Subscribe(*topic, byte(*qos), onMessageReceived)
		if token.Wait() && token.Error() != nil {
			panic(token.Error())
		} else {
			//fmt.Printf("Subscribe to the %s\n", *topic)
		}
	})
	if err != nil {
		fmt.Printf("Fail to create MQTT client %s\n", err.Error())
		return
	}

	<-c
}

func newMQTTClient(callback func(client MQTT.Client)) (MQTT.Client, error) {
	x509cert := tls.Certificate{}
	x509cert, err := tls.LoadX509KeyPair(
		"/Users/weichou/Downloads/azure/rsa_cert.pem",
		"/Users/weichou/Downloads/azure/rsa_private.pem")
	if err != nil {
		fmt.Println("Failed loading x509 data using pub/private key pair: " + err.Error())
		return nil, err
	}

	connOpts := MQTT.NewClientOptions().AddBroker(*server).SetClientID(*clientid).SetCleanSession(true)
	if *username != "" {
		connOpts.SetUsername(*username)
		if *password != "" {
			connOpts.SetPassword(*password)
		}
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
		Certificates:       []tls.Certificate{x509cert},
	}
	connOpts.OnConnect = func(c MQTT.Client) {
		//fmt.Printf("On connected to %s\n", *server)
		callback(c)
	}
	connOpts.SetTLSConfig(tlsConfig)

	mqttClient := MQTT.NewClient(connOpts)
	token := mqttClient.Connect()
	if token.WaitTimeout(time.Second*time.Duration(30)) && token.Error() != nil {
		return nil, fmt.Errorf("fail to connect broker %v", token.Error())
	} else {
		fmt.Printf("Connected to %s\n", *server)
	}

	return mqttClient, nil
}

func executePutCommand(cmd string, body []byte) string {
	// set the HTTP method, url, and request body
	url := "http://localhost:48082/api/v1/device/name/Coriance_Device/command/" + cmd
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StatusCode)
	return strconv.Itoa(resp.StatusCode)
}
