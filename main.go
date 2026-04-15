package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type StatusRoot struct {
	Icestats IcecastStats
}

type IcecastStats struct {
	Source Stream
}

type Stream struct {
	Listeners  int
	ServerName string `json:"server_name"`
}

type IcecastClient struct {
	Username   string
	Password   string
	httpClient *http.Client
}

var (
	listeners = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "icecast_listeners",
		Help: "Gauge representing current Icecast stream listeners",
	}, []string{"name", "id"})
)

func (c IcecastClient) LoadIcecastStatus(url string) (stats *StatusRoot, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error get request: %w", err)
	}
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error get response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("icecast returned unexpected status: %s", resp.Status)
	}
	stats = new(StatusRoot)

	json.NewDecoder(resp.Body).Decode(&stats)

	return stats, nil
}

func publishVClock(clock string, listeners int) {
	s := fmt.Sprintf("http://%s/?Command=SetMem=Listeners,%d", clock, listeners)

	resp, err := http.Get(s)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	return
}

func (c IcecastClient) updateListeners(url string, wait int, clock string) {
	go func() {
		for {
			resp, err := c.LoadIcecastStatus(url)

			if err != nil {
				log.Println("Error polling Icecast endpoint, trying again in", wait)
				log.Printf("Error: %v", err)
			} else {
				listeners.WithLabelValues(resp.Icestats.Source.ServerName, "0").Set(float64(resp.Icestats.Source.Listeners))
				go publishVClock(clock, resp.Icestats.Source.Listeners)
			}

			time.Sleep(time.Duration(wait) * time.Second)
		}
	}()
}

func main() {
	urlPtr := flag.String("url", "", "Icecast status endpoint (normally: http://icecast.example.com/status-json.xsl)")
	passwordPtr := flag.String("password", "", "Icecast admin password. Non recomended, use env ICECAST_PASSWORD")
	usernamePtr := flag.String("username", "", "Icecast admin username")
	IcecastConfigPtr := flag.String("icecast-config", "", "Icecast config. For authorization")
	portPtr := flag.Int("port", 2112, "Port to listen on for metrics")
	endpointPtr := flag.String("endpoint", "/metrics", "Metrics endpoint to listen on")
	waitPtr := flag.Int("interval", 15, "Interval to update statistics from Icecast")
	clockPtr := flag.String("clock", "", "VClock URL")

	flag.Parse()

	if *urlPtr == "" {
		log.Fatalf("Missing required argument -url, see '%s -help' for information", os.Args[0])
	}
	if *passwordPtr == "" {
		*passwordPtr = os.Getenv("ICECAST_PASSWORD")
	}
	if *usernamePtr != "" && *passwordPtr == "" {
		log.Println("Warning: Username provided but password is empty. Check ICECAST_PASSWORD env.")
	}

	client := IcecastClient{
		Username: *usernamePtr,
		Password: *passwordPtr,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	if *IcecastConfigPtr != "" {
		icecastConfig, err := ParseIcecastConfig(*IcecastConfigPtr)
		if err != nil {
			panic(err)
		}
		client.Username = icecastConfig.AdminUser
		client.Password = icecastConfig.AdminPassword
	}

	log.Println("Starting Icecast Exporter")

	client.updateListeners(*urlPtr, *waitPtr, *clockPtr)

	http.Handle(*endpointPtr, promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil)
}
