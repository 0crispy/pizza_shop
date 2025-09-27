package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Pizza struct {
	ID          int
	Name        string
	Ingredients []IngredientWithID
}

func (p Pizza) String() string {
	var ingredientNames []string
	for _, ingr := range p.Ingredients {
		ingredientNames = append(ingredientNames, ingr.Ingr.Name)
	}
	return fmt.Sprintf("Pizza(Name=%s, [%s])", p.Name, strings.Join(ingredientNames, ", "))
}

func CreatePizza(pizzaName string, ingredientNames []string) (Pizza, error) {
	ingredients := []IngredientWithID{}
	for _, ingrName := range ingredientNames {
		ingr, err := GetIngredient(ingrName)
		if err != nil {
			return Pizza{}, err
		}
		ingredients = append(ingredients, ingr)
	}

	pizza := Pizza{Name: pizzaName, Ingredients: ingredients}
	// use transaction to not fuck up the database
	tx, err := DATABASE.Begin()
	if err != nil {
		return Pizza{}, err
	}

	res, err := tx.Exec("INSERT INTO pizza (name) VALUES (?)", pizzaName)
	if err != nil {
		tx.Rollback()
		return Pizza{}, err
	}

	pizzaID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return Pizza{}, err
	}
	pizza.ID = int(pizzaID)

	stmt, err := tx.Prepare("INSERT INTO pizza_ingredient (pizza_id, ingredient_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return Pizza{}, err
	}
	defer stmt.Close()

	for _, ingr := range ingredients {
		_, err := stmt.Exec(pizza.ID, ingr.ID)
		if err != nil {
			tx.Rollback()
			return Pizza{}, fmt.Errorf("failed to add ingredient %d: %w", ingr.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Pizza{}, err
	}

	return pizza, nil
}

func GetAllPizzas() ([]Pizza, error) {
	rows, err := DATABASE.Query("SELECT id, name FROM pizza")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pizzas []Pizza
	for rows.Next() {
		var pizza Pizza
		err := rows.Scan(&pizza.ID, &pizza.Name)
		if err != nil {
			fmt.Println("Failed to scan pizza:", err)
			continue
		}
		ingrRows, err := DATABASE.Query(`
			SELECT i.id, i.name, i.cost, i.has_meat, i.has_animal_products
			FROM ingredient i
			JOIN pizza_ingredient pi ON pi.ingredient_id = i.id
			WHERE pi.pizza_id = ?`, pizza.ID,
		)

		if err != nil {
			fmt.Println("Failed to query ingredients:", err)
			continue
		}

		var ingredients []IngredientWithID
		for ingrRows.Next() {
			var ingr IngredientWithID
			var costStr string
			err := ingrRows.Scan(&ingr.ID, &ingr.Ingr.Name, &costStr, &ingr.Ingr.HasMeat, &ingr.Ingr.HasAnimalProducts)
			if err != nil {
				fmt.Println("Failed to scan ingredient:", err)
				continue
			}

			ingr.Ingr.Cost, err = decimal.NewFromString(costStr)
			if err != nil {
				fmt.Println("Invalid decimal from database:", costStr)
				continue
			}

			ingredients = append(ingredients, ingr)
		}
		ingrRows.Close()

		pizza.Ingredients = ingredients
		pizzas = append(pizzas, pizza)
	}

	return pizzas, nil
}

type PizzaInformation struct {
	Cost         decimal.Decimal
	IsVegetarian bool
	IsVegan      bool
}

func GetPizzaInformation(pizzaName string) (PizzaInformation, error) {
	var pizzaID int

	err := DATABASE.QueryRow("SELECT id FROM pizza WHERE name = ?", pizzaName).Scan(&pizzaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return PizzaInformation{}, fmt.Errorf("pizza not found: %s", pizzaName)
		}
		return PizzaInformation{}, err
	}

	rows, err := DATABASE.Query(`
		SELECT cost, has_meat, has_animal_products
		FROM ingredient i
		JOIN pizza_ingredient pi ON pi.ingredient_id = i.id
		WHERE pi.pizza_id = ?
	`, pizzaID)
	if err != nil {
		return PizzaInformation{}, err
	}
	defer rows.Close()

	ingredientsCost := getPizzaDoughCost()
	isVegetarian := true
	isVegan := true

	for rows.Next() {
		var costStr string
		var hasMeat bool
		var hasAnimalProducts bool

		err := rows.Scan(&costStr, &hasMeat, &hasAnimalProducts)
		if err != nil {
			return PizzaInformation{}, err
		}

		cost, err := decimal.NewFromString(costStr)
		if err != nil {
			return PizzaInformation{}, fmt.Errorf("invalid cost in database: %s", costStr)
		}

		ingredientsCost = ingredientsCost.Add(cost)

		if hasMeat {
			isVegetarian = false
			isVegan = false
		} else if hasAnimalProducts {
			isVegan = false
		}
	}

	totalCost := getPizzaFinalCost(ingredientsCost)

	return PizzaInformation{
		Cost:         totalCost,
		IsVegetarian: isVegetarian,
		IsVegan:      isVegan,
	}, nil
}

func getPizzaDoughCost() decimal.Decimal {
	return decimal.NewFromFloat(5.0)
}

func getPizzaFinalCost(ingredientsCost decimal.Decimal) decimal.Decimal {
	// Add 40% margin :)
	totalCost := ingredientsCost.Mul(decimal.NewFromFloat(1.4))
	// Add 9% VAT (stupid)
	totalCost = totalCost.Mul(decimal.NewFromFloat(1.09))
	return totalCost
}
