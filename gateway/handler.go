package main

import (
	"bytes"
	"errors"
	"general/convert"
	"general/server"
	"general/types"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
)

// NewGatewayServer returns a new server that will be functioning as a API gateway and a function that starts up the server
func NewGatewayServer(handler *GatewayHandler, broker *kafka.Broker, servername, host string, port int) (newServer *http.Server, start func()) {
	s, channel, startServer := server.NewServer(servername, host, port, initRoutes(handler), broker, nil, handler.logger)
	newServer = s
	start = func() {
		go func() {
			for service := range channel {
				handler.services[service.Name] = service
			}
		}()
		startServer()
	}
	return
}

// initRoutes will returns a router with the necessary routes registered to it.
func initRoutes(handler *GatewayHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/signup", handler.redirect("users"))
	router.HandleFunc("/login", handler.redirect("users"))
	router.HandleFunc("/validate/", handler.redirect("users"))
	router.HandleFunc("/api/like", handler.redirect("likes"))
	router.HandleFunc("/api/dislike", handler.redirect("likes"))
	router.HandleFunc("/api/artists/{firstLetter}", handler.redirect("discography"))
	router.HandleFunc("/api/artist/{artist}", handler.redirect("discography"))
	router.HandleFunc("/admin/artist", handler.redirect("discography"))
	router.HandleFunc("/admin/song", handler.redirect("discography"))

	internRouter := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internRouter.Use(server.GetInternalRequestMiddleware(handler.logger))
	internRouter.HandleFunc("/service", handler.getServices)

	return router
}

// GatewayHandler is the handler that will be used for redirecting requests to the right service.
// It will also be used for getting the addresses of the differenct services for internal requests.
type GatewayHandler struct {
	logger      *log.Logger
	client      http.Client
	sendMessage func(string, []byte) error
	services    map[string]types.Service
}

// NewGatewayHandler returns a GatewayHandler with the given data.
// It returns an error if sendMessage is nil
func NewGatewayHandler(logger *log.Logger, client http.Client, sendMessage func(string, []byte) error) (*GatewayHandler, error) {
	if sendMessage == nil {
		return nil, errors.New("sendMessage can't be nil")
	}
	return &GatewayHandler{logger: logger, client: client, sendMessage: sendMessage, services: make(map[string]types.Service)}, nil
}
func (handler *GatewayHandler) redirect(serviceName string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		service, ok := handler.services[serviceName]
		if !ok {
			handler.logger.Printf("Failed to redirect request due to missing service: %v\n", serviceName)
			server.SendError(response, http.StatusInternalServerError)
			return
		}
		cookie, cookieErr := request.Cookie("token")
		target := service.Address + request.URL.Path
		if len(request.URL.RawQuery) > 0 {
			target += "?" + request.URL.RawQuery
		}
		log.Printf("Redirect request to: %s", target)
		body := new(bytes.Buffer)
		body.ReadFrom(request.Body)
		bodyRequest, writer := io.Pipe()
		go func() {
			writer.Write(body.Bytes())
			writer.Close()
		}()
		req, err := http.NewRequest(request.Method, target, bodyRequest)
		if cookieErr == nil {
			req.Header.Add("Token", cookie.Value)
		}
		if err != nil {
			handler.logger.Printf("Failed to create request: %s\n", err)
			server.SendError(response, http.StatusInternalServerError)
			return
		}
		resp, err := handler.client.Do(req)
		if err != nil {
			handler.logger.Printf("Failed to redirect request: %s\n", err)
			server.SendError(response, http.StatusInternalServerError)
			return
		}
		response.WriteHeader(resp.StatusCode)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		response.Write(buf.Bytes())
		handler.logger.Printf("%v: Sending response: %v\n", target, resp.StatusCode)
	}
}

func (handler *GatewayHandler) getServices(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := convert.WriteToJSON(&handler.services, response)
	if err != nil {
		handler.logger.Printf("[ERROR] %s\n", err)
	}
}
