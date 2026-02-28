# Reverse-Proxy (Go)

A **custom HTTP reverse proxy written in Go**, designed to explore **load-balancing strategies** and **concurrency**.

This project is built **incrementally**, starting from a minimal working proxy and evolving toward a more **robust and extensible**.

---

## Project Structure

```text
Reverse-Proxy/
│
├── admin/          # Admin API (monitoring & management)
│   └── admin.go
│
├── client/         # Simple HTTP client (testing purposes)
│   └── client.go
│
├── server/         # Reverse proxy HTTP server
│   └── server.go
│
├── serverPool/     # Backend & server pool management & load balancer interface
│   └── serverPool.go
│
├── config.json     # Proxy configuration
│
├── main.go
├── go.mod          # Go module definition
```


## ✅ Project TODO Checklist

### 🔹 Core Architecture
- [x] Initialize Go module (`go mod init`)
- [x] Define core data structures (`Backend`, `ServerPool`)
- [x] Separate concerns (server, pool, admin)
- [x] Implement thread-safe server pool
- [x] Handle case when no backend is available

---

### 🔹 Load Balancing
- [x] Define Round-Robin load-balancing strategy
- [x] Ensure fair rotation using atomic index
- [x] Select only healthy backends
- [x] Implement Least-Connections strategy
- [x] Allow dynamic strategy selection via configuration

---

### 🔹 Reverse Proxy Server
- [x] Implement HTTP reverse proxy handler
- [x] Integrate `httputil.ReverseProxy`
- [x] Forward client requests to selected backend
- [x] Propagate request context to backend
- [x] Increment backend connection counter
- [x] Decrement backend connection counter on request completion

---

### 🔹 Health Monitoring
- [x] Implement background health checker using goroutines
- [x] Use `time.Ticker` for periodic health checks
- [x] Ping backend servers to verify availability
- [x] Update backend alive status safely

---

### 🔹 Admin API
- [x] Run admin API on a separate port
- [x] `GET /status` — show backend health and connection counts
- [x] `POST /backends` — add a new backend dynamically
- [x] `DELETE /backends` — remove an existing backend


---

### 🔹 Configuration & Startup
- [x] Load proxy configuration from `config.json`
- [x] Initialize server pool from configuration
- [x] Start proxy server and admin API concurrently

---

### 🔹 Concurrency & Safety
- [x] Protect shared state using `sync.RWMutex`
- [x] Use `sync/atomic` for connection counters
- [x] Avoid race conditions under concurrent requests

---

### 🔹 Graceful Behavior
- [x] Handle client cancellation using `context.Context`

