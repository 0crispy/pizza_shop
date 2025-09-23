package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

var DATABASE *sql.DB

func Init() {
	_ = godotenv.Load()

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/pizza_shop", user, pass)

	var err error
	DATABASE, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := DATABASE.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MySQL!")

	InitDatabaseDev()
}

func Close() {
	DATABASE.Close()
}
