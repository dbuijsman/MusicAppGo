package main

import (
	"general"
	"log"
	"os"

	"discography/database"
	"discography/handlers"

	"github.com/optiopay/kafka/v2"

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
	if topicErr := general.CreateTopics(broker, logger, "newArtist", "newSong"); topicErr != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	logger.Printf("Handler is ready for sending get requests")
	producer := broker.Producer(kafka.NewProducerConf())
	handler, err := handlers.NewMusicHandler(logger, database.NewMusicDB(db), general.GetSendMessage(producer), nil)
	if err != nil {
		logger.Fatalf("[ERROR] Can't create handler due to: %s\n", err)
	}
	_, startServer := handlers.NewMusicServer(handler, broker, servername, port)
	startServer()
}
