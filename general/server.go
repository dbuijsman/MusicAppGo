package general

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/optiopay/kafka/v2/proto"
)

var kafkaAddrs = []string{"localhost:9092", "localhost:9093"}

// StartServer starts up a server with the given configuration. The server will gracefully shut down after entering crtl+C
func StartServer(servername, port string, router *mux.Router, logger *log.Logger) {
	server := &http.Server{
		Addr:     port,
		Handler:  router,
		ErrorLog: logger,
	}
	go func() {
		logger.Printf("Starting server %v on port %v\n", servername, server.Addr)
		err := server.ListenAndServe()
		if err != nil {
			logger.Printf("Shutting down server %v: %s\n", servername, err)
			return
		}
	}()
	// Trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	sig := <-c
	logger.Printf("Shutting down server %v due to %v signal\n", servername, sig)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(ctx)
	logger.Printf("Server %v is shut down!\n", servername)
}

// ConnectToKafka creates a connection to Kafka and returns a broker and a closing function
func ConnectToKafka(logger *log.Logger, servername string) (*kafka.Broker, func()) {
	conf := kafka.NewBrokerConf(servername)
	conf.AllowTopicCreation = true
	broker, err := kafka.Dial(kafkaAddrs, conf)
	if err != nil {
		logger.Fatalf("[ERROR] Can't connect to kafka cluster: %s", err)
	}
	logger.Printf("Connected %v to Kafka\n", servername)
	return broker, broker.Close
}

// CreateTopics will create all given topics
func CreateTopics(broker *kafka.Broker, logger *log.Logger, topics ...string) error {
	listTopics := make([]proto.TopicInfo, 0, len(topics))
	for _, topic := range topics {
		listTopics = append(listTopics, proto.TopicInfo{Topic: topic, NumPartitions: 1, ReplicationFactor: 1})
	}
	response, err := broker.CreateTopic(listTopics, 10*time.Second, true)
	if err != nil {
		logger.Printf("[ERROR] Can't create topics %v due to:%s\n", topics, err)
		return err
	}
	for _, topicError := range response.TopicErrors {
		if topicError.Err != nil && topicError.ErrorCode != 36 { // ErrorCode for duplicate topic. This avoids errors when restarting the service.
			logger.Printf("[ERROR] Can't create topic %v due to:%s\n", topicError.Topic, topicError.Err)
			return topicError.Err
		}
		logger.Printf("Topic %v is now available for messaging\n", topicError.Topic)
	}
	return nil
}

// GetSendMessage returns a function that can be used for sending messages to kafka
func GetSendMessage(producer kafka.Producer) func(topic string, message []byte) error {
	return func(topic string, message []byte) error {
		msg := &proto.Message{Value: []byte(message)}
		_, err := producer.Produce(topic, 0, msg)
		return err
	}
}

// StartConsumer consumes messages from the given topic and calls the given function
func StartConsumer(broker kafka.Client, logger *log.Logger, topic string, processMessage func([]byte)) {
	conf := kafka.NewConsumerConf(topic, 0)
	conf.StartOffset = kafka.StartOffsetNewest
	consumer, err := broker.Consumer(conf)
	if err != nil {
		logger.Printf("[ERROR] Cannot create kafka consumer for %v:%s", topic, err)
		return
	}
	logger.Printf("Starting to consume messages from topic %v\n", topic)
	for {
		msg, err := consumer.Consume()
		if err != nil {
			if err != kafka.ErrNoData {
				logger.Printf("Cannot consume %v topic message: %s", topic, err)
			}
			break
		}
		processMessage(msg.Value)
	}
	logger.Printf("Consumer %v quit!", topic)
}

// SendError sends an error message corresponding to the errorcode to the response. It does not end the request
func SendError(response http.ResponseWriter, errorcode int) {
	http.Error(response, http.StatusText(errorcode), errorcode)
}
