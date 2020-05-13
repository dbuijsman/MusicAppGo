package main

import (
	"fmt"
	"general/env"
	"general/server"
	"likes/database"
	"likes/handlers"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var servername = env.SetString("SERVER_NAME", false, "likes")
var serverhost = env.SetString("SERVER_HOST", false, "localhost")
var serverport = env.SetInt("SERVER_PORT", false, 9004)
var dbName = env.SetString("DB_NAME", true, "")
var dbUsername = env.SetString("DB_USERNAME", true, "")
var dbPass = env.SetString("DB_PASSWORD", true, "")

func main() {
	if err := env.Parse(); err != nil {
		log.Fatalf("Failed to process configurations due to: \n%s\n", err)
	}
	logger := log.New(os.Stdout, *servername, log.LstdFlags|log.Lshortfile)
	db, err := server.ConnectToMYSQL(logger, *servername, fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v", *dbUsername, *dbPass, *dbName))
	if err != nil {
		logger.Printf("Stop starting server")
		return
	}
	defer db.Close()
	broker, closeBroker := server.ConnectToKafka(logger, *servername)
	defer closeBroker()
	logger.Printf("Handler is ready for sending get requests")
	handler := handlers.NewLikesHandler(logger, database.NewLikesDB(db), nil)
	_, startServer := handlers.NewLikesServer(handler, broker, *servername, *serverhost, *serverport)
	startServer()
}
