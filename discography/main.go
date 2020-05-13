package main

import (
	"fmt"
	"general/env"
	"general/server"
	"log"
	"os"

	"discography/database"
	"discography/handlers"

	"github.com/optiopay/kafka/v2"

	_ "github.com/go-sql-driver/mysql"
)

var servername = env.SetString("SERVER_NAME", false, "discography", "Name of the discography service")
var serverhost = env.SetString("SERVER_HOST", false, "localhost", "Host of the discography service")
var serverport = env.SetInt("SERVER_PORT", false, 9001, "Port of the discography service")
var dbName = env.SetString("DB_NAME", true, "", "Name of the database for the discography service")
var dbUsername = env.SetString("DB_USERNAME", true, "", "Username for connecting with the database")
var dbPass = env.SetString("DB_PASSWORD", true, "", "Password for connecting with the database")

func main() {
	env.Parse()
	logger := log.New(os.Stdout, *servername, log.LstdFlags|log.Lshortfile)
	db, err := server.ConnectToMYSQL(logger, *servername, fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v", *dbUsername, *dbPass, *dbName))
	if err != nil {
		logger.Printf("Stop starting server")
		return
	}
	defer db.Close()
	broker, closeBroker := server.ConnectToKafka(logger, *servername)
	defer closeBroker()
	if topicErr := server.CreateTopics(broker, logger, "newArtist", "newSong"); topicErr != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	logger.Printf("Handler is ready for sending get requests")
	producer := broker.Producer(kafka.NewProducerConf())
	handler, err := handlers.NewMusicHandler(logger, database.NewMusicDB(db), server.GetSendMessage(producer), nil)
	if err != nil {
		logger.Fatalf("[ERROR] Can't create handler due to: %s\n", err)
	}
	_, startServer := handlers.NewMusicServer(handler, broker, *servername, *serverhost, *serverport)
	startServer()
}
