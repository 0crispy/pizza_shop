package database

import "log"

// Used for first-time initialization of the whole database.
func InitDatabaseDev() {
	queries := []string{
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

		`DROP TABLE IF EXISTS user;`,
		`CREATE TABLE user(
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(128) NOT NULL,
			role ENUM('ADMIN', 'DELIVERY', 'CUSTOMER') NOT NULL
		);`,
	}

	for _, query := range queries {
		_, err := DATABASE.Exec(query)
		if err != nil {
			log.Println("Error executing query: ", query)
			log.Fatal(err)
		}
	}

	create_ingredients := []Ingredient{
		NewIngredient("pepperoni", 101, true, true),
		NewIngredient("cheese", 60, false, true),
		NewIngredient("jalapeno", 30, false, false),
	}

	for _, ingr := range create_ingredients {
		if err := CreateIngredient(ingr); err != nil {
			log.Fatal(err)
		}
	}

	ingredients, err := GetAllIngredients()
	if err != nil {
		log.Fatal(err)
	}
	for _, ingr := range ingredients {
		log.Println(ingr)
	}

	log.Println("Database has been initialized.")
}
