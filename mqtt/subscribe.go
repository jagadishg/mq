package mqtt

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type SubscribeData struct {
	Topics []string `json:"topics"`
}

func Subscribe(subscribeData SubscribeData) {
	// Connet to mq server
	conn, err := net.Dial("tcp", MQ_TCP_URL)
	if err != nil {
		log.Println("Error connecting to mqtt server. Make sure you have connected to the broker via `mq connect` command")
		log.Fatalln(err)
	}

	// Publish message to mq server first which will publish to mqtt server
	payload := MQRequestPayload{
		RequestName: "subscribe",
		Payload:     subscribeData,
	}
	jsonPayload, _ := json.Marshal(payload)

	// Base64 encode and send
	b64EncodedPayload := b64.StdEncoding.EncodeToString(jsonPayload)
	fmt.Fprintf(conn, b64EncodedPayload+"\n")

	message, _ := bufio.NewReader(conn).ReadString('\n')
	if message != "OK\n" {
		log.Println("Unknown error occurred publishing message. Received " + message)
		os.Exit(1)
	}

	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		// Base64 Decode
		sDec, err := b64.StdEncoding.DecodeString(netData)
		if err != nil {
			log.Println(err)
			return
		}

		var mqRequest MQRequestPayload
		if err := json.Unmarshal(sDec, &mqRequest); err != nil {
			log.Println(err)
			return
		}

		if mqRequest.RequestName == "messageReceived" {
			msg := mqRequest.Payload.(string)
			log.Println(msg)
		}
	}
}
