package main

import (
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	"log"
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int `json:"id"`
	Name  string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}


var db *sql.DB

func CreateTable() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to Database error", err)
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS customer(
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err = db.Exec(createTb)

	if err != nil {
		log.Fatal("Cannot create table", err)
	}

	fmt.Println("Create Table success")
}


func createCustomerHandler(c *gin.Context) {


	var cus Customer
	if err:= c.ShouldBindJSON(&cus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO customer (name, email, status) values ($1, $2 ,$3) RETURNING id",cus.Name,cus.Email,cus.Status)
	var id int
	err := row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status":"Error"})
		return
	}
	cus.ID = id 
	c.JSON(http.StatusCreated, cus)
}

var customers = []Customer{}
func getCustomerHandler(c *gin.Context) {

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer")
	if err != nil {
		log.Fatal("Cannot prepare query all customer", err)
	}

	rows, err  := stmt.Query()
	if err != nil {
		log.Fatal("Cannot query all customers", err)
	}

	for rows.Next() {

		cus := Customer{}
		err := rows.Scan(&cus.ID, &cus.Name, &cus.Email, &cus.Status)

		if err != nil {
			log.Fatal("Cannot scan row into variable",err)
		}
		customers = append(customers, cus )	
	}
	c.JSON(http.StatusOK, customers)


}

func getCustomerByIDHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	stmt, err := db.Prepare("SELECT id, name, email , status FROM customer WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	row := stmt.QueryRow(id)
	cus := Customer{}
	err = row.Scan(&cus.ID, &cus.Name, &cus.Email, &cus.Status)
	if err != nil {
		log.Fatal("Cannot Scan row into variables",err)
	}
	c.JSON(http.StatusOK, cus)
	return
}

func updateCustomerByIDHandler(c *gin.Context){
	cus := Customer{}
	err := c.ShouldBindJSON(&cus)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("UPDATE customer SET name=$2, email=$3, status=$4 WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if _, err := stmt.Exec(id, cus.Name, cus.Email, cus.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	cus.ID = id

	c.JSON(http.StatusOK, cus)
}

func deleteCustomerhandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))


	stmt, err := db.Prepare("DELETE FROM customer WHERE id = $1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if _, err := stmt.Exec(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
	
}


func setUp() *gin.Engine {
	r := gin.Default()

	r.Use(loginMiddleware)
	r.POST("/customers", createCustomerHandler)
	r.GET("/customers", getCustomerHandler)
	r.GET("/customers/:id", getCustomerByIDHandler)
	r.PUT("/customers/:id", updateCustomerByIDHandler)
	r.DELETE("/customers/:id", deleteCustomerhandler)

	return r
}

func loginMiddleware(c *gin.Context) {
	log.Println("starting middleware")
	authKey := c.GetHeader("Authorization")
	if authKey == "token2019wrong_token" {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		c.Abort()
		return
	}
	c.Next()

	log.Println("ending middleware")
}

func main() {
	CreateTable()
	
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Cannot connect to detabase", err)
	}
	defer db.Close() 

	
	r := setUp()

	r.Run(":2019")
	


}