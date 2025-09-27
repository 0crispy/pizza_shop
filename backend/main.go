package main

import (
	"fmt"
	"net/http"
	database "pizza_shop/backend/database"
	"pizza_shop/backend/handlers"
)

const PORT = "8080"

func main() {
	database.Init()
	defer database.Close()

	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/pizza", handlers.PizzaHandler)
	http.HandleFunc("/home", handlers.HomeHandler)
	http.HandleFunc("/menu", handlers.MenuHandler)

	fmt.Printf("Server running on http://localhost:%s\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
}
