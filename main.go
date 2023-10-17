package main

import (
	"database/sql"
	"net/http"
	"simple-web-app-with-db/config"

	"github.com/gin-gonic/gin"
)

type Book struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	ReleaseYear string `json:"releaseYear"`
	Pages       int    `json:"pages"`
}

var db = config.ConnectDB()

func main() {
	router := gin.Default()

	// route for create new book
	router.POST("/books", createBook)

	// Route for display all books
	router.GET("/books", getAllBooks)

	router.Run(":8080")
}

// Handler to create a new book
func createBook(c *gin.Context) {
	var newBook Book
	err := c.ShouldBind(&newBook)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : err.Error()})
		return
	}

	insertQuery := "INSERT INTO mst_book (title, author, release_year, pages) VALUES ($1, $2, $3, $4) RETURNING id"

	var bookID int
	err = db.QueryRow(insertQuery, newBook.Title, newBook.Author,
	newBook.ReleaseYear, newBook.Pages).Scan(&bookID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "gagal menambahkan buku"})
	}

	newBook.ID = bookID

	c.JSON(http.StatusCreated, newBook)

}

// Handler to display all books or books based on title search
func getAllBooks(c *gin.Context) {
	searchTitle := c.Query("title")

	query := "SELECT id, title, author, release_year, pages FROM mst_book"

	var err error
	var rows *sql.Rows
	
	// Check whether the title in the parameter is empty or not
	// if not empty :
	if searchTitle != "" {
		query += " WHERE title ILIKE '%' || $1 || '%'"
		rows, err = db.Query(query, searchTitle)
	}else {
		rows, err = db.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "Internal server error"})
		return		
	}
	defer rows.Close()

	var matchedBook []Book

	for rows.Next(){
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author,
		&book.ReleaseYear, &book.Pages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : "Internal server error"})
			return
		}

		matchedBook = append(matchedBook, book)
			
	}

	// if title not matched
	if len(matchedBook) > 0 {
		c.JSON(http.StatusOK, matchedBook)
	}else {
		c.JSON(http.StatusNotFound, gin.H{"error" : "Book not found"})
	}

}
