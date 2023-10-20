package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"simple-web-app-with-db/config"
	"strconv"
	"strings"

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

	// Route for create new book
	router.POST("/books", createBook)

	// Route for display all books
	router.GET("/books", getAllBooks)

	// Route for display book by id
	router.GET("/books/:id", getBookById)

	// Route for update book by id
	router.PUT("/books/:id", updateBookById)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "internal server error"})
		return		
	}
	defer rows.Close()

	var matchedBook []Book

	for rows.Next(){
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author,
		&book.ReleaseYear, &book.Pages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : "internal server error"})
			return
		}

		matchedBook = append(matchedBook, book)
			
	}

	// if title not matched
	if len(matchedBook) > 0 {
		c.JSON(http.StatusOK, matchedBook)
	}else {
		c.JSON(http.StatusNotFound, gin.H{"error" : "book not found"})
	}

}

// Handler to get book by id
func getBookById(c *gin.Context) {
	idParams := c.Param("id")

	bookID, err := strconv.Atoi(idParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "invalid book id"})
		return
	}

	query := "SELECT * FROM mst_book WHERE id = $1"

	var book Book

	// Scans a row of books containing matching id
	err = db.QueryRow(query, bookID).Scan(&book.ID, &book.Title, 
	&book.Author, &book.ReleaseYear, &book.Pages)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error" : "book not found"})
	}else {
		if book.ID == bookID {
			c.JSON(http.StatusOK, book)
			return
		}
	}

}

func updateBookById(c *gin.Context) {
	idParams := c.Param("id")

	bookID, err := strconv.Atoi(idParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "invalid book id"})
		return
	}

	var updatedBook Book

	err = c.ShouldBind(&updatedBook)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error" : err.Error()})
		return
	}

	var book Book
	query := "SELECT id, title, author, release_year, pages FROM mst_book WHERE id = $1"
	err = db.QueryRow(query, bookID).Scan(&book.ID, &book.Title, 
	&book.Author, &book.ReleaseYear, &book.Pages)
	// fmt.Println("book in select : ", book)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error" : "book not found"})
		return
	}
		
	if book.ID == bookID {
		// check if the object have a value or not
		if strings.TrimSpace(updatedBook.Title) != "" {
			book.Title = updatedBook.Title
		}
		if strings.TrimSpace(updatedBook.Author) != "" {
			book.Author = updatedBook.Author
		}
		if strings.TrimSpace(updatedBook.ReleaseYear) != "" {
			book.ReleaseYear = updatedBook.ReleaseYear
		}
		if updatedBook.Pages != 0 {
			book.Pages = updatedBook.Pages
		}
		// fmt.Println("book in condition : ", book)
		query = "UPDATE mst_book SET title = $2, author = $3, release_year = $4, pages = $5 WHERE id = $1"
		_, err = db.Exec(query, bookID, &book.Title, &book.Author, &book.ReleaseYear, &book.Pages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : "internal server error"})
			return
		}
		updatedBook.ID = bookID
		c.JSON(http.StatusOK, book)
		fmt.Println("update book : ", book)
	}

}

// task : bagaimana cara data yang diisikan tidak di update jika bernilai string kosong