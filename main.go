package main

import (
	"database/sql"
	"errors"
	"net/http"
	_ "strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	_ "api-crud-with-postgres/docs"
)

type Book struct {
	ID          string  `json:"id" example:"1"`
	Title       string  `json:"title" example:"1984"`
	Author      string  `json:"author" example:"George Orwell"`
	Publisher   string  `json:"publisher" example:"Secker & Warburg"`
	Language    string  `json:"language" example:"English"`
	Price       float64 `json:"price" example:"8.99"`
	IsAvailable bool    `json:"is_available" example:"true"`
	Description string  `json:"description" example:"A dystopian novel set in Airstrip One, a province of the superstate Oceania."`
}

type CreateBookRequest struct {
	Title       string  `json:"title" example:"1984"`
	Author      string  `json:"author" example:"George Orwell"`
	Publisher   string  `json:"publisher" example:"Secker & Warburg"`
	Language    string  `json:"language" example:"English"`
	Price       float64 `json:"price" example:"0.0"`
	IsAvailable bool    `json:"is_available" example:"true"`
	Description string  `json:"description"`
}

type UpdateBookRequest struct {
	Title       string  `json:"title" example:"1984"`
	Author      string  `json:"author" example:"George Orwell"`
	Publisher   string  `json:"publisher" example:"Secker & Warburg"`
	Language    string  `json:"language" example:"English"`
	Price       float64 `json:"price" example:"0.0"`
	IsAvailable bool    `json:"is_available" example:"true"`
	Description string  `json:"description"`
}

// @title Book API
// @version 1.0
// @description This is a sample server for managing books.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /

func main() {
	r := gin.Default()
	// Swagger setup
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/books", getBooks)
	r.GET("/books/:id", getBookByID)
	r.POST("/books", createBook)
	r.PUT("/books/:id", updateBook)
	r.DELETE("/books/:id", deleteBook)

	r.Run() // listen and serve on 0.0.0.0:8080
}

// getBooks godoc
// @Summary Get all books
// @Description Get details of all books
// @Tags books
// @Produce json
// @Success 200 {array} Book
// @Router /books [get]
func getBooks(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, author, publisher, language, price, is_available, description FROM books")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Publisher, &book.Language, &book.Price, &book.IsAvailable, &book.Description); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

// getBookByID godoc
// @Summary Get a book by ID
// @Description Get details of a book by its ID
// @Tags books
// @Produce json
// @Param id path string true "Book ID"
// @Router /books/{id} [get]
func getBookByID(c *gin.Context) {
	id := c.Param("id")

	var book Book
	err := db.QueryRow("SELECT id, title, author, publisher, language, price, is_available, description FROM books WHERE id = $1", id).
		Scan(&book.ID, &book.Title, &book.Author, &book.Publisher, &book.Language, &book.Price, &book.IsAvailable, &book.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, book)
}

// createBook godoc
// @Summary Create a new book
// @Description Create a new book with the input payload
// @Tags books
// @Accept json
// @Produce json
// @Param book body CreateBookRequest true "Book to create"
// @Router /books [post]
func createBook(c *gin.Context) {
	var newBookRequest CreateBookRequest
	if err := c.BindJSON(&newBookRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var newBook Book
	err := db.QueryRow(
		"INSERT INTO books (title, author, publisher, language, price, is_available, description) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		newBookRequest.Title, newBookRequest.Author, newBookRequest.Publisher, newBookRequest.Language, newBookRequest.Price, newBookRequest.IsAvailable, newBookRequest.Description).
		Scan(&newBook.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newBook.Title = newBookRequest.Title
	newBook.Author = newBookRequest.Author
	newBook.Publisher = newBookRequest.Publisher
	newBook.Language = newBookRequest.Language
	newBook.Price = newBookRequest.Price
	newBook.IsAvailable = newBookRequest.IsAvailable
	newBook.Description = newBookRequest.Description

	c.JSON(http.StatusCreated, newBook)
}

// updateBook godoc
// @Summary Update a book
// @Description Update details of a book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path string true "Book ID"
// @Param book body UpdateBookRequest true "Updated book"\
// @Router /books/{id} [put]
func updateBook(c *gin.Context) {
	id := c.Param("id")
	var updatedBookRequest UpdateBookRequest
	if err := c.BindJSON(&updatedBookRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec(
		"UPDATE books SET title = $1, author = $2, publisher = $3, language = $4, price = $5, is_available = $6, description = $7 WHERE id = $8",
		updatedBookRequest.Title, updatedBookRequest.Author, updatedBookRequest.Publisher, updatedBookRequest.Language, updatedBookRequest.Price, updatedBookRequest.IsAvailable, updatedBookRequest.Description, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "book updated"})
}

// deleteBook godoc
// @Summary Delete a book
// @Description Delete a book by its ID
// @Tags books
// @Param id path string true "Book ID"
// @Router /books/{id} [delete]
func deleteBook(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "book deleted"})
}
