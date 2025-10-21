package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"pizza_shop/backend/database"
)

var firstNames = []string{
	"Alex", "Jamie", "Jordan", "Taylor", "Morgan",
	"Casey", "Riley", "Avery", "Quinn", "Sage",
	"Blake", "Drew", "Reese", "Cameron", "Dakota",
	"Rowan", "Skyler", "Phoenix", "River", "Emerson",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones",
	"Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
	"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
	"Thomas", "Taylor", "Moore", "Jackson", "Martin",
}

var streets = []string{
	"Main St", "Oak Ave", "Maple Dr", "Cedar Ln", "Pine Rd",
	"Elm St", "Park Ave", "Washington St", "Lake Dr", "Hill Rd",
	"River Rd", "Forest Ave", "Sunset Blvd", "Broadway", "Market St",
}

var genders = []string{"Male", "Female", "Non-binary", "Prefer not to say"}

var statuses = []string{"IN_PROGRESS", "DELIVERED", "FAILED"}

func randomString(options []string) string {
	return options[rand.Intn(len(options))]
}

func randomDate(minYear, maxYear int) string {
	year := minYear + rand.Intn(maxYear-minYear)
	month := 1 + rand.Intn(12)
	day := 1 + rand.Intn(28)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func randomAddress() string {
	number := 100 + rand.Intn(9900)
	return fmt.Sprintf("%d %s", number, randomString(streets))
}

func randomPostalCode() string {
	return fmt.Sprintf("%05d", 10000+rand.Intn(90000))
}

func generateCustomers(count int) []int {
	fmt.Printf("Generating %d customers...\n", count)
	var customerIDs []int

	for i := 0; i < count; i++ {
		firstName := randomString(firstNames)
		lastName := randomString(lastNames)
		username := fmt.Sprintf("%s%s%d", firstName, lastName, rand.Intn(1000))
		name := fmt.Sprintf("%s %s", firstName, lastName)

		customer := database.Customer{
			Username:    username,
			Password:    "password123",
			Name:        name,
			Gender:      randomString(genders),
			BirthDate:   randomDate(1970, 2005),
			NoBirthDate: rand.Float32() < 0.1,
			Address:     randomAddress(),
			PostCode:    randomPostalCode(),
		}

		success, msg := database.TryAddCustomer(customer)
		if !success {
			log.Printf("Failed to create customer %s: %s\n", username, msg)
			continue
		}

		userID, err := database.GetUserIDFromUsername(username)
		if err != nil {
			log.Printf("Failed to get user ID for %s: %v\n", username, err)
			continue
		}

		customerID, err := database.GetCustomerIDFromUserID(userID)
		if err != nil {
			log.Printf("Failed to get customer ID for %s: %v\n", username, err)
			continue
		}

		customerIDs = append(customerIDs, customerID)
		fmt.Printf("  ✓ Created customer: %s (ID: %d)\n", username, customerID)
	}

	return customerIDs
}

func generateDeliveryPeople(count int) {
	fmt.Printf("Generating %d delivery people...\n", count)

	for i := 0; i < count; i++ {
		firstName := randomString(firstNames)
		lastName := randomString(lastNames)
		username := fmt.Sprintf("delivery_%s%s%d", firstName, lastName, rand.Intn(1000))
		name := fmt.Sprintf("%s %s", firstName, lastName)

		err := database.AddUser(username, "password123", database.DeliveryRole)
		if err != nil {
			log.Printf("Failed to create delivery user %s: %v\n", username, err)
			continue
		}

		userID, err := database.GetUserIDFromUsername(username)
		if err != nil {
			log.Printf("Failed to get user ID for %s: %v\n", username, err)
			continue
		}

		_, err = database.DATABASE.Exec(
			"INSERT INTO delivery_person (name, user_id) VALUES (?, ?)",
			name,
			userID,
		)
		if err != nil {
			log.Printf("Failed to create delivery_person record for %s: %v\n", username, err)
			continue
		}

		fmt.Printf("  ✓ Created delivery person: %s (username: %s)\n", name, username)
	}
}

type ExtraItem struct {
	ID          int
	Name        string
	Category    string
	Price       float64
	Description string
}

func getExtraItems() ([]ExtraItem, error) {
	query := `SELECT id, name, category, price FROM extra_item`
	rows, err := database.DATABASE.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExtraItem
	for rows.Next() {
		var item ExtraItem
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func generateOrders(customerIDs []int, count int) {
	fmt.Printf("Generating %d orders...\n", count)

	pizzas, err := database.GetAllPizzasWithPrice()
	if err != nil {
		log.Fatalf("Failed to get pizzas: %v\n", err)
	}

	if len(pizzas) == 0 {
		log.Println("⚠ No pizzas found in database. Please add pizzas first!")
		return
	}

	extraItems, err := getExtraItems()
	if err != nil {
		log.Fatalf("Failed to get extra items: %v\n", err)
	}

	for i := 0; i < count; i++ {
		if len(customerIDs) == 0 {
			log.Println("No customers available to create orders")
			return
		}

		customerID := customerIDs[rand.Intn(len(customerIDs))]

		var customer database.Customer
		err := database.DATABASE.QueryRow(
			"SELECT name, address, postal_code FROM customer WHERE id = ?",
			customerID,
		).Scan(&customer.Name, &customer.Address, &customer.PostCode)

		if err != nil {
			log.Printf("Failed to get customer details for ID %d: %v\n", customerID, err)
			continue
		}

		numPizzas := 1 + rand.Intn(4)
		var pizzaItems []struct {
			PizzaID  int
			Quantity int
		}

		for j := 0; j < numPizzas; j++ {
			pizza := pizzas[rand.Intn(len(pizzas))]
			quantity := 1 + rand.Intn(3)
			pizzaItems = append(pizzaItems, struct {
				PizzaID  int
				Quantity int
			}{pizza.ID, quantity})
		}

		var extraItemsToOrder []struct {
			ExtraItemID int
			Quantity    int
		}

		if len(extraItems) > 0 && rand.Float32() < 0.6 {
			numExtras := 1 + rand.Intn(3)
			for j := 0; j < numExtras && j < len(extraItems); j++ {
				extraItem := extraItems[rand.Intn(len(extraItems))]
				quantity := 1 + rand.Intn(2)
				extraItemsToOrder = append(extraItemsToOrder, struct {
					ExtraItemID int
					Quantity    int
				}{extraItem.ID, quantity})
			}
		}

		orderID, err := database.CreateOrderWithTransaction(
			customerID,
			customer.Address,
			customer.PostCode,
			pizzaItems,
			extraItemsToOrder,
		)

		if err != nil {
			log.Printf("Failed to create order: %v\n", err)
			continue
		}

		status := randomString(statuses)
		err = database.UpdateOrderStatus(orderID, status)
		if err != nil {
			log.Printf("Failed to update order status: %v\n", err)
		}

		orderDate := time.Now().Add(-time.Duration(rand.Intn(90*24)) * time.Hour)
		_, err = database.DATABASE.Exec(
			"UPDATE orders SET timestamp = ? WHERE id = ?",
			orderDate,
			orderID,
		)
		if err != nil {
			log.Printf("Failed to update order timestamp: %v\n", err)
		}

		fmt.Printf("  ✓ Created order #%d for customer %s (%d pizzas, %d extras, status: %s)\n",
			orderID, customer.Name, len(pizzaItems), len(extraItemsToOrder), status)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Pizza Shop Test Data Generator ===\n")

	database.Init()
	defer database.Close()

	fmt.Println("Connected to database!\n")

	customerIDs := generateCustomers(20)
	fmt.Printf("\n✓ Created %d customers\n\n", len(customerIDs))

	generateDeliveryPeople(5)
	fmt.Println("\n✓ Created 5 delivery people\n")

	generateOrders(customerIDs, 30)
	fmt.Println("\n✓ Created 30 orders\n")

	fmt.Println("=== Test Data Generation Complete! ===")
	fmt.Println("\nSample credentials:")
	fmt.Println("  Customers: Check usernames like AlexSmith*, JamieJones*, etc.")
	fmt.Println("  Delivery: Check usernames like delivery_AlexSmith*, etc.")
	fmt.Println("  Password for all: password123")
}
