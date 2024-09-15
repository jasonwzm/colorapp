package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/exp/rand"
)

const defaultPort = "8080"
const defaultColor = "green"
const defaultGroup = "control"

func getStatsigServerSDKKey() string {
	key := os.Getenv("STATSIG_SERVER_SDK_KEY")
	if key == "" {
		log.Println("Error: STATSIG_SERVER_SDK_KEY not set")
		return ""
	}
	return key
}

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

func getGroup() string {
	group := os.Getenv("GROUP")
	if group != "" {
		return group
	}

	return defaultGroup
}

type colorHandler struct{}

func (h *colorHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	sdkKey := getStatsigServerSDKKey()
	if sdkKey != "" {
		client := &http.Client{}
		uid := fmt.Sprintf("%s-%s-%s-%s-%s",
			fmt.Sprintf("%x", rand.Uint32()),
			fmt.Sprintf("%x", rand.Uint32()),
			fmt.Sprintf("%x", rand.Uint32()),
			fmt.Sprintf("%x", rand.Uint32()),
			fmt.Sprintf("%x", rand.Uint32()),
		)
		exposureData := fmt.Sprintf(`{"exposures": [{"user": {"customIDs": {"requestID": "%s"}}, "experimentName": "colorapp", "group": "%s"}]}`, uid, getGroup())
		req, err := http.NewRequest("POST", "https://events.statsigapi.net/v1/log_custom_exposure", strings.NewReader(exposureData))
		if err != nil {
			log.Printf("Error creating request: %v", err)
		} else {
			req.Header.Set("statsig-api-key", sdkKey)
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error sending exposure event: %v", err)
			} else {
				defer resp.Body.Close()
			}
		}
	}
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
