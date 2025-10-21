package database

import (
	"errors"
	"log"
)

var (
	ErrOrderNotAvailable    = errors.New("order is not available")
	ErrOrderAlreadyAssigned = errors.New("order is already assigned")
	ErrInvalidStatus        = errors.New("invalid status")
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

func GetDeliveryPersonIDFromUserID(userID int) (int, error) {
	var id int
	err := DATABASE.QueryRow("SELECT id FROM delivery_person WHERE user_id = ?", userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetAvailableDeliveries() ([]OrderDetails, error) {
	query := `
		SELECT o.id
		FROM orders o
		WHERE o.status = 'IN_PROGRESS'
		AND NOT EXISTS (
			SELECT 1 
			FROM delivery_assignment da 
			WHERE da.order_id = o.id
		)
		ORDER BY o.timestamp ASC
	`
	rows, err := DATABASE.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderDetails
	for rows.Next() {
		var orderID int
		err := rows.Scan(&orderID)
		if err != nil {
			return nil, err
		}

		orderDetails, err := GetOrderDetails(orderID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *orderDetails)
	}
	return orders, nil
}

func GetAssignedDeliveries(deliveryPersonID int) ([]OrderDetails, error) {
	query := `
		SELECT o.id
		FROM orders o
		JOIN delivery_assignment da ON o.id = da.order_id
		WHERE da.delivery_person_id = ?
		AND o.status IN ('IN_PROGRESS', 'OUT_FOR_DELIVERY')
		ORDER BY o.timestamp ASC
	`
	rows, err := DATABASE.Query(query, deliveryPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderDetails
	for rows.Next() {
		var orderID int
		err := rows.Scan(&orderID)
		if err != nil {
			return nil, err
		}

		orderDetails, err := GetOrderDetails(orderID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *orderDetails)
	}
	return orders, nil
}

func AssignDelivery(orderID, deliveryPersonID int) error {
	tx, err := DATABASE.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if order is available
	var status string
	err = tx.QueryRow("SELECT status FROM orders WHERE id = ?", orderID).Scan(&status)
	if err != nil {
		return err
	}
	if status != "IN_PROGRESS" {
		return ErrOrderNotAvailable
	}

	// Check if order is already assigned
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM delivery_assignment WHERE order_id = ?", orderID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrOrderAlreadyAssigned
	}

	// Assign delivery
	_, err = tx.Exec(
		"INSERT INTO delivery_assignment (order_id, delivery_person_id) VALUES (?, ?)",
		orderID,
		deliveryPersonID,
	)
	if err != nil {
		return err
	}

	// Update order status
	_, err = tx.Exec(
		"UPDATE orders SET status = 'OUT_FOR_DELIVERY' WHERE id = ?",
		orderID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func UpdateDeliveryStatus(orderID int, status string) error {
	if status != "DELIVERED" && status != "FAILED" {
		return ErrInvalidStatus
	}

	_, err := DATABASE.Exec(
		"UPDATE orders SET status = ? WHERE id = ?",
		status,
		orderID,
	)
	return err
}
