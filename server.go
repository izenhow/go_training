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

func getTodosHandler(c *gin.Context) {
    statusFilter := c.Query("status")
    preparedStatement := "SELECT id, title, status FROM todos"
    filterNum := 0
    // SQL injection flaws
    if len(statusFilter) != 0 {
        preparedStatement += " WHERE status = '" + statusFilter + "'"
        filterNum++
    }
    titleFilter := c.Query("title")
    if len(titleFilter) != 0 {
        if filterNum == 0 {
            preparedStatement += " WHERE "
        } else {
            preparedStatement += " AND "
        }
        preparedStatement += " title = '" + titleFilter + "'"
    }
    stmt, err := DB.Prepare(preparedStatement)
    if err != nil {
        c.JSON(http.StatusBadRequest, err)
        return
    }
    rows, err := stmt.Query()
    if err != nil {
        c.JSON(http.StatusBadRequest, err)
        return
    }

    var todos []Todo
    for rows.Next() {
        t := Todo{}
        rows.Scan(&t.ID, &t.Title, &t.Status)
        todos = append(todos, t)
    }
    c.JSON(http.StatusOK, todos)
}

func getTodoByIDHandler(c *gin.Context) {
    preparedStatement := "SELECT id, title, status FROM todos WHERE id = " + c.Param("id")
    stmt, err := DB.Prepare(preparedStatement)
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

    serv := gin.Default()
    serv.GET("/todos", getTodosHandler)
    serv.GET("/todos/:id", getTodoByIDHandler)
    serv.POST("/todos", postTodosHandler)
    serv.PUT("/todos/:id", putTodosHandler)
    serv.DELETE("/todos/:id", deleteTodosHandler)
    serv.Run(os.Getenv("PORT"))
}

