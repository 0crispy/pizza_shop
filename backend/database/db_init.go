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
			diet_type ENUM('OMNIVORE', 'VEGETARIAN', 'VEGAN') NOT NULL
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
		NewIngredient("pepperoni", 101, Omnivore),
		NewIngredient("cheese", 60, Vegetarian),
		NewIngredient("jalapeno", 30, Vegan),
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
