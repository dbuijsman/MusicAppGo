package main

import (
	"fmt"
	"general/env"
	"general/server"
	"likes/database"
	"likes/handlers"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	_, startServer := handlers.NewLikesServer(handler, broker, *servername, string(*serverport))
	startServer()
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *handlers.LikesHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	clientR := router.PathPrefix("/api").Subrouter()
	clientR.Use(server.GetValidateTokenMiddleWare(handler.Logger))

	getR := clientR.Methods(http.MethodGet).Subrouter()
	getR.Use(server.GetOffsetMaxMiddleware(handler.Logger))
	getR.PathPrefix("/like").HandlerFunc(handler.GetLikes)
	getR.PathPrefix("/dislike").HandlerFunc(handler.GetDislikes)

	likesR := clientR.PathPrefix("/like").Subrouter()
	likesR.Methods(http.MethodPost).HandlerFunc(handler.AddLike)
	likesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveLike)

	dislikesR := clientR.PathPrefix("/dislike").Subrouter()
	dislikesR.Methods(http.MethodPost).HandlerFunc(handler.AddDislike)
	dislikesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveDislike)

	internalR := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internalR.Use(server.GetInternalRequestMiddleware(handler.Logger))
	internalR.Path("/preference/{user}/{artist}").HandlerFunc(handler.GetPreferencesOfArtist)
	return router
}
