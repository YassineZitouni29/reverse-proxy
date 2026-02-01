package admin

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"reverse-proxy/serverPool"
)

func Start(pool *serverpool.ServerPool, port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := pool.Snapshot()
		total := len(status)
		active := 0
		for _, b := range status {
			if b.Alive {
				active++
			}
		}
		resp := struct {
			TotalBackends  int                        `json:"total_backends"`
			ActiveBackends int                        `json:"active_backends"`
			Backends       []serverpool.BackendStatus `json:"backends"`
		}{
			TotalBackends:  total,
			ActiveBackends: active,
			Backends:       status,
		}
		w.Header().Set("Content-Type", "application/json")
		body, _ := json.MarshalIndent(resp, "", "  ")
		w.Write(body)
	})

	type backendInput struct {
		URL string `json:"url"`
	}

	mux.HandleFunc("/backends", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var input backendInput
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &input); err != nil || input.URL == "" {
				http.Error(w, "invalid backend input", http.StatusBadRequest)
				return
			}
			u, err := url.Parse(input.URL)
			if err != nil || u.Host == "" || u.Scheme == "" {
				http.Error(w, "invalid url", http.StatusBadRequest)
				return
			}
			pool.AddBackend(&serverpool.Backend{URL: u})
			go pool.CheckHealth()
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			var input backendInput
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &input); err != nil || input.URL == "" {
				http.Error(w, "invalid backend input", http.StatusBadRequest)
				return
			}
			u, err := url.Parse(input.URL)
			if err != nil || u.Host == "" || u.Scheme == "" {
				http.Error(w, "invalid url", http.StatusBadRequest)
				return
			}
			if removed := pool.RemoveBackend(u); !removed {
				http.Error(w, "backend not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("Admin server listening on :" + port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("admin server error: %v", err)
		}
	}()
}
