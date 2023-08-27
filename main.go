package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/db_books")
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	initDB()
	defer db.Close()

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	router.Use(cors.New(config))

	router.GET("/books", getBooks)
	router.GET("/books/:id", bookByIDHandler)
	router.POST("/books", createBook)
	router.PATCH("/checkout", checkoutBook)
	router.PATCH("/return", returnBook)
	router.Run("localhost:8080")
}

func getBooks(c *gin.Context) {
	var books []book
	query := "SELECT id, title, author, quantity FROM books"
	rows, err := db.Query(query)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var b book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Quantity); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, b)
	}

	c.IndentedJSON(http.StatusOK, books)
}

func checkoutBook(c *gin.Context) {
	id := c.DefaultQuery("id", "")

	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "missing id query parameter"})
		return
	}

	book, err := getBookByID(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}

	if book.Quantity <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "book out of stock"})
		return
	}

	book.Quantity--

	// Update the book's quantity in the database
	if err := updateBookQuantity(book); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to update book quantity"})
		return
	}

	c.IndentedJSON(http.StatusOK, book)
}

func returnBook(c *gin.Context) {
	id := c.DefaultQuery("id", "")

	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "missing id query parameter"})
		return
	}

	book, err := getBookByID(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
		return
	}

	book.Quantity++

	// Update the book's quantity in the database
	if err := updateBookQuantity(book); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to update book quantity"})
		return
	}

	c.IndentedJSON(http.StatusOK, book)
}

func updateBookQuantity(b *book) error {
	query := "UPDATE books SET quantity = ? WHERE id = ?"
	_, err := db.Exec(query, b.Quantity, b.ID)
	return err
}
func bookByIDHandler(c *gin.Context) {
	id := c.Param("id")
	book, err := getBookByID(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "books not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, book)
}

func getBookByID(id string) (*book, error) {
	var b book
	query := "SELECT id, title, author, quantity FROM books WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&b.ID, &b.Title, &b.Author, &b.Quantity)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func createBook(c *gin.Context) {
	var newBook book

	if err := c.BindJSON(&newBook); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO books (id, title, author, quantity) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, newBook.ID, newBook.Title, newBook.Author, newBook.Quantity)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newBook)
}
