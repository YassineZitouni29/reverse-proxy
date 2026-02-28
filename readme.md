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

## How to Run the Project
This project implements a Reverse Proxy Load Balancer with multiple backend servers and an admin interface.

### Prerequisites:
Make sure you have installed:
    - Go (version 1.20+ recommended)
    - curl (for backend management via Admin API)
Check installation:
    go version

### Step 1 : Start Backend Servers

Each backend server is started using environment variables:
PORT=<port_number> NAME=<server_name> go run server/server.go

Example (default configuration)
Start three backend servers in separate terminals:

PORT=9000 NAME=A go run server/server.go
PORT=9001 NAME=B go run server/server.go
PORT=9002 NAME=C go run server/server.go


Each server exposes its service at:

| Server | URL                           |
| ------ | ----------------------------- |
| A      | [http://localhost:9000/books] |
| B      | [http://localhost:9001/books] |
| C      | [http://localhost:9002/books] |

### Step 2 : Start Reverse Proxy & Admin Interface

Run:
    go run main.go

This starts:

| Service       | Default Port | URL                             |
| ------------- | ------------ | ------------------------------- |
| Reverse Proxy | 8080         | [http://localhost:8080/books]   |
| Admin Page    | 8081         | [http://localhost:8081/status]  |


⚠️ These ports are the ones in the configuration file. To change them change in the configuration file.

### Step 3 : Test Request Forwarding

You can verify load balancing in two ways:

#### ✅ Option 1 : Browser Requests

Refresh:
    http://localhost:8080/books

Each refresh sends a GET request.

Check the reverse proxy terminal to see which backend handled the request.

#### ✅ Option 2 : Using the Client

Run:
    go run client/client.go

This sends POST requests through the reverse proxy.
The proxy logs will display the selected backend server.

### Step 4 : Add a New Backend Dynamically

Example: add backend running on port 8082.

  Register backend via Admin API:
    curl -X POST http://localhost:8081/backends \
    -H "Content-Type: application/json" \
    -d '{"url": "http://localhost:8082"}'

  ⚠️ At this stage, the backend is registered but not alive (you can see that in the admin page)

  Start the backend server
    PORT=8082 NAME=D go run server/server.go

  You can now verify its status at:
    http://localhost:8081/status


### Step 5 : Remove a Backend

Example: remove backend on port 9001.

    curl -X DELETE http://localhost:8081/backends \
    -H "Content-Type: application/json" \
    -d '{"url":"http://localhost:9001"}'

The backend will immediately disappear from the admin dashboard.



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

