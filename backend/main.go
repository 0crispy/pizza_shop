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
	http.HandleFunc("/account", handlers.AccountHandler)
	http.HandleFunc("/getAccountDetails", handlers.GetAccountDetailsHandler)

	http.HandleFunc("/admin", handlers.AdminHandler)
	http.HandleFunc("/admin/ingredient/create", handlers.AdminCreateIngredientHandler)
	http.HandleFunc("/admin/ingredient/list", handlers.AdminListIngredientsHandler)
	http.HandleFunc("/admin/ingredient/delete", handlers.AdminDeleteIngredientHandler)
	http.HandleFunc("/admin/pizza/create", handlers.AdminCreatePizzaHandler)
	http.HandleFunc("/admin/pizza/list", handlers.AdminListPizzasHandler)
	http.HandleFunc("/admin/pizza/delete", handlers.AdminDeletePizzaHandler)

	fmt.Printf("Server running on http://localhost:%s\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
}
