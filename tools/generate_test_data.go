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

var statuses = []string{"IN_PROGRESS", "OUT_FOR_DELIVERY", "DELIVERED", "FAILED"}

func randomString(options []string) string {
	return options[rand.Intn(len(options))]
}

func randomDate(minYear, maxYear int) string {
	year := minYear + rand.Intn(maxYear-minYear+1)
	month := 1 + rand.Intn(12)
	day := 1 + rand.Intn(28)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func randomDateInRange(startDate, endDate time.Time) time.Time {
	delta := endDate.Unix() - startDate.Unix()
	sec := rand.Int63n(delta)
	return startDate.Add(time.Duration(sec) * time.Second)
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
		username := fmt.Sprintf("%s%s%d", firstName, lastName, rand.Intn(10000))
		name := fmt.Sprintf("%s %s", firstName, lastName)

		// Generate birth dates for different age groups
		// Ages: 18-75 years old (born between 1950-2007)
		birthYear := 1950 + rand.Intn(58)
		noBirthDate := rand.Float32() < 0.05 // 5% chance of no birth date

		customer := database.Customer{
			Username:    username,
			Password:    "password123",
			Name:        name,
			Gender:      randomString(genders),
			BirthDate:   randomDate(birthYear, birthYear),
			NoBirthDate: noBirthDate,
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
		if (i+1)%10 == 0 {
			fmt.Printf("  ✓ Created %d customers...\n", i+1)
		}
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

	// Order dates: last 2 years
	now := time.Now()
	twoYearsAgo := now.AddDate(-2, 0, 0)

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

		// Get userID from customer
		var userID int
		err = database.DATABASE.QueryRow(
			"SELECT user_id FROM customer WHERE id = ?",
			customerID,
		).Scan(&userID)

		if err != nil {
			log.Printf("Failed to get user_id for customer ID %d: %v\n", customerID, err)
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
			userID,
			customer.Address,
			customer.PostCode,
			pizzaItems,
			extraItemsToOrder,
			nil,
		)

		if err != nil {
			log.Printf("Failed to create order: %v\n", err)
			continue
		}

		// Random status weighted towards DELIVERED
		statusRand := rand.Float32()
		var status string
		if statusRand < 0.80 { // 80% delivered
			status = "DELIVERED"
		} else if statusRand < 0.90 { // 10% in progress
			status = "IN_PROGRESS"
		} else if statusRand < 0.95 { // 5% out for delivery
			status = "OUT_FOR_DELIVERY"
		} else { // 5% failed
			status = "FAILED"
		}

		err = database.UpdateOrderStatus(orderID, status)
		if err != nil {
			log.Printf("Failed to update order status: %v\n", err)
		}

		// Random order date in last 2 years
		orderDate := randomDateInRange(twoYearsAgo, now)
		_, err = database.DATABASE.Exec(
			"UPDATE orders SET timestamp = ? WHERE id = ?",
			orderDate,
			orderID,
		)
		if err != nil {
			log.Printf("Failed to update order timestamp: %v\n", err)
		}

		if (i+1)%50 == 0 {
			fmt.Printf("  ✓ Created %d orders...\n", i+1)
		}
	}
	fmt.Printf("  ✓ Created %d orders total\n", count)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Pizza Shop Test Data Generator ===")
	fmt.Println()

	database.Init()
	defer database.Close()

	fmt.Println("Connected to database!")
	fmt.Println()

	customerIDs := generateCustomers(100)
	fmt.Printf("\n✓ Created %d customers\n\n", len(customerIDs))

	generateDeliveryPeople(10)
	fmt.Println("\n✓ Created 10 delivery people\n")

	generateOrders(customerIDs, 500)
	fmt.Println("\n✓ Created 500 orders\n")

	fmt.Println("=== Test Data Generation Complete! ===")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  • 100 customers with ages 18-75\n")
	fmt.Printf("  • 10 delivery people\n")
	fmt.Printf("  • 500 orders spanning last 2 years\n")
	fmt.Println()
	fmt.Println("Sample credentials:")
	fmt.Println("  Customers: Check usernames like AlexSmith*, JamieJones*, etc.")
	fmt.Println("  Delivery: Check usernames like delivery_AlexSmith*, etc.")
	fmt.Println("  Password for all: password123")
}
