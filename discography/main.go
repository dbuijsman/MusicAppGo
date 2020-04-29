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

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const port string = ":9002"
const servername string = "discography"

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
	broker, closeBroker := general.ConnectToKafka(logger, servername)
	defer closeBroker()
	if topicErr := general.CreateTopics(broker, logger, "newArtist", "newSong"); err != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	producer := broker.Producer(kafka.NewProducerConf())
	music := handlers.NewMusicHandler(logger, database.NewMusicDB(db), general.GetSendMessage(producer))
	general.StartServer(servername, port, initRoutes(music), logger)
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *handlers.MusicHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	getR := router.PathPrefix("/api").Methods(http.MethodGet).Subrouter()
	getR.Use(general.GetOffsetMaxMiddleware(handler.Logger))
	getR.Path("/artists/{firstLetter}").HandlerFunc(handler.ArtistStartingWith)
	getR.Path("/artist/{artist}").HandlerFunc(handler.SongsFromArtist)

	adminR := router.PathPrefix("/admin").Subrouter()
	adminR.Use(general.GetIsAdminMiddleware(handler.Logger))
	adminR.Path("/artist").Methods(http.MethodPost).HandlerFunc(handler.AddArtist)

	internalR := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internalR.Use(general.GetInternalRequestMiddleware(handler.Logger))
	internalR.Path("/song/{id}").HandlerFunc(handler.FindSongByID)
	return router
}
