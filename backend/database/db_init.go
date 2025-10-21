package database

import (
	"log"
)

// Used for first-time initialization of the whole database.
func InitDatabaseDev() {
	queries := []string{
		`SET FOREIGN_KEY_CHECKS = 0;`,

		// Drop all tables first (in reverse dependency order)
		`DROP TABLE IF EXISTS discount_usage;`,
		`DROP TABLE IF EXISTS order_extra_item;`,
		`DROP TABLE IF EXISTS order_pizza;`,
		`DROP TABLE IF EXISTS orders;`,
		`DROP TABLE IF EXISTS delivery_person;`,
		`DROP TABLE IF EXISTS discount_code;`,
		`DROP TABLE IF EXISTS extra_item;`,
		`DROP TABLE IF EXISTS customer;`,
		`DROP TABLE IF EXISTS user;`,
		`DROP TABLE IF EXISTS pizza_ingredient;`,
		`DROP TABLE IF EXISTS ingredient;`,
		`DROP TABLE IF EXISTS pizza;`,

		// Create tables in dependency order
		`CREATE TABLE pizza (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE
		);`,

		`CREATE TABLE ingredient(
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			cost DECIMAL(10, 2) NOT NULL CHECK (cost > 0),
			has_meat BOOLEAN NOT NULL,
			has_animal_products BOOLEAN NOT NULL
		);`,

		`CREATE TABLE pizza_ingredient(
			pizza_id INT NOT NULL,
			ingredient_id INT NOT NULL,

			PRIMARY KEY (pizza_id, ingredient_id),
			FOREIGN KEY (pizza_id) REFERENCES pizza(id)
				ON DELETE CASCADE,
			FOREIGN KEY (ingredient_id) REFERENCES ingredient(id)
				ON DELETE CASCADE
		);`,

		`CREATE TABLE user(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(256) NOT NULL,
			salt VARCHAR(256) NOT NULL,
			role ENUM('ADMIN', 'DELIVERY', 'CUSTOMER') NOT NULL
		);`,

		`CREATE TABLE customer(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			name VARCHAR(100) NOT NULL,
			gender VARCHAR(50) NOT NULL,
			birth_date DATE,
			address VARCHAR(256) NOT NULL,
			postal_code VARCHAR(10) NOT NULL,
			pizza_counter TINYINT NOT NULL DEFAULT 0,

			FOREIGN KEY (user_id) REFERENCES user(id)
		)`,

		`CREATE TABLE discount_code (
			id INT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(50) NOT NULL UNIQUE,
			discount_percentage INT NOT NULL CHECK (discount_percentage > 0 AND discount_percentage <= 100),
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		)`,

		`CREATE TABLE delivery_person(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			user_id BIGINT NOT NULL,
			vehicle_type VARCHAR(50) DEFAULT 'bike',
			unavailable_until TIMESTAMP NULL DEFAULT NULL,

			FOREIGN KEY (user_id) REFERENCES user(id)
		)`,

		`CREATE TABLE orders(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			customer_id BIGINT NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			status ENUM('IN_PROGRESS', 'OUT_FOR_DELIVERY', 'DELIVERED', 'FAILED') NOT NULL,
			postal_code VARCHAR(10) NOT NULL,
			delivery_address VARCHAR(256) NOT NULL,
			discount_code_id INT DEFAULT NULL,
			delivery_person_id BIGINT DEFAULT NULL,

			FOREIGN KEY (customer_id) REFERENCES customer(id),
			FOREIGN KEY (discount_code_id) REFERENCES discount_code(id),
			FOREIGN KEY (delivery_person_id) REFERENCES delivery_person(id)
		);`,

		`CREATE TABLE order_pizza (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			order_id BIGINT NOT NULL,
			pizza_id INT NOT NULL,
			quantity INT NOT NULL CHECK (quantity > 0),
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (pizza_id) REFERENCES pizza(id)
		)`,

		`CREATE TABLE extra_item (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			category ENUM('dessert', 'drink') NOT NULL,
			price DECIMAL(10, 2) NOT NULL CHECK (price >= 0)
		)`,

		`CREATE TABLE order_extra_item (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			order_id BIGINT NOT NULL,
			extra_item_id INT NOT NULL,
			quantity INT NOT NULL CHECK (quantity > 0),
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (extra_item_id) REFERENCES extra_item(id)
		)`,

		`CREATE TABLE discount_usage (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			discount_code_id INT NOT NULL,
			used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(id),
			FOREIGN KEY (discount_code_id) REFERENCES discount_code(id),
			UNIQUE KEY unique_user_discount (user_id, discount_code_id)
		)`,

		`SET FOREIGN_KEY_CHECKS = 1;`,
	}

	for _, query := range queries {
		_, err := DATABASE.Exec(query)
		if err != nil {
			log.Println("Error executing query: ", query)
			log.Fatal(err)
		}
	}

	createIngredientDbg("Pepperoni", 101, true, true)
	createIngredientDbg("Mozzarella", 60, false, true)
	createIngredientDbg("Jalapeno", 30, false, false)
	createIngredientDbg("Tomato sauce", 50, false, false)
	createIngredientDbg("'Nduja", 200, true, true)
	createIngredientDbg("Chili pepper", 30, false, false)
	createIngredientDbg("Prosciutto", 85, true, true)
	createIngredientDbg("Mushrooms", 35, false, false)
	createIngredientDbg("Anchovies", 75, true, true)
	createIngredientDbg("Gorgonzola", 40, false, true)
	createIngredientDbg("Fontina", 40, false, true)
	createIngredientDbg("Parmigiano Reggiano", 30, false, true)
	createIngredientDbg("Nutella", 200, false, true)

	createIngredientDbg("Lobster", 500, true, true)
	createIngredientDbg("Caviar", 1000, false, true)
	createIngredientDbg("Gold leaf", 1000, false, false)
	createIngredientDbg("Foie gras", 1000, true, true)

	createIngredientDbg("Old tomato sauce", 1, false, false)
	createIngredientDbg("Weird vegan cheese", 4, false, false)

	createPizzaDbg("Marinara", []string{"Tomato sauce"})
	createPizzaDbg("Margherita", []string{"Tomato sauce", "Mozzarella"})
	createPizzaDbg("Diavola", []string{"Tomato sauce", "Mozzarella", "'Nduja", "Chili pepper"})
	createPizzaDbg("Priosciutto e Funghi", []string{"Tomato sauce", "Mozzarella", "Prosciutto", "Mushrooms"})
	createPizzaDbg("Napoli", []string{"Tomato sauce", "Mozzarella", "Anchovies"})
	createPizzaDbg("Quatro Formaggi", []string{"Tomato sauce", "Mozzarella", "Gorgonzola", "Fontina", "Parmigiano Reggiano"})
	createPizzaDbg("Nutella", []string{"Nutella"})
	createPizzaDbg("Billionaire's dream", []string{"Tomato sauce", "Mozzarella", "Lobster", "Caviar", "Gold leaf", "Foie gras"})
	createPizzaDbg("Student's dream", []string{"Old tomato sauce", "Weird vegan cheese"})

	createExtraItemDbg("Tiramisu", "dessert", 5.50)
	createExtraItemDbg("Panna Cotta", "dessert", 4.50)
	createExtraItemDbg("Gelato", "dessert", 3.50)
	createExtraItemDbg("Cannoli", "dessert", 4.00)
	createExtraItemDbg("Chocolate Cake", "dessert", 5.00)
	createExtraItemDbg("Coca Cola", "drink", 2.50)
	createExtraItemDbg("Sprite", "drink", 2.50)
	createExtraItemDbg("Fanta", "drink", 2.50)
	createExtraItemDbg("Water", "drink", 1.50)
	createExtraItemDbg("Orange Juice", "drink", 3.00)
	createExtraItemDbg("Apple Juice", "drink", 3.00)
	createExtraItemDbg("Iced Tea", "drink", 2.50)
	createExtraItemDbg("Lemonade", "drink", 2.50)
	createExtraItemDbg("Coffee", "drink", 2.00)
	createExtraItemDbg("Espresso", "drink", 2.50)

	// Create some discount codes
	createDiscountCodeDbg("SAVE10", 10)
	createDiscountCodeDbg("SAVE20", 20)
	createDiscountCodeDbg("HALFOFF", 50)
	createDiscountCodeDbg("STUDENT", 15)
	createDiscountCodeDbg("BIRTHDAY", 25) // Special birthday discount - 25% off!

	ingredients, _ := GetAllIngredients()
	log.Println("Ingredients:")
	for _, ingr := range ingredients {
		log.Println(ingr)
	}

	pizzas, _ := GetAllPizzas()
	log.Println("Pizzas:")
	for _, pizza := range pizzas {
		info, _ := GetPizzaInformation(pizza.Name)

		log.Println(pizza, info)
	}

	if err := AddUser("admin", "admin", AdminRole); err != nil {
		log.Fatal(err)
	}

	success, msg := TryAddCustomer(Customer{
		Username:    "walta",
		Password:    "pasword",
		Name:        "Walter White",
		Gender:      "Male",
		BirthDate:   "1958-09-07",
		NoBirthDate: false,
		Address:     "308 Negra Arroyo Lane, Albuquerque, New Mexico ",
		PostCode:    "87104",
	})
	if !success {
		log.Fatal(msg)
	}

	success, msg = TryAddDeliveryPerson(DeliveryPerson{
		Username: "dexta",
		Password: "kill",
		Name:     "Dexter Morgan",
	})
	if !success {
		log.Fatal(msg)
	}

	log.Println("Database has been initialized.")
}

func createIngredientDbg(name string, cost int64, hasMeat bool, hasMeatProducts bool) {
	if _, err := CreateIngredient(NewIngredient(name, cost, hasMeat, hasMeatProducts)); err != nil {
		log.Fatal(err)
	}
}

func createPizzaDbg(name string, ingredients []string) {
	if _, err := CreatePizza(name, ingredients); err != nil {
		log.Fatal(err)
	}
}

func createExtraItemDbg(name string, category string, price float64) {
	query := `INSERT INTO extra_item (name, category, price) VALUES (?, ?, ?)`
	_, err := DATABASE.Exec(query, name, category, price)
	if err != nil {
		log.Fatal(err)
	}
}

func createDiscountCodeDbg(code string, percentage int) {
	query := `INSERT INTO discount_code (code, discount_percentage, is_active) VALUES (?, ?, TRUE)`
	_, err := DATABASE.Exec(query, code, percentage)
	if err != nil {
		log.Fatal(err)
	}
}
