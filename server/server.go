package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type Book struct {
	Id      int      `json:"id"`
	Title    string   `json:"title"`
	Author     string      `json:"author"`
}

type Server struct {
	mu       sync.Mutex
	nextId   int
	books []Book
	name     string
}

func main() {
	port := getenv("PORT", "9001")
	name := getenv("NAME", "backend-"+port)

	s := &Server{
		nextId:   1,
		books: []Book{},
		name:     name,
	}

	http.HandleFunc("/books", s.BooksHandler)
	log.Println("running", name, "on port", port)
	http.ListenAndServe(":"+port, nil)
}

func (s *Server) BooksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.handleGetBooks(w)
	case http.MethodPost:
		s.handlePostBook(w, r)
	default:
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func (s *Server) handlePostBook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read body", http.StatusBadRequest)
		return
	}

	var toAdd Book
	if err := json.Unmarshal(body, &toAdd); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	toAdd.Id = s.nextId
	s.nextId++
	s.books = append(s.books, toAdd)
	s.mu.Unlock()

	resp, err := json.Marshal(toAdd)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	w.Write(resp)
}

func (s *Server) handleGetBooks(w http.ResponseWriter) {
	s.mu.Lock()
	copyList := make([]Book, len(s.books))
	copy(copyList, s.books)
	s.mu.Unlock()

	resp, err:= json.Marshal(copyList)
	if err!= nil{
		http.Error(w, "invalid", http.StatusBadRequest)
		return 
	}
	w.Write(resp)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
