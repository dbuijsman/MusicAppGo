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

var servername = env.SetString("SERVER_NAME", false, "likes", "Name of the likes service")
var serverhost = env.SetString("SERVER_HOST", false, "localhost", "Host of the likes service")
var serverport = env.SetInt("SERVER_PORT", false, 9004, "Port of the likes service")
var dbName = env.SetString("DB_NAME", true, "", "Name of the database for the likes service")
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
	logger.Printf("Handler is ready for sending get requests")
	handler := handlers.NewLikesHandler(logger, database.NewLikesDB(db), nil)
	_, startServer := handlers.NewLikesServer(handler, broker, *servername, *serverhost, *serverport)
	startServer()
}
