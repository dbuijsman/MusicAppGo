package main

import (
	"general"
	"log"
	"net/http"
	"os"

	"discography/database"
	"discography/handlers"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/go-sql-driver/mysql"
)

const port string = ":9002"
const servername string = "discography"

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	db, err := general.ConnectToMYSQL(logger, servername, "adminMusicApp:admin@tcp(127.0.0.1:3306)/discography")
	if err != nil {
		logger.Printf("Stop starting server")
		return
	}
	defer db.Close()
	broker, closeBroker := general.ConnectToKafka(logger, servername)
	defer closeBroker()
	if topicErr := general.CreateTopics(broker, logger, "newArtist", "newSong"); err != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	producer := broker.Producer(kafka.NewProducerConf())
	getRequest, err := general.DefealtGETRequest(servername)
	if err != nil {
		logger.Fatalf("Can't create a client for sending get requests: %s\n", err)
	}
	logger.Printf("Handler is ready for sending get requests")
	music := handlers.NewMusicHandler(logger, database.NewMusicDB(db), general.GetSendMessage(producer), getRequest)
	general.StartServer(servername, port, initRoutes(music), logger)
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *handlers.MusicHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	getR := router.PathPrefix("/api").Methods(http.MethodGet).Subrouter()
	getR.Use(general.GetAddTokenToContextMiddleware(handler.Logger))
	getR.Use(general.GetOffsetMaxMiddleware(handler.Logger))
	getR.Path("/artists/{firstLetter}").HandlerFunc(handler.ArtistStartingWith)
	getR.Path("/artist/{artist}").HandlerFunc(handler.SongsFromArtist)

	adminR := router.PathPrefix("/admin").Subrouter()
	adminR.Use(general.GetIsAdminMiddleware(handler.Logger))
	adminR.Path("/artist").Methods(http.MethodPost).HandlerFunc(handler.AddArtist)
	adminR.Path("/song").Methods(http.MethodPost).HandlerFunc(handler.AddSongHandler)

	internalR := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internalR.Use(general.GetInternalRequestMiddleware(handler.Logger))
	internalR.Path("/song/{id}").HandlerFunc(handler.FindSongByID)
	return router
}
