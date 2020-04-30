package main

import (
	"bytes"
	"general"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// These configurations will be exported to a file
const port string = ":9919"
const servername string = "gateway"

const portUser string = ":9001"
const portDiscography string = ":9002"
const portLikes string = ":9004"

func main() {
	logger := log.New(os.Stdout, servername, log.LstdFlags|log.Lshortfile)
	client := http.Client{}
	general.StartServer(servername, port, initRoutes(client, logger), logger)
}

// initRoutes will returns a router with the necessary routes registered to it.
func initRoutes(client http.Client, logger *log.Logger) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/signup", redirect(client, logger, portUser))
	router.HandleFunc("/login", redirect(client, logger, portUser))
	router.HandleFunc("/validate/", redirect(client, logger, portUser))
	router.HandleFunc("/api/like", redirect(client, logger, portLikes))
	router.HandleFunc("/api/dislike", redirect(client, logger, portLikes))
	router.HandleFunc("/api/artists/{firstLetter}", redirect(client, logger, portDiscography))
	router.HandleFunc("/api/artist/{artist}", redirect(client, logger, portDiscography))
	router.HandleFunc("/admin/artist", redirect(client, logger, portDiscography))
	router.HandleFunc("/admin/song", redirect(client, logger, portDiscography))
	return router
}

func redirect(client http.Client, logger *log.Logger, portRedirect string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		cookie, cookieErr := request.Cookie("token")
		target := "http://localhost" + portRedirect + request.URL.Path
		if len(request.URL.RawQuery) > 0 {
			target += "?" + request.URL.RawQuery
		}
		log.Printf("redirect to: %s", target)
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
			logger.Printf("Failed to create request: %s\n", err)
			general.SendError(response, http.StatusInternalServerError)
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("Failed to redirect request: %s\n", err)
			general.SendError(response, http.StatusInternalServerError)
		}
		response.WriteHeader(resp.StatusCode)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		response.Write(buf.Bytes())
		logger.Printf("%v: Sending response: %v\n", target, resp.StatusCode)
	}
}
