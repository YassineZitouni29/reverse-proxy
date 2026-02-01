package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reverse-proxy/admin"
	"reverse-proxy/serverPool"
	"strings"
	"syscall"
	"time"
)

type ProxyConfig struct {
	Port            string   `json:"proxyPort"`
	AdminPort       string   `json:"adminPort"`
	Strategy        string   `json:"strategy"`
	HealthCheckFreq string   `json:"healthCheckFreq"`
	Urls            []string `json:"urls"`
}

func main() {
	pool := &serverpool.ServerPool{}
	var config ProxyConfig

	file, err := os.ReadFile("config.json")

	if err != nil {
		panic("Cannot decode the urls")
	}

	if err = json.Unmarshal(file, &config); err != nil {
		panic("Cannot decode config file")
	}

	strategy := strings.ToLower(config.Strategy)
	if strategy == "" {
		strategy = "round-robin"
	}

	for _, rawURL := range config.Urls {
		urll, _ := url.Parse(rawURL)
		backend := serverpool.Backend{URL: urll, Alive: false}
		pool.AddBackend(&backend)
	}

	pool.CheckHealth()
	if pool.GetNextValidPeer(strategy) == nil {
		log.Println("no healthy backends yet; requests will return 503 until one is up")
	}

	type backendKey struct{}

	proxy := httputil.ReverseProxy{
		Director: func(r *http.Request) {
			next_backend, _ := r.Context().Value(backendKey{}).(*serverpool.Backend)
			if next_backend == nil {
				next_backend = pool.GetNextValidPeer(strategy)
			}
			if next_backend == nil {
				r.URL = &url.URL{}
				r.Host = ""
				return
			}
			next_backend.CurrentConns.Add(1)
			fmt.Println("forwarded to ", next_backend.URL)

			ctx := context.WithValue(r.Context(), backendKey{}, next_backend)
			*r = *r.WithContext(ctx)

			r.URL.Scheme = next_backend.URL.Scheme
			r.URL.Host = next_backend.URL.Host
			r.Host = next_backend.URL.Host
		},
		ModifyResponse: func(r *http.Response) error {
			if b, ok := r.Request.Context().Value(backendKey{}).(*serverpool.Backend); ok {
				b.CurrentConns.Add(^int64(0))

			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if b, ok := r.Context().Value(backendKey{}).(*serverpool.Backend); ok {
				b.CurrentConns.Add(^int64(0))

				if errors.Is(err, syscall.ECONNREFUSED) {
					b.SetAlive(false)
				}
			}
			log.Printf("proxy error: %v", err)
			http.Error(w, "no backends available", http.StatusServiceUnavailable)
		},
	}

	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		backend := pool.GetNextValidPeer(strategy)
		if backend == nil {
			http.Error(w, "no backends available", http.StatusServiceUnavailable)
			return
		}
		ctx := context.WithValue(r.Context(), backendKey{}, backend)
		proxy.ServeHTTP(w, r.WithContext(ctx))
	})

	healthFreq, err := time.ParseDuration(config.HealthCheckFreq)
	if err != nil || healthFreq <= 0 {
		healthFreq = 5 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(healthFreq)
		defer ticker.Stop()
		for range ticker.C {
			pool.CheckHealth()
		}
	}()

	adminPort := config.AdminPort
	if adminPort == "" {
		adminPort = "8081"
	}
	admin.Start(pool, adminPort)

	port := config.Port
	if port == "" {
		port = "8080"
	}

	log.Println("Reverse proxy listening on :" + port)
	http.ListenAndServe(":"+port, nil)

}
