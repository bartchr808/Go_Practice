package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// ~~~~~~~~~~~~~ Models ~~~~~~~~~~~~~

// Todo -> model for demo database
type Todo struct {
	gorm.Model
	Title     string `json:"title"`
	Completed int    `json:"completed"`
}

// TransformedTodo -> transformed Todo model for user response
type TransformedTodo struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// ~~~~~~~~~~~~~ DB ~~~~~~~~~~~~~

// Database -> get connection to CockroachDB
func Database() *gorm.DB {
	//open a db connection
	db, err := gorm.Open("postgres", "postgresql://maxroach@localhost:26257/demo?sslmode=disable")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

// ~~~~~~~~~~~~~ Routes ~~~~~~~~~~~~~

// CreateTodo -> input Todo struct row into DB
func CreateTodo(c *gin.Context) {
	// get string from json and convert to int
	completed, _ := strconv.Atoi(c.PostForm("completed"))

	// create table entry struct
	todo := Todo{Title: c.PostForm("title"), Completed: completed}

	// connect to DB and insert row
	db := Database()
	db.Save(&todo)

	// returned json
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Todo item created successfully!", "resourceId": todo.ID})
}

// FetchAllTodo -> get all Todo DB entries in demo table
func FetchAllTodo(c *gin.Context) {
	var todos []Todo
	var _todos []TransformedTodo

	db := Database()
	db.Find(&todos)

	if len(todos) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	//transforms the todos for building a good response
	for _, item := range todos {
		var completed bool
		if item.Completed == 1 {
			completed = true
		} else {
			completed = false
		}
		_todos = append(_todos, TransformedTodo{ID: item.ID, Title: item.Title, Completed: completed})
	}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _todos})
}

// FetchSingleTodo -> get specific Todo DB entry by ID
func FetchSingleTodo(c *gin.Context) {
	var todo Todo
	todoID := c.Param("id")

	db := Database()
	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	var completed bool
	if todo.Completed == 1 {
		completed = true
	} else {
		completed = false
	}

	_todo := TransformedTodo{ID: todo.ID, Title: todo.Title, Completed: completed}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _todo})
}

// UpdateTodo -> update specific Todo DB entry by ID
func UpdateTodo(c *gin.Context) {
	var todo Todo
	todoID := c.Param("id")
	db := Database()
	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	db.Model(&todo).Update("title", c.PostForm("title"))
	completed, _ := strconv.Atoi(c.PostForm("completed"))
	db.Model(&todo).Update("completed", completed)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Todo updated successfully!"})
}

// DeleteTodo -> delete specific Todo DB entry by ID
func DeleteTodo(c *gin.Context) {
	var todo Todo
	todoID := c.Param("id")
	db := Database()
	db.First(&todo, todoID)

	if todo.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}

	db.Delete(&todo)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Todo deleted successfully!"})
}

func main() {
	db := Database()
	db.AutoMigrate(&Todo{})

	router := gin.Default()
	v1 := router.Group("/api/v1/todos")
	{
		v1.POST("/", CreateTodo)
		v1.GET("/", FetchAllTodo)
		v1.GET("/:id", FetchSingleTodo)
		v1.PUT("/:id", UpdateTodo)
		v1.DELETE("/:id", DeleteTodo)
	}
	router.Run(":5432")
}
