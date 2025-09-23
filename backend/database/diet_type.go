package database

import (
	"database/sql/driver"
	"fmt"
)

type DietType int

const (
	Omnivore DietType = iota
	Vegetarian
	Vegan
)

func (d DietType) String() string {
	switch d {
	case Omnivore:
		return "OMNIVORE"
	case Vegetarian:
		return "VEGETARIAN"
	case Vegan:
		return "VEGAN"
	}
	panic(fmt.Sprintf("unhandled DietType: %d", d))
}

func (d DietType) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d *DietType) Scan(value any) error {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("failed to scan DietType: %v", value)
	}

	switch str {
	case "OMNIVORE":
		*d = Omnivore
	case "VEGETARIAN":
		*d = Vegetarian
	case "VEGAN":
		*d = Vegan
	default:
		return fmt.Errorf("invalid DietType: %s", str)
	}
	return nil
}
