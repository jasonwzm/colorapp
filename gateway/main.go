package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const defaultPort = "8081"
const maxColors = 1000

var colors [maxColors]string
var colorsIdx int
var colorsMutext = &sync.Mutex{}

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}

	return defaultPort
}

func getColorTellerEndpoint() (string, error) {
	colorTellerEndpoint := os.Getenv("COLOR_TELLER_ENDPOINT")
	if colorTellerEndpoint == "" {
		return "", errors.New("COLOR_TELLER_ENDPOINT is not set")
	}
	return colorTellerEndpoint, nil
}

type colorHandler struct{}

func (h *colorHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	color, err := getColorFromColorTeller(request)
	if err != nil {
		http.Error(writer, "500 - Unexpected Error", http.StatusInternalServerError)
		return
	}

	colorsMutext.Lock()
	defer colorsMutext.Unlock()

	addColor(color)
	stats := getRatios()

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Color: {{.Color}}</title>
		<style>
			.color-square { width: 200px; height: 200px; margin: 20px 0; }
			.stats { font-family: monospace; }
		</style>
	</head>
	<body>
		<div class="color-square" style="background-color: {{.Color}};"></div>
		<h2>Color: {{.Color}}</h2>
		<h3>Stats:</h3>
		<pre class="stats">{{range $color, $ratio := .Stats}}{{$color}}: {{$ratio}}
{{end}}</pre>
	</body>
	</html>`

	t, err := template.New("colorPage").Parse(tmpl)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/html")
	err = t.Execute(writer, colorData{Color: color, Stats: stats})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func addColor(color string) {
	colors[colorsIdx] = color

	colorsIdx += 1
	if colorsIdx >= maxColors {
		colorsIdx = 0
	}
}

func getRatios() map[string]float64 {
	counts := make(map[string]int)
	var total = 0

	for _, c := range colors {
		if c != "" {
			counts[c] += 1
			total += 1
		}
	}

	ratios := make(map[string]float64)
	for k, v := range counts {
		ratio := float64(v) / float64(total)
		ratios[k] = math.Round(ratio*100) / 100
	}

	return ratios
}

type clearColorStatsHandler struct{}

func (h *clearColorStatsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	colorsMutext.Lock()
	defer colorsMutext.Unlock()

	colorsIdx = 0
	for i := range colors {
		colors[i] = ""
	}

	fmt.Fprint(writer, "cleared")
}

func getColorFromColorTeller(request *http.Request) (string, error) {
	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		return "-n/a-", err
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", colorTellerEndpoint), nil)
	if err != nil {
		return "-n/a-", err
	}

	resp, err := client.Do(req.WithContext(request.Context()))
	if err != nil {
		return "-n/a-", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "-n/a-", err
	}

	color := strings.TrimSpace(string(body))
	if len(color) < 1 {
		return "-n/a-", errors.New("Empty response from colorTeller")
	}

	return color, nil
}

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("ping requested, reponding with HTTP 200")
	writer.WriteHeader(http.StatusOK)
}

func main() {
	log.Println("Starting server, listening on port " + getServerPort())

	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Using color-teller at " + colorTellerEndpoint)

	http.Handle("/color", &colorHandler{})
	http.Handle("/color/clear", &clearColorStatsHandler{})
	http.Handle("/ping", &pingHandler{})
	log.Fatal(http.ListenAndServe(":"+getServerPort(), nil))
}

type colorData struct {
	Color string
	Stats map[string]float64
}
