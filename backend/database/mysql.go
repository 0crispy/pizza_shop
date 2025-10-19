package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

var DATABASE *sql.DB
var PASSWORD_HASH_PEPPER []byte

func Init() {
	_ = godotenv.Load()

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	fmt.Println(pass)
	PASSWORD_HASH_PEPPER = []byte(os.Getenv("DB_PEPPER"))

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/pizza_shop?parseTime=true&loc=Local", user, pass)

	var err error
	DATABASE, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := DATABASE.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MySQL!")

	// Control DB reset with env var. If DB_RESET=true/1, drop and reseed.
	reset := strings.EqualFold(os.Getenv("DB_RESET"), "true") || os.Getenv("DB_RESET") == "1"
	if reset {
		InitDatabaseDev()
	}
}

func Close() {
	DATABASE.Close()
}
