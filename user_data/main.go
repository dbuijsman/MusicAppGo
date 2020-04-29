package main

import (
	"general"
	"log"
	"net/http"
	"os"
	"user_data/database"
	"user_data/handlers"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/go-sql-driver/mysql"
)

// These configurations will be exported to a file
const port string = ":9111"
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
	if topicErr := general.CreateTopics(broker, logger, "newUser", "login"); err != nil {
		logger.Fatalf("[ERROR] Failed to create topics due to: %s\n", topicErr)
	}
	producer := broker.Producer(kafka.NewProducerConf())
	users := handlers.NewUserHandler(logger, database.NewUserDB(db), general.GetSendMessage(producer))
	general.StartServer(servername, port, initRoutes(users), logger)
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
	validateRouter.Use(general.GetValidateTokenMiddleWare(users.Logger))
	return router
}
