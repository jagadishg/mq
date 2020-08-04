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

type PublishData struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func Publish(publishData PublishData) {
	// Connet to mq server
	conn, err := net.Dial("tcp", MQ_TCP_URL)
	if err != nil {
		log.Println("Error connecting to mqtt server. Make sure you have connected to the broker via `mq connect` command")
		log.Fatalln(err)
	}

	// Publish message to mq server first which will publish to mqtt server
	payload := MQRequestPayload{
		RequestName: "publish",
		Payload:     publishData,
	}
	jsonPayload, _ := json.Marshal(payload)

	// Base64 encode and send
	b64EncodedPayload := b64.StdEncoding.EncodeToString(jsonPayload)
	fmt.Fprintf(conn, b64EncodedPayload+"\n")

	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print(message)
	if message != "OK\n" {
		log.Println("Unknown error occurred publishing message")
		os.Exit(1)
	}

	conn.Close()
}
