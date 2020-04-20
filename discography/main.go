package main

import (
	"MusicAppGo/common"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"discography/database"
	"discography/handlers"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/optiopay/kafka/v2/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const port string = ":9002"
const servername string = "discography"

var kafkaAddrs = []string{"localhost:9092", "localhost:9093"}

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	// Opening the database
	db, err := sql.Open("mysql", "adminMusicApp:admin@tcp(127.0.0.1:3306)/discography")
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
	handler := handlers.NewMusicHandler(logger, database.NewMusicDB(db), sendMessage)
	// Create the Server
	server := &http.Server{
		Addr:     port,
		Handler:  initRoutes(handler),
		ErrorLog: logger,
	}

	// start the server
	go func() {
		logger.Printf("Starting server %v on port %v\n", servername, server.Addr)
		err := server.ListenAndServe()
		if err != nil {
			logger.Printf("[ERROR] Closing server %v: %s\n", servername, err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	sig := <-c
	logger.Printf("Closing server %v due to %v signal\n", servername, sig)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(ctx)
	logger.Printf("Closed server %v!\n", servername)
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *handlers.MusicHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	getR := router.Methods(http.MethodGet).Subrouter()
	getR.Use(common.GetOffsetMaxMiddleware(handler.Logger))
	getR.Path("/artists/{firstLetter}").HandlerFunc(handler.ArtistStartingWith)
	getR.Path("/artist/{artist}").HandlerFunc(handler.SongsFromArtist)
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

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)
