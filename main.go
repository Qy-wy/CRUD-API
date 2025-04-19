package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Book struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

var storage = make(map[string]Book)
var mu = &sync.Mutex{}

func logError(err error, req *http.Request, message string) {
	logrus.WithFields(logrus.Fields{
		"error":    err.Error(),
		"method":   req.Method,
		"endpoint": req.URL.Path,
	}).Error(message)
}

func returnAllBooks(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mu.Lock()
	defer mu.Unlock()

	var books []Book

	for _, book := range storage {
		books = append(books, book)
	}

	if err := json.NewEncoder(w).Encode(books); err != nil {
		logError(err, req, "Error when encoding JSON")
		return
	}
}

func returnBooksByID(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	bookID := vars["id"]

	mu.Lock()
	book, exist := storage[bookID]
	mu.Unlock()

	if !exist {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(book); err != nil {
		logError(err, req, "Error when encoding JSON")
		return
	}
}

func createBook(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newBook Book
	if err := json.NewDecoder(req.Body).Decode(&newBook); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		logError(err, req, "Error when decoding JSON")
		return
	}

	if _, exists := storage[newBook.ID]; exists {
		http.Error(w, "Record already exists", http.StatusBadRequest)
		return
	}

	mu.Lock()
	storage[newBook.ID] = newBook
	mu.Unlock()

	if err := json.NewEncoder(w).Encode(newBook); err != nil {
		logError(err, req, "Error when encoding JSON")
		return
	}
}

func updateBook(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	bookID := vars["id"]

	var updatedBook Book
	if err := json.NewDecoder(req.Body).Decode(&updatedBook); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		logError(err, req, "Error when decoding JSON")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	_, exist := storage[bookID]
	if !exist {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	storage[bookID] = updatedBook

	if err := json.NewEncoder(w).Encode(updatedBook); err != nil {
		logError(err, req, "Error when encoding JSON")
		return
	}
}

func deleteBook(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	bookID := vars["id"]

	mu.Lock()
	defer mu.Unlock()

	_, exist := storage[bookID]
	if !exist {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	delete(storage, bookID)

	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Book deleted successfully"}); err != nil {
		logError(err, req, "Error when encoding JSON")
		return
	}
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	router := mux.NewRouter()

	router.HandleFunc("/book", returnAllBooks).Methods("GET")
	router.HandleFunc("/book/{id}", returnBooksByID).Methods("GET")
	router.HandleFunc("/book", createBook).Methods("POST")
	router.HandleFunc("/book/{id}", updateBook).Methods("PUT")
	router.HandleFunc("/book/{id}", deleteBook).Methods("DELETE")

	if err := http.ListenAndServe(":8080", router); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("Error when starting the server")
	}
}
