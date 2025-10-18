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

		`DROP TABLE IF EXISTS orders;`,
		`CREATE TABLE orders(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			customer_id BIGINT NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			status ENUM('IN_PROGRESS', 'DELIVERED', 'FAILED') NOT NULL,
			postal_code VARCHAR(10) NOT NULL,
			delivery_address VARCHAR(256) NOT NULL,

			FOREIGN KEY (customer_id) REFERENCES customer(id)
		);`,

		`DROP TABLE IF EXISTS order_pizza;`,
		`CREATE TABLE order_pizza (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			order_id BIGINT NOT NULL,
			pizza_id INT NOT NULL,
			quantity INT NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (pizza_id) REFERENCES pizza(id)
		)`,

		`DROP TABLE IF EXISTS delivery_person;`,
		`CREATE TABLE delivery_person(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			order_id BIGINT DEFAULT NULL,
			user_id BIGINT NOT NULL,

			FOREIGN KEY (order_id) REFERENCES orders(id),
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
