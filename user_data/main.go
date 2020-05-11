package main

import (
	"general"
	"log"
	"os"
	"user_data/database"
	"user_data/handlers"

	"github.com/optiopay/kafka/v2"

	_ "github.com/go-sql-driver/mysql"
)

// These configurations will be exported to a file
const port string = ":9001"
const servername string = "users"

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	db, err := general.ConnectToMYSQL(logger, servername, "credentialsMusicApp:validate@tcp(127.0.0.1:3306)/userdata")
	if err != nil {
		logger.Printf("Stop starting server")
		return
	}
	defer db.Close()
	broker, closeBroker := general.ConnectToKafka(logger, servername)
	defer closeBroker()
	if topicErr := general.CreateTopics(broker, logger, "newUser", "login"); topicErr != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	handler, err := handlers.NewUserHandler(logger, database.NewUserDB(db), general.GetSendMessage(broker.Producer(kafka.NewProducerConf())))
	if err != nil {
		logger.Fatalf("[ERROR] Can't create handler due to: %s\n", err)
	}
	_, startServer := handlers.NewUserServer(handler, broker, servername, port)
	startServer()
}
