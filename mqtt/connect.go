package mqtt

import (
	"bufio"
	"container/list"
	b64 "encoding/base64"
	"encoding/json"
	"log"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const MQ_TCP_URL = "127.0.0.1:5687"

type ConnectParams struct {
	HostURL  string
	ClientID string
	Username string
	Password string
}

type MQRequestPayload struct {
	RequestName string      `json:"requestName"`
	Payload     interface{} `json:"payload"`
}

type route struct {
	topic   string
	clients []net.Conn
}

var _connectParams ConnectParams
var mqttClient MQTT.Client

func Connect(connectParams ConnectParams) {
	_connectParams = connectParams
	opts := MQTT.NewClientOptions().AddBroker(connectParams.HostURL)

	if connectParams.ClientID == "" {
		connectParams.ClientID = "mq-" + uuid.New().String()
	}
	opts.SetClientID(connectParams.ClientID)

	if connectParams.Username != "" {
		opts.SetUsername(connectParams.Username)
		opts.SetPassword(connectParams.Password)
	}

	opts.SetConnectionLostHandler(onConnectionLost)
	//opts.SetPingTimeout(60)
	//opts.SetDefaultPublishHandler(mqttMessageHandler)

	mqttClient = MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Error connecting to mqtt broker. ", token.Error())
	}

	log.Println("Connected to " + connectParams.HostURL)

	// captureAllTopics := func(client MQTT.Client, msg MQTT.Message) {
	// 	fmt.Println("Topic #: " + msg.Topic() + ", Message: " + string(msg.Payload()))
	// }

	// captureTopic5 := func(client MQTT.Client, msg MQTT.Message) {
	// 	fmt.Println("Topic 5: " + msg.Topic() + ", Message: " + string(msg.Payload()))
	// }

	// if token := mqttClient.Subscribe("testtopic/#", 1, captureAllTopics); token.Wait() && token.Error() != nil {
	// 	log.Println(token.Error())
	// }

	// if token := mqttClient.Subscribe("testtopic/5", 1, captureTopic5); token.Wait() && token.Error() != nil {
	// 	log.Println(token.Error())
	// }

	startMQServer()
}

func onConnectionLost(client MQTT.Client, err error) {
	log.Println("Connection lost to the broker")
	log.Println(err.Error())
	log.Println("Trying to reconnect")
	Connect(_connectParams)
}

var l *list.List

func startMQServer() {

	l = list.New()

	// Start mq server on default port 5687
	l, err := net.Listen("tcp4", MQ_TCP_URL)
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	netData, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		return
	}

	// Base64 Decode
	sDec, err := b64.StdEncoding.DecodeString(netData)
	if err != nil {
		return
	}

	var mqRequest MQRequestPayload
	if err := json.Unmarshal(sDec, &mqRequest); err != nil {
		return
	}

	if mqRequest.RequestName == "publish" {
		var pubData *PublishData
		mapstructure.Decode(mqRequest.Payload, &pubData)
		handlePublishRequest(c, *pubData)
	} else if mqRequest.RequestName == "subscribe" {
		var subData *SubscribeData
		mapstructure.Decode(mqRequest.Payload, &subData)
		handleSubscribeRequest(c, *subData)
	}
}

func handlePublishRequest(c net.Conn, publishData PublishData) {
	mqttClient.Publish(publishData.Topic, 1, false, publishData.Payload)
	c.Write([]byte("OK\n"))
	c.Close()
}

func messageReceivedHandler(client MQTT.Client, msg MQTT.Message) {
	// log.Println("Receiving message " + msg.Topic())
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(*route).match(msg.Topic()) {
			// fmt.Println("Match found for " + e.Value.(*route).topic + ", clients: " + string(len(e.Value.(*route).clients)))
			for _, c := range e.Value.(*route).clients {
				payload := msg.Payload()
				mqReq := MQRequestPayload{
					RequestName: "messageReceived",
					Payload:     string(payload),
				}
				jsonPayload, _ := json.Marshal(mqReq)
				b64EncodedPayload := b64.StdEncoding.EncodeToString(jsonPayload)
				// log.Println("Sending")
				c.Write([]byte(b64EncodedPayload + "\n"))
			}
		}
	}
}

func handleSubscribeRequest(c net.Conn, subscribeData SubscribeData) {
	for _, topic := range subscribeData.Topics {
		addRoute(topic, c)
		if token := mqttClient.Subscribe(topic, 1, messageReceivedHandler); token.Wait() && token.Error() != nil {
			log.Println(token.Error())
		}
	}

	c.Write([]byte("OK\n"))
}

func addRoute(topic string, client net.Conn) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(*route).match(topic) {
			r := e.Value.(*route)
			r.clients = append(r.clients, client)
			return false
		}
	}
	l.PushBack(&route{topic: topic, clients: []net.Conn{client}})
	return true
}

func match(route []string, topic []string) bool {
	if len(route) == 0 {
		if len(topic) == 0 {
			return true
		}
		return false
	}

	if len(topic) == 0 {
		if route[0] == "#" {
			return true
		}
		return false
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return match(route[1:], topic[1:])
	}
	return false
}

func routeIncludesTopic(route, topic string) bool {
	return match(routeSplit(route), strings.Split(topic, "/"))
}

// removes $share and sharename when splitting the route to allow
// shared subscription routes to correctly match the topic
func routeSplit(route string) []string {
	var result []string
	if strings.HasPrefix(route, "$share") {
		result = strings.Split(route, "/")[2:]
	} else {
		result = strings.Split(route, "/")
	}
	return result
}

// match takes the topic string of the published message and does a basic compare to the
// string of the current Route, if they match it returns true
func (r *route) match(topic string) bool {
	return r.topic == topic || routeIncludesTopic(r.topic, topic)
}
