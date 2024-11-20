package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Book type with Name, Author and ISBN
type Book struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	ISBN        string `json:"isbn"`
	Description string `json:"description,omitempty"`
}

// Books slice of all known books
var Books = []Book{
	{Title: "The Hitchhiker's Guide to the Galaxy", Author: "Douglas Adams", ISBN: "0345391802"},
	{Title: "Cloud Native Go", Author: "M.-Leander Reimer", ISBN: "0000000000"},
}

// ToJSON to be used for marshalling of Book type
func (b Book) ToJSON() []byte {
	ToJSON, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return ToJSON
}

// FromJSON to be used for unmarshalling of Book type
func FromJSON(data []byte) Book {
	book := Book{}
	err := json.Unmarshal(data, &book)
	if err != nil {
		panic(err)
	}
	return book
}

// BooksHandleFunc to be used as http.HandleFunc for Book API
func BooksHandleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stderr, "print to stderr\n")
	fmt.Fprintf(os.Stdout, "print to stdout\n")

	b, err := json.Marshal(Books)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}
