package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/go-sql-driver/mysql"

	"github.com/joho/godotenv"

	_ "github.com/joho/godotenv/autoload"
)

type List struct {
	ID     int64
	Task   string
	Status string
}

type Todo struct {
	Id        int
	Item      string
	Completed string
}

var err error

const filename = "list.html"

func Redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var db *sql.DB
var lis List

func main() {
	// Capture connection properties.

	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "recordings",
		AllowNativePasswords: false,
	}
	_ = godotenv.Load("pass.env")

	secretKey := os.Getenv("SECRET_KEY")
	dsn := "root:" + secretKey + "@tcp(127.0.0.1:3306)/recordings"

	//item := c.PostForm("item")
	var err error
	//item := lis.Task
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	dsn = dsn + "?allowNativePasswords=false"
	log.Println("using", dsn)
	db, _ = sql.Open("mysql", dsn)

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	router := gin.Default()

	router.GET("/albums", Task)
	router.GET("/albums/:id", GetId)
	router.GET("/complete/:id", Complete)
	router.GET("/delete/:id", Delete)
	router.POST("/albums", func(c *gin.Context) {

		log.Print(err)
		item := c.PostForm("item")
		result, err := db.Exec("INSERT INTO list (task, status) VALUES (?, ?)", item, "")
		if err != nil {
			panic(err.Error())
		}
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(lastInsertId)
		id := lastInsertId

		c.Redirect(http.StatusFound, "/albums/"+strconv.Itoa(int(id)))

	})
	router.LoadHTMLFiles("list.html")

	router.Run("localhost:8080")
}
func GetId(c *gin.Context) {
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(intID)
	}
	log.Print(err)

	list, err := listByID(int64(intID))

	todos := make([]Todo, 0)
	todos = append(todos, Todo{
		Id:   intID,
		Item: list.Task,
	})
	todos = append(todos, Todo{
		Id:   intID,
		Item: "ej",
	})
	c.HTML(http.StatusOK, "list.html", gin.H{
		"Title": "hej",
		"todos": todos,
	})

	if err == sql.ErrNoRows {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, list)
	fmt.Println(todos)

}

func listbyTasks(name string) ([]List, error) {
	
	var lists []List

	rows, err := db.Query("SELECT * FROM list WHERE task = ?", name)
	if err != nil {
		return nil, fmt.Errorf("listbyTasks %q: %v", name, err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var lis List
		if err := rows.Scan(&lis.ID, &lis.Task, &lis.Status); err != nil {
			return nil, fmt.Errorf("listbyTasks %q: %v", name, err)
		}
		lists = append(lists, lis)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listbyTasks %q: %v", name, err)
	}
	return lists, nil
}

func listByID(id int64) (List, error) {

	var lis List

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", lis)

	row := db.QueryRow("SELECT * from list where id = ?", id)
	if err := row.Scan(&lis.ID, &lis.Task, &lis.Status); err != nil {
		if err == sql.ErrNoRows {
			return lis, fmt.Errorf("albumsById %d: no such album", id)
		}
		return lis, fmt.Errorf("albumsById %d: %v", id, err)

	}
	log.Print(err)
	return lis, nil

}
func Task(c *gin.Context) {

	tasks := make([]*List, 0)
	task := &List{}
	todos := make([]Todo, 0)

	rows, err := db.Query("SELECT id, task, status FROM list")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&task.ID, &task.Task, &task.Status)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		tasks = append(tasks, task)
		todos = append(todos, Todo{
			Id:        int(task.ID),
			Item:      task.Task,
			Completed: task.Status,
		})

	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"Title": "hej",
		"todos": todos,
	})
	if err = rows.Err(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

}

func Complete(c *gin.Context) {
	id := c.Param("id")

	intID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(intID)
	}
	c.Redirect(http.StatusFound, "/albums/")

	_, a := db.Exec("UPDATE list SET status = 'Completed' WHERE id = ?", id)
	fmt.Println(a)
	list, err := listByID(int64(intID))

	todos := make([]Todo, 0)
	todos = append(todos, Todo{
		Id:        intID,
		Item:      list.Task,
		Completed: "Completed",
	})

	c.HTML(http.StatusOK, "list.html", gin.H{
		"Title": "hej",
		"todos": todos,
	})

}

func Delete(c *gin.Context) {

	id := c.Param("id")
	c.Redirect(http.StatusFound, "/albums/")
	_, a := db.Exec("DELETE FROM list WHERE id = ?", id)
	fmt.Println(a)

}
