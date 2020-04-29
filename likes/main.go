package main

import (
	"general"
	"likes/database"
	"likes/handlers"
	"log"
	"net/http"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const port string = ":9033"
const servername string = "likes"

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	// Opening the database
	db, err := sql.Open("mysql", "likesMusicApp:likelikes@tcp(127.0.0.1:3306)/pref_likes")
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
	GETRequest, err := handlers.DefealtGETRequest(servername)
	if err != nil {
		logger.Fatalf("Can't create a client for sending get requests: %s\n", err)
	}
	logger.Printf("Handler is ready for sending get requests")
	handler := handlers.NewLikesHandler(logger, database.NewLikesDB(db), GETRequest)
	go handler.StartConsuming(broker)
	general.StartServer(servername, port, initRoutes(handler), logger)
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *handlers.LikesHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	clientR := router.PathPrefix("/api").Subrouter()
	clientR.Use(general.GetValidateTokenMiddleWare(handler.Logger))

	likesR := clientR.PathPrefix("/like").Subrouter()
	likesR.Methods(http.MethodPost).HandlerFunc(handler.AddLike)
	likesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveLike)

	dislikesR := clientR.PathPrefix("/like").Subrouter()
	dislikesR.Methods(http.MethodPost).HandlerFunc(handler.AddDislike)
	dislikesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveDislike)

	return router
}
