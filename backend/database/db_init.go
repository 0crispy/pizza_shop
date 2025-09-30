package database

import (
	"log"
)

// Used for first-time initialization of the whole database.
func InitDatabaseDev() {
	queries := []string{
		`SET FOREIGN_KEY_CHECKS = 0;`,

		`DROP TABLE IF EXISTS pizza;`,
		`CREATE TABLE pizza (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE
		);`,

		`DROP TABLE IF EXISTS ingredient;`,
		`CREATE TABLE ingredient(
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			cost DECIMAL(10, 2) NOT NULL,
			has_meat BOOLEAN NOT NULL,
			has_animal_products BOOLEAN NOT NULL
		);`,

		`DROP TABLE IF EXISTS pizza_ingredient;`,
		`CREATE TABLE pizza_ingredient(
			pizza_id INT NOT NULL,
			ingredient_id INT NOT NULL,

			PRIMARY KEY (pizza_id, ingredient_id),
			FOREIGN KEY (pizza_id) REFERENCES pizza(id)
				ON DELETE CASCADE,
			FOREIGN KEY (ingredient_id) REFERENCES ingredient(id)
				ON DELETE CASCADE
		);`,

		`DROP TABLE IF EXISTS user;`,
		`CREATE TABLE user(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(256) NOT NULL,
			salt VARCHAR(256) NOT NULL,
			role ENUM('ADMIN', 'DELIVERY', 'CUSTOMER') NOT NULL
		);`,

		`DROP TABLE IF EXISTS customer;`,
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

	log.Println("Database has been initialized.")
}

// InitDatabaseIfEmpty ensures schema exists and seeds only if there is no data.
func InitDatabaseIfEmpty() {
	// Create tables if they don't exist
	queries := []string{
		`SET FOREIGN_KEY_CHECKS = 0;`,
		`CREATE TABLE IF NOT EXISTS pizza (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE
		);`,
		`CREATE TABLE IF NOT EXISTS ingredient(
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			cost DECIMAL(10, 2) NOT NULL,
			has_meat BOOLEAN NOT NULL,
			has_animal_products BOOLEAN NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS pizza_ingredient(
			pizza_id INT NOT NULL,
			ingredient_id INT NOT NULL,
			PRIMARY KEY (pizza_id, ingredient_id),
			FOREIGN KEY (pizza_id) REFERENCES pizza(id)
				ON DELETE CASCADE,
			FOREIGN KEY (ingredient_id) REFERENCES ingredient(id)
				ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS user(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(256) NOT NULL,
			salt VARCHAR(256) NOT NULL,
			role ENUM('ADMIN', 'DELIVERY', 'CUSTOMER') NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS customer(
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
		`SET FOREIGN_KEY_CHECKS = 1;`,
	}
	for _, q := range queries {
		if _, err := DATABASE.Exec(q); err != nil {
			log.Println("Error executing query: ", q)
			log.Fatal(err)
		}
	}

	// Seed only if there is no data
	var count int
	if err := DATABASE.QueryRow("SELECT COUNT(*) FROM ingredient").Scan(&count); err != nil {
		// If table empty or error (e.g., first run), try seeding base data
		log.Println("ingredient count check error:", err)
	}
	if count == 0 {
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
	}

	// Ensure at least one admin user
	var userCount int
	if err := DATABASE.QueryRow("SELECT COUNT(*) FROM user").Scan(&userCount); err != nil {
		log.Println("user count check error:", err)
	}
	if userCount == 0 {
		if err := AddUser("admin", "admin", AdminRole); err != nil {
			log.Fatal(err)
		}
	}

	// Seed pizzas only if none
	var pizzaCount int
	if err := DATABASE.QueryRow("SELECT COUNT(*) FROM pizza").Scan(&pizzaCount); err != nil {
		log.Println("pizza count check error:", err)
	}
	if pizzaCount == 0 {
		createPizzaDbg("Marinara", []string{"Tomato sauce"})
		createPizzaDbg("Margherita", []string{"Tomato sauce", "Mozzarella"})
		createPizzaDbg("Diavola", []string{"Tomato sauce", "Mozzarella", "'Nduja", "Chili pepper"})
		createPizzaDbg("Priosciutto e Funghi", []string{"Tomato sauce", "Mozzarella", "Prosciutto", "Mushrooms"})
		createPizzaDbg("Napoli", []string{"Tomato sauce", "Mozzarella", "Anchovies"})
		createPizzaDbg("Quatro Formaggi", []string{"Tomato sauce", "Mozzarella", "Gorgonzola", "Fontina", "Parmigiano Reggiano"})
		createPizzaDbg("Nutella", []string{"Nutella"})
		createPizzaDbg("Billionaire's dream", []string{"Tomato sauce", "Mozzarella", "Lobster", "Caviar", "Gold leaf", "Foie gras"})
		createPizzaDbg("Student's dream", []string{"Old tomato sauce", "Weird vegan cheese"})
	}

	// Ensure sample customer only if none
	var custCount int
	if err := DATABASE.QueryRow("SELECT COUNT(*) FROM customer").Scan(&custCount); err != nil {
		log.Println("customer count check error:", err)
	}
	if custCount == 0 {
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
	}

	// Log small summary
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
	log.Println("Database checked/initialized.")
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
