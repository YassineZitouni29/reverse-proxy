package serverpool

import (
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL *url.URL `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns atomic.Int64 `json:"current_connections"`
	mux sync.RWMutex
}


func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

type ServerPool struct{
	Backends []*Backend `json:"backends"`
	Current uint64  `json:"current"`
	mu sync.Mutex
}


type LoadBalancer interface {
	GetNextValidPeer(strategy string) *Backend
	AddBackend(backend *Backend)
	SetBackendStatus(uri *url.URL, alive bool)
}


func (sp *ServerPool) GetNextValidPeer(strategy string) *Backend {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	n := len(sp.Backends)
	if n == 0 {
		return nil
	}

	switch strategy {
	case "least-connections", "least_conn", "least":
		var best *Backend
		var bestConn int64
		for _, b := range sp.Backends {
			if !b.Alive {
				continue
			}
			conn := b.CurrentConns.Load()
			if best == nil || conn < bestConn {
				best = b
				bestConn = conn
			}
		}
		return best
	default:
		start := sp.Current
		for t := range n {
			i := (start + uint64(t)) % uint64(n)
			if sp.Backends[i].Alive {
				sp.Current = (i + 1) % uint64(n)
				return sp.Backends[i]
			}
		}
	}
	return nil
}


func (sp *ServerPool) AddBackend(backend *Backend){
	sp.mu.Lock()
	sp.Backends = append(sp.Backends, backend)
	sp.mu.Unlock()
}

func (sp *ServerPool) SetBackendStatus(url *url.URL, alive bool){
	sp.mu.Lock()
	for _,v:= range sp.Backends{
		if v.URL.String()==url.String(){
			v.Alive = alive
			sp.mu.Unlock()
			return
		}
	}
}

func (sp *ServerPool) CheckHealth(){
	var wg sync.WaitGroup
	for _,v:= range sp.Backends{
		wg.Add(1)
		go func(backend *Backend){
			defer wg.Done()
			target := backend.URL.ResolveReference(&url.URL{Path: "/books"})
			client:= http.Client{
				Timeout:2*time.Second,
			}
			resp, err:= client.Get(target.String())
			if err!=nil{
				backend.mux.Lock()
				backend.Alive = false
				backend.mux.Unlock()
				return
			}
			if resp.StatusCode <200 || resp.StatusCode>=400{
				backend.mux.Lock()
				backend.Alive = false
				backend.mux.Unlock()
				return
			}
			backend.mux.Lock()
			backend.Alive=true
			backend.mux.Unlock()
		}(v)
	}
	wg.Wait()
}
