package database

import (
	"log"
)

type DeliveryPerson struct {
	Username string
	Password string
	Name     string
}

func TryAddDeliveryPerson(
	delivery_person DeliveryPerson,
) (bool, string) {

	if len(delivery_person.Username) == 0 {
		return false, "Username cannot be empty!"
	}
	if len(delivery_person.Username) > 100 {
		return false, "Username is too long! (max 100 characters)"
	}
	if len(delivery_person.Password) == 0 {
		return false, "Password cannot be empty!"
	}
	if len(delivery_person.Name) == 0 {
		return false, "Name cannot be empty!"
	}
	if len(delivery_person.Name) > 100 {
		return false, "Name is too long! (max 100 characters)"
	}

	if exists, _ := doesUserExist(delivery_person.Username); exists {
		return false, "Username is already taken!"
	}

	err := AddUser(delivery_person.Username, delivery_person.Password, DeliveryRole)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again."
	}

	userID, err := getUserIDFromUsername(delivery_person.Username)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again."
	}

	_, err = DATABASE.Exec(
		"INSERT INTO delivery_person (user_id, name) VALUES (?, ?)",
		userID,
		delivery_person.Name,
	)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again."
	}
	return true, ""
}
