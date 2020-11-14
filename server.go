package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "database/sql"
    "log"
    "os"
    _ "github.com/lib/pq"
)

type Todo struct {
    ID int `json:"id"`
    Title string `json:"title"`
    Status string `json:"status"`
}

var DB *sql.DB

func init() {
    var err error
    DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("DB connection error", err)
    }
}

// SQL injection flaws
func getTodosHandler(c *gin.Context) {
    status := c.Query("status")
    statement := "SELECT id, title, status FROM todos"
    filterNum := 0
    if len(status) != 0 {
        statement += " WHERE status = '" + status + "'"
        filterNum++
    }
    title := c.Query("title")
    if len(title) != 0 {
        if filterNum == 0 {
            statement += " WHERE "
        } else {
            statement += " AND "
        }
        statement += " title = '" + title + "'"
    }
    stmt, err := DB.Prepare(statement)
    if err != nil {
        c.JSON(http.StatusBadRequest, err)
        return
    }
    rows, err := stmt.Query()
    if err != nil {
        c.JSON(http.StatusBadRequest, err)
        return
    }

    todos := []Todo{}
    for rows.Next() {
        t := Todo{}
        rows.Scan(&t.ID, &t.Title, &t.Status)
        todos = append(todos, t)
    }
    c.JSON(http.StatusOK, todos)
}

func getTodoByIDHandler(c *gin.Context) {
    statement := "SELECT id, title, status FROM todos WHERE id = " + c.Param("id")
    stmt, err := DB.Prepare(statement)
    if err != nil {
        c.JSON(http.StatusBadRequest, err)
        return
    }
    rows, err := stmt.Query()
    if err != nil {
        c.JSON(http.StatusBadRequest, err)
        return
    }

    var t Todo
    for rows.Next() {
        rows.Scan(&t.ID, &t.Title, &t.Status)
    }
    if t.ID == 0 {
        c.JSON(http.StatusOK, gin.H{})
    } else {
        c.JSON(http.StatusOK, t)
    }
}

func postTodosHandler(c *gin.Context) {
    var json Todo
    if err := c.ShouldBindJSON(&json); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{ "error": err.Error() })
        return
    }

    row := DB.QueryRow("INSERT INTO todos (title, status) VALUES ($1, $2) RETURNING id", json.Title, json.Status)
    err := row.Scan(&json.ID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{ "error": err.Error() })
        return
    }
    c.JSON(http.StatusOK, json)
}

func putTodosHandler(c *gin.Context) {
    var json Todo
    if err := c.ShouldBindJSON(&json); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{ "error": err.Error() })
        return
    }

    DB.QueryRow("UPDATE todos SET title = $2, status = $3 WHERE id = $1", c.Param("id"), json.Title, json.Status)
    c.JSON(http.StatusOK, gin.H{ "status": "success" })
}

func deleteTodosHandler(c *gin.Context) {
    DB.QueryRow("DELETE FROM todos WHERE id = $1", c.Param("id"))
    c.JSON(http.StatusOK, gin.H{ "status": "deleted" })
}

func main() {
    defer DB.Close()

    srv := gin.Default()
    srv.GET("/todos", getTodosHandler)
    srv.GET("/todos/:id", getTodoByIDHandler)
    srv.POST("/todos", postTodosHandler)
    srv.PUT("/todos/:id", putTodosHandler)
    srv.DELETE("/todos/:id", deleteTodosHandler)
    srv.Run(os.Getenv("PORT"))
}

