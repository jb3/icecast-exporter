package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"time"
)

type StatusRoot struct {
	Icestats IcecastStats
}

type IcecastStats struct {
	Source []Stream
}

type Stream struct {
	Listeners  int
	ServerName string `json:"server_name"`
}

var (
	listeners = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "icecast_listeners",
		Help: "Gauge representing current Icecast stream listeners",
	}, []string{"name", "id"})
)

func LoadIcecastStatus(url string) (stats *StatusRoot, err error) {
	resp, err := http.Get(url)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	stats = new(StatusRoot)

	json.NewDecoder(resp.Body).Decode(&stats)

	return
}

func updateListeners(url string, wait int) {
	go func() {
		for {
			resp, err := LoadIcecastStatus(url)

			if err != nil {
				log.Println("Error polling Icecast endpoint, trying again in", wait)
				return
			}

			for index, element := range resp.Icestats.Source {
				listeners.WithLabelValues(element.ServerName, fmt.Sprint(index)).Set(float64(element.Listeners))
			}

			time.Sleep(15 * time.Second)
		}
	}()
}

func main() {
	urlPtr := flag.String("url", "", "Icecast status endpoint (normally: http://icecast.example.com/status-json.xsl)")
	portPtr := flag.Int("port", 2112, "Port to listen on for metrics")
	endpointPtr := flag.String("endpoint", "/metrics", "Metrics endpoint to listen on")
	waitPtr := flag.Int("interval", 15, "Interval to update statistics from Icecast")

	flag.Parse()

	if *urlPtr == "" {
		log.Fatalf("Missing required argument -url, see '%s -help' for information", os.Args[0])
	}

	log.Println("Starting Icecast Exporter...")

	updateListeners(*urlPtr, *waitPtr)

	http.Handle(*endpointPtr, promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil)
}
