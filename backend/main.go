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
	http.HandleFunc("/admin/ingredient/update", handlers.AdminUpdateIngredientHandler)
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
	http.HandleFunc("/extra-items/list", handlers.ListExtraItemsHandler)

	http.HandleFunc("/admin/extra-items/create", handlers.CreateExtraItemHandler)
	http.HandleFunc("/admin/extra-items/update", handlers.UpdateExtraItemHandler)
	http.HandleFunc("/admin/extra-items/delete", handlers.DeleteExtraItemHandler)
	http.HandleFunc("/admin/users/list", handlers.AdminGetAllUsersHandler)
	http.HandleFunc("/admin/users/delete", handlers.AdminDeleteUserHandler)
	http.HandleFunc("/admin/users/create", handlers.AdminCreateUserHandler)
	http.HandleFunc("/admin/orders/list", handlers.AdminGetAllOrdersHandler)
	http.HandleFunc("/admin/orders/delete", handlers.AdminDeleteOrderHandler)
	http.HandleFunc("/admin/orders/update-status", handlers.AdminUpdateOrderStatusHandler)
	http.HandleFunc("/admin/delivery/list", handlers.AdminGetAllDeliveryPersonsHandler)
	http.HandleFunc("/admin/delivery/delete", handlers.AdminDeleteDeliveryPersonHandler)

	http.HandleFunc("/admin/discount/create", handlers.CreateDiscountCodeHandler)
	http.HandleFunc("/admin/discount/update", handlers.UpdateDiscountCodeHandler)
	http.HandleFunc("/admin/discount/delete", handlers.DeleteDiscountCodeHandler)
	http.HandleFunc("/admin/orders/assign-delivery", handlers.AssignDeliveryPersonHandler)

	http.HandleFunc("/api/validate-discount", handlers.ValidateDiscountCodeHandler)
	http.HandleFunc("/api/extra-items", handlers.ListExtraItemsHandler)
	http.HandleFunc("/api/check-birthday", handlers.CheckBirthdayDiscountHandler)

	http.HandleFunc("/delivery_person", handlers.DeliveryPerson)

	// Delivery person endpoints
	http.HandleFunc("/delivery/available", handlers.GetAvailableDeliveriesHandler)
	http.HandleFunc("/delivery/assigned", handlers.GetAssignedDeliveriesHandler)
	http.HandleFunc("/delivery/assign", handlers.AssignDeliveryHandler)
	http.HandleFunc("/delivery/update-status", handlers.UpdateDeliveryStatusHandler)

	fmt.Printf("Server running on http://localhost:%s\n", PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil)
}
