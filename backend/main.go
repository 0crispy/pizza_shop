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
	http.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/cart.html")
	})
	http.HandleFunc("/order-confirmation", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/order-confirmation.html")
	})

	http.HandleFunc("/order/create", handlers.CreateOrderHandler)
	http.HandleFunc("/order/list", handlers.GetOrdersHandler)
	http.HandleFunc("/order/details", handlers.GetOrderDetailsHandler)

	// Admin management routes
	http.HandleFunc("/admin/users/list", handlers.AdminGetAllUsersHandler)
	http.HandleFunc("/admin/users/delete", handlers.AdminDeleteUserHandler)
	http.HandleFunc("/admin/users/create", handlers.AdminCreateUserHandler)
	http.HandleFunc("/admin/orders/list", handlers.AdminGetAllOrdersHandler)
	http.HandleFunc("/admin/orders/delete", handlers.AdminDeleteOrderHandler)
	http.HandleFunc("/admin/orders/update-status", handlers.AdminUpdateOrderStatusHandler)
	http.HandleFunc("/admin/delivery/list", handlers.AdminGetAllDeliveryPersonsHandler)
	http.HandleFunc("/admin/delivery/delete", handlers.AdminDeleteDeliveryPersonHandler)

	http.HandleFunc("/delivery_person", handlers.DeliveryPerson)

	fmt.Printf("Server running on http://localhost:%s\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
}
