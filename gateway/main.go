package main

import (
	"general"
	"log"
	"net/http"
	"os"

	"github.com/optiopay/kafka/v2"
)

// These configurations will be exported to a file
const servername, port string = "gateway", ":9919"

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	broker, closeBroker := general.ConnectToKafka(logger, servername)
	defer closeBroker()
	if topicErr := general.CreateTopics(broker, logger, "newService"); topicErr != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	sendMessage := general.GetSendMessage(broker.Producer(kafka.NewProducerConf()))
	handler, err := NewGatewayHandler(logger, http.Client{}, sendMessage)
	if err != nil {
		logger.Fatalf("[ERROR] Can't create handler due to: %s\n", err)
	}
	_, startServer := NewGatewayServer(handler, broker, servername, port)
	startServer()
}
