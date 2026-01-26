package main


import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Book struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

func main() {

	book1 := Book{
		Title:  "The Go Programming Language",
		Author: "Alan A. A. Donovan",
	}

	book2 := Book{
		Title:  "Introduction to Algorithms",
		Author: "Thomas H. Cormen",
	}

	booksToCreate := []Book{book1, book2}

	for _, b := range booksToCreate {

		body, err := json.Marshal(b)
		if err != nil {
			log.Fatal(err)
		}

		_, err = http.Post(
			"http://localhost:8080/books",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Book added")
	}

	/*resp, err := http.Get("http://localhost:8081/books")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var books []Book
	err = json.Unmarshal(data, &books)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nBooks received:")
	for _, b := range books {
		fmt.Println(b.Id, "-", b.Title, "by", b.Author)
	}
		*/
}
