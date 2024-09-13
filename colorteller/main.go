package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"
const defaultColor = "green"

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}

	return defaultPort
}

func getColor() string {
	color := os.Getenv("COLOR")
	if color != "" {
		return color
	}

	return defaultColor
}

type colorHandler struct{}

func (h *colorHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, getColor())
}

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func main() {
	log.Println("starting server, listening on port " + getServerPort())
	http.Handle("/", &colorHandler{})
	http.Handle("/ping", &pingHandler{})
	http.ListenAndServe(":"+getServerPort(), nil)
}
