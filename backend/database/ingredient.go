package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
)

type Ingredient struct {
	Name              string
	Cost              decimal.Decimal
	HasMeat           bool
	HasAnimalProducts bool
}

func NewIngredient(name string, costCents int64, hasMeat bool, hasAnimalProducts bool) Ingredient {
	return Ingredient{name, decimal.NewFromInt(costCents).Shift(-2), hasMeat, hasAnimalProducts}
}

type IngredientWithID struct {
	ID   int
	Ingr Ingredient
}

func CreateIngredient(ingr Ingredient) (IngredientWithID, error) {
	res, err := DATABASE.Exec("INSERT INTO ingredient (name, cost, has_meat, has_animal_products) VALUES (?, ?, ?, ?)", ingr.Name, ingr.Cost.String(), ingr.HasMeat, ingr.HasAnimalProducts)
	if err != nil {
		return IngredientWithID{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return IngredientWithID{}, err
	}

	return IngredientWithID{
		ID: int(id),
		Ingr: Ingredient{
			Name:              ingr.Name,
			Cost:              ingr.Cost,
			HasMeat:           ingr.HasMeat,
			HasAnimalProducts: ingr.HasAnimalProducts,
		},
	}, nil
}

func GetAllIngredients() ([]IngredientWithID, error) {
	rows, err := DATABASE.Query("SELECT id, name, cost, has_meat, has_animal_products FROM ingredient")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []IngredientWithID

	for rows.Next() {
		var ingr IngredientWithID
		var ingr_cost_str string

		err := rows.Scan(&ingr.ID, &ingr.Ingr.Name, &ingr_cost_str, &ingr.Ingr.HasMeat, &ingr.Ingr.HasAnimalProducts)
		if err != nil {
			log.Println(err)
			continue
		}

		ingr.Ingr.Cost, err = decimal.NewFromString(ingr_cost_str)
		if err != nil {
			log.Println("Invalid decimal from database: ", ingr_cost_str)
			continue
		}

		ingredients = append(ingredients, ingr)
	}

	return ingredients, nil
}

func GetIngredient(ingredientName string) (IngredientWithID, error) {
	var ingr IngredientWithID
	var ingr_cost_str string

	err := DATABASE.QueryRow(
		"SELECT id, name, cost, has_meat, has_animal_products FROM ingredient WHERE name = ?",
		ingredientName,
	).Scan(&ingr.ID, &ingr.Ingr.Name, &ingr_cost_str, &ingr.Ingr.HasMeat, &ingr.Ingr.HasAnimalProducts)

	if err != nil {
		if err == sql.ErrNoRows {
			return IngredientWithID{}, fmt.Errorf("ingredient not found: %s", ingredientName)
		}
		return IngredientWithID{}, err
	}

	ingr.Ingr.Cost, err = decimal.NewFromString(ingr_cost_str)
	if err != nil {
		return IngredientWithID{}, fmt.Errorf("invalid decimal from database: %s", ingr_cost_str)
	}

	return ingr, nil
}
