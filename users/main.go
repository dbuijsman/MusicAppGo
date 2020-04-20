package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"users/database"
	"users/handlers"

	"MusicAppGo/common"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/optiopay/kafka/v2/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// These configurations will be exported to a file
const port string = ":9001"
const servername string = "users"

var kafkaAddrs = []string{"localhost:9092", "localhost:9093"}

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	db, err := sql.Open("mysql", "credentialsMusicApp:validate@tcp(127.0.0.1:3306)/userdata")
	if err != nil {
		logger.Fatalf("[ERROR] Failed to open connection to %v database: %v\n", servername, err.Error())
		return
	}
	if err = db.Ping(); err != nil {
		logger.Fatalf("[ERROR] Failed to open connection to %v database: %v\n", servername, err.Error())
		return
	}
	defer db.Close()
	producer, closeBroker := connectToKafka()
	defer closeBroker()
	sendMessage := func(topic string, message []byte) error {
		msg := &proto.Message{Value: []byte(message)}
		_, err := producer.Produce(topic, 0, msg)
		return err
	}
	users := handlers.NewUserHandler(logger, database.NewUserDB(db), sendMessage)
	server := &http.Server{
		Addr:     port,
		Handler:  initRoutes(users),
		ErrorLog: logger,
	}
	go func() {
		logger.Printf("Starting server %v on port %v\n", servername, server.Addr)
		err := server.ListenAndServe()
		if err != nil {
			logger.Printf("Closing server %v: %s\n", servername, err)
			os.Exit(1)
		}
	}()

	// Trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	sig := <-c
	logger.Printf("Closing server %v due to %v signal\n", servername, sig)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(ctx)
	logger.Printf("Closed server %v!\n", servername)
}

// initRoutes will returns a router with the necessary routes registered to it.
func initRoutes(users *handlers.UserHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/signup", users.SignUp)
	postRouter.HandleFunc("/login", users.Login)

	validateRouter := router.PathPrefix("/validate").Subrouter()
	validateRouter.HandleFunc("/", users.GetRole)
	validateRouter.Use(common.GetValidateTokenMiddleWare(users.Logger))
	return router
}

func connectToKafka() (producer kafka.Producer, closeBroker func()) {
	conf := kafka.NewBrokerConf(servername)
	conf.AllowTopicCreation = true

	// connect to kafka cluster
	broker, err := kafka.Dial(kafkaAddrs, conf)
	if err != nil {
		log.Fatalf("cannot connect to kafka cluster: %s", err)
	}
	producer = broker.Producer(kafka.NewProducerConf())
	closeBroker = broker.Close
	return
}
