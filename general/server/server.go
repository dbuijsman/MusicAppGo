package server

import (
	"context"
	"database/sql"
	"fmt"
	"general/convert"
	"general/types"
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
var gateway, addressGateway = "gateway", ":9919"

// ConnectToMYSQL connects
func ConnectToMYSQL(logger *log.Logger, servername, dataSourceName string) (*sql.DB, error) {
	// Opening the database
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		logger.Fatalf("[ERROR] Failed to open connection to %v database: %v\n", servername, err.Error())
		return nil, err
	}
	if err = db.Ping(); err != nil {
		logger.Fatalf("[ERROR] Failed to open connection to %v database: %v\n", servername, err.Error())
		return nil, err
	}
	return db, nil
}

// NewServer returns a server on the given port with the given router, a channel that sends addresses of other services and a start function in order to start the server
func NewServer(servername, port string, router *mux.Router, broker *kafka.Broker, messageConsumer func(), logger *log.Logger) (server *http.Server, channelNewService chan types.Service, start func()) {
	server = &http.Server{
		Addr:     port,
		Handler:  router,
		ErrorLog: logger,
	}
	channelNewService = make(chan types.Service)
	start = func() {
		go getAddressesServices(logger, channelNewService, servername)
		go registerService(logger, broker, servername, port)
		go StartConsumer(broker, logger, "newService", getConsumeNewService(logger, channelNewService))
		if messageConsumer != nil {
			go messageConsumer()
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		logger.Printf("Server %v is shut down!\n", servername)
	}
	return
}

func getAddressesServices(logger *log.Logger, channel chan<- types.Service, servername string) {
	if servername == gateway {
		return
	}
	getRequest, err := GetInternalGETRequest(servername)
	if err != nil {
		logger.Printf("[ERROR] Can't create a get request due to: %s\n", err)
		return
	}
	response, err := getRequest(fmt.Sprintf("http://localhost%v/intern/service", addressGateway))
	if err != nil {
		logger.Printf("[ERROR] Failed to retrieve other services due to: %s\n", err)
		return
	}
	if response.StatusCode != 200 {
		logger.Printf("Failed to obtain list of services due to failed request with errorcode: %v\n", response.StatusCode)
		return
	}
	var services map[string]types.Service
	if err := convert.ReadFromJSONNoValidation(&services, response.Body); err != nil {
		logger.Printf("[ERROR] Failed to decode response of getting a list of services: %s\n", err)
		return
	}
	for _, service := range services {
		channel <- service
	}
	logger.Printf("Obtained all addresses of services\n")
}

func registerService(logger *log.Logger, broker *kafka.Broker, servername, port string) {
	messageService, err := convert.ToJSONBytes(&types.Service{Name: servername, Address: port})
	if err != nil {
		logger.Fatalf("Can't register service %v due to: %s\n", servername, err)
	}
	sendMessage := GetSendMessage(broker.Producer(kafka.NewProducerConf()))
	sendMessage("newService", messageService)
}

func getConsumeNewService(logger *log.Logger, channel chan<- types.Service) func(message []byte) {
	return func(message []byte) {
		var newService types.Service
		if err := convert.FromJSONBytes(&newService, message); err != nil {
			logger.Printf("Failed to deserialize message: %v due to: %s\n", string(message), err)
			return
		}
		logger.Printf("Received address for service %v: %v\n", newService.Name, newService.Address)
		channel <- newService
	}
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

// GetInternalGETRequest returns a function that sends a get request with an internal token to the given url and returns the response
func GetInternalGETRequest(servername string) (func(url string) (*http.Response, error), error) {
	tokenString, errToken := CreateTokenInternalRequests(servername)
	token := &tokenString
	client := http.Client{}
	if errToken != nil {
		return nil, errToken
	}
	return func(url string) (*http.Response, error) {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Token", *token)
		resp, respError := client.Do(request)
		if resp.StatusCode == http.StatusUnauthorized {
			*token, errToken = CreateTokenInternalRequests(servername)
			request, err = http.NewRequest("GET", url, nil)
			request.Header.Add("Token", *token)
			return client.Do(request)
		}
		return resp, respError
	}, nil
}

// StartConsumer consumes messages from the given topic and calls the given function
func StartConsumer(broker *kafka.Broker, logger *log.Logger, topic string, processMessage func([]byte)) {
	conf := kafka.NewConsumerConf(topic, 0)
	conf.StartOffset = kafka.StartOffsetNewest
	consumer, err := broker.Consumer(conf)
	if err != nil {
		logger.Printf("[Warning] Cannot create kafka consumer for %v:%s\nTrying to create topic...\n", topic, err)
		if topicErr := CreateTopics(broker, logger, topic); topicErr != nil {
			logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
			return
		}
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
