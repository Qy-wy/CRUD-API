package main

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Book struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

type BookService struct {
	Storage map[string]Book
	Mu      *sync.RWMutex
	Logger  *logrus.Logger
}

func (bs *BookService) logError(err error, c *gin.Context, message string) {
	bs.Logger.WithFields(logrus.Fields{
		"error":    err.Error(),
		"method":   c.Request.Method,
		"endpoint": c.FullPath(),
	}).Error(message)
}

func NoExist(exist bool, c *gin.Context) bool {
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return true
	}

	return false
}

func (bs *BookService) returnAllBooks(c *gin.Context) {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()

	var books []Book

	for _, book := range bs.Storage {
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

func (bs *BookService) returnBooksByID(c *gin.Context) {

	bookID := c.Param("id")

	bs.Mu.RLock()
	book, exist := bs.Storage[bookID]
	bs.Mu.RUnlock()

	if NoExist(exist, c) {
		return
	}

	c.JSON(http.StatusOK, book)
}

func (bs *BookService) createBook(c *gin.Context) {

	var newBook Book
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		bs.logError(err, c, "Error when decoding JSON")
		return
	}

	if _, exists := bs.Storage[newBook.ID]; exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record already exists"})
		return
	}

	bs.Mu.Lock()
	bs.Storage[newBook.ID] = newBook
	bs.Mu.Unlock()

	c.JSON(http.StatusOK, newBook)
}

func (bs *BookService) updateBook(c *gin.Context) {

	bookID := c.Param("id")

	var updatedBook Book
	if err := c.ShouldBindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		bs.logError(err, c, "Error when decoding JSON")
		return
	}

	bs.Mu.Lock()
	defer bs.Mu.Unlock()

	_, exist := bs.Storage[bookID]

	if NoExist(exist, c) {
		return
	}

	bs.Storage[bookID] = updatedBook

	c.JSON(http.StatusOK, updatedBook)
}

func (bs *BookService) deleteBook(c *gin.Context) {
	bookID := c.Param("id")

	bs.Mu.Lock()
	defer bs.Mu.Unlock()

	_, exist := bs.Storage[bookID]

	if NoExist(exist, c) {
		return
	}

	delete(bs.Storage, bookID)

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

func main() {

	bs := &BookService{
		Storage: make(map[string]Book),
		Mu:      &sync.RWMutex{},
		Logger:  logrus.New(),
	}

	bs.Logger.SetFormatter(&logrus.JSONFormatter{})

	router := gin.Default()

	router.GET("/book", bs.returnAllBooks)
	router.GET("/book/:id", bs.returnBooksByID)
	router.POST("/book", bs.createBook)
	router.PUT("/book/:id", bs.updateBook)
	router.DELETE("/book/:id", bs.deleteBook)

	if err := http.ListenAndServe(":8080", router); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("Error when starting the server")
	}
}
