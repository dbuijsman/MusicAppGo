package main

import (
	"general/env"
	"general/server"
	"log"
	"net/http"
	"os"

	"github.com/optiopay/kafka/v2"
)

var servername = env.SetString("SERVER_NAME", false, "gateway")
var serverhost = env.SetString("SERVER_HOST", false, "localhost")
var serverport = env.SetInt("SERVER_PORT", false, 9919)

func main() {
	if err := env.Parse(); err != nil {
		log.Fatalf("Failed to process configurations due to: \n%s\n", err)
	}
	logger := log.New(os.Stdout, *servername, log.LstdFlags|log.Lshortfile)
	broker, closeBroker := server.ConnectToKafka(logger, *servername)
	defer closeBroker()
	if topicErr := server.CreateTopics(broker, logger, "newService"); topicErr != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	sendMessage := server.GetSendMessage(broker.Producer(kafka.NewProducerConf()))
	handler, err := NewGatewayHandler(logger, http.Client{}, sendMessage)
	if err != nil {
		logger.Fatalf("[ERROR] Can't create handler due to: %s\n", err)
	}
	_, startServer := NewGatewayServer(handler, broker, *servername, ":"+string(*serverport))
	startServer()
}
