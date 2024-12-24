package internal

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/exp/slog"
)

type PromToScrapeServer struct {
	client             *APIClient
	conf               *Config
	server             http.Server
	data               string
	lastSuccessfulTime time.Time

	sync.RWMutex
}

func NewPromToScrapeServer(client *APIClient, conf *Config, addr string) *PromToScrapeServer {
	s := &PromToScrapeServer{
		client: client,
		conf:   conf,
		data:   "",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.metricsHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	s.server = http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go s.run()

	return s
}

// metricsHandler is the HTTP handler for the "/metrics" endpoint.
func (s *PromToScrapeServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	s.RLock()
	defer s.RUnlock()
	if time.Since(s.lastSuccessfulTime) < 5*time.Minute {
		_, err := fmt.Fprint(w, s.data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("can't serve metrics", "error", err)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("can't serve metrics", "error", "metrics queried are stale (more than 5 minutes old)")
	}
}

// Run on loop getting the metrics data we need
func (s *PromToScrapeServer) run() string {
	s.queryMetrics()
	// to provide some jitter
	ticker := time.NewTicker(59 * time.Second)

	for {
		select {
		case <-ticker.C:
			s.queryMetrics()
		}
	}
}

// there's an alternate way to implement this:
//
//	keep the objects returned from the query, or convert them into something a bit more ergonomic
//	and create ConstMetrics with the prometheus client. I happened to have the code lying around for working
//	with model.Sample, but the ConstMetrics route is probably more idiomatic and safe.
func (s *PromToScrapeServer) queryMetrics() {
	start := time.Now()
	queriedMetrics, err := QueryMetrics(s.conf, s.client)
	if err != nil {
		slog.Error("failed to query metrics", "error", err)
		return
	}
	s.Lock()
	s.data = SamplesToString(queriedMetrics)
	s.lastSuccessfulTime = time.Now()
	s.Unlock()
	slog.Debug("successful metric retrieval", "time", time.Since(start))
}

// Start runs the embedded http.Server.
func (s *PromToScrapeServer) Start() error {
	slog.Info("listening on", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}
