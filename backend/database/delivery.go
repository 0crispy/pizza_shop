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

// Admin functions for delivery person management

func GetAllDeliveryPersons() ([]map[string]interface{}, error) {
	query := `
		SELECT u.id, u.username, dp.name
		FROM user u
		INNER JOIN delivery_person dp ON u.id = dp.user_id
		WHERE u.role = 'DELIVERY'
		ORDER BY dp.name
	`
	rows, err := DATABASE.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveryPersons []map[string]interface{}
	for rows.Next() {
		var id int
		var username, name string
		err := rows.Scan(&id, &username, &name)
		if err != nil {
			return nil, err
		}

		person := map[string]interface{}{
			"id":       id,
			"username": username,
			"name":     name,
		}
		deliveryPersons = append(deliveryPersons, person)
	}

	return deliveryPersons, nil
}

func DeleteDeliveryPerson(userID int) error {
	tx, err := DATABASE.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete from delivery_person table
	_, err = tx.Exec("DELETE FROM delivery_person WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Delete from user table
	_, err = tx.Exec("DELETE FROM user WHERE id = ? AND role = 'DELIVERY'", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
