package database

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

var (
	ErrOrderNotAvailable         = errors.New("order is not available")
	ErrOrderAlreadyAssigned      = errors.New("order is already assigned")
	ErrDeliveryPersonUnavailable = errors.New("delivery person is unavailable")
	ErrDeliveryPersonBusy        = errors.New("delivery person already has an active delivery")
	ErrInvalidStatus             = errors.New("invalid status")
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
		SELECT u.id, u.username, dp.name, dp.vehicle_type
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
		var username, name, vehicleType string
		err := rows.Scan(&id, &username, &name, &vehicleType)
		if err != nil {
			return nil, err
		}

		person := map[string]interface{}{
			"id":           id,
			"username":     username,
			"name":         name,
			"vehicle_type": vehicleType,
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

// IsDeliveryPersonAvailable returns true if delivery person is not in cooldown and has no active deliveries
func IsDeliveryPersonAvailable(deliveryPersonID int) (bool, error) {
	var unavailableUntil sql.NullTime
	err := DATABASE.QueryRow("SELECT unavailable_until FROM delivery_person WHERE id = ?", deliveryPersonID).Scan(&unavailableUntil)
	if err != nil {
		return false, err
	}
	if unavailableUntil.Valid {
		var now time.Time
		if err := DATABASE.QueryRow("SELECT NOW()").Scan(&now); err == nil {
			if unavailableUntil.Time.After(now) {
				return false, nil
			}
		}
	}
	var activeCount int
	err = DATABASE.QueryRow("SELECT COUNT(*) FROM orders WHERE delivery_person_id = ? AND status IN ('IN_PROGRESS','OUT_FOR_DELIVERY')", deliveryPersonID).Scan(&activeCount)
	if err != nil {
		return false, err
	}
	if activeCount > 0 {
		return false, nil
	}
	return true, nil
}

func GetAvailableDeliveries() ([]Order, error) {
	query := `
		SELECT o.id, o.customer_id, c.name, o.timestamp, o.status, o.postal_code, o.delivery_address
		FROM orders o
		JOIN customer c ON o.customer_id = c.id
		WHERE o.status = 'IN_PROGRESS'
		AND o.delivery_person_id IS NULL
		ORDER BY o.timestamp ASC
	`
	rows, err := DATABASE.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.CustomerName, &order.Timestamp, &order.Status, &order.PostalCode, &order.DeliveryAddress)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func GetAssignedDeliveries(deliveryPersonID int) ([]Order, error) {
	query := `
		SELECT o.id, o.customer_id, c.name, o.timestamp, o.status, o.postal_code, o.delivery_address
		FROM orders o
		JOIN customer c ON o.customer_id = c.id
		WHERE o.delivery_person_id = ?
		AND o.status IN ('IN_PROGRESS', 'OUT_FOR_DELIVERY', 'DELIVERED', 'FAILED')
		ORDER BY o.timestamp DESC
	`
	rows, err := DATABASE.Query(query, deliveryPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.CustomerName, &order.Timestamp, &order.Status, &order.PostalCode, &order.DeliveryAddress)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
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

	// Check if order is already assigned (orders.delivery_person_id)
	var existingAssigned sql.NullInt64
	err = tx.QueryRow("SELECT delivery_person_id FROM orders WHERE id = ?", orderID).Scan(&existingAssigned)
	if err != nil {
		return err
	}
	if existingAssigned.Valid {
		return ErrOrderAlreadyAssigned
	}

	// Check if delivery person is currently unavailable or already has an active assignment
	var unavailableUntil sql.NullTime
	err = tx.QueryRow("SELECT unavailable_until FROM delivery_person WHERE id = ?", deliveryPersonID).Scan(&unavailableUntil)
	if err != nil {
		return err
	}
	if unavailableUntil.Valid {
		// if unavailable_until > now, they are not available
		var now time.Time
		err = tx.QueryRow("SELECT NOW()").Scan(&now)
		if err == nil && unavailableUntil.Time.After(now) {
			return ErrDeliveryPersonUnavailable
		}
	}

	// Check if delivery person already has an IN_PROGRESS or OUT_FOR_DELIVERY order assigned
	var activeCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM orders WHERE delivery_person_id = ? AND status IN ('IN_PROGRESS','OUT_FOR_DELIVERY')", deliveryPersonID).Scan(&activeCount)
	if err != nil {
		return err
	}
	if activeCount > 0 {
		return ErrDeliveryPersonBusy
	}

	// Assign delivery by updating orders.delivery_person_id and status
	_, err = tx.Exec(
		"UPDATE orders SET delivery_person_id = ?, status = 'OUT_FOR_DELIVERY' WHERE id = ?",
		deliveryPersonID,
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

	tx, err := DATABASE.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update order status
	_, err = tx.Exec("UPDATE orders SET status = ? WHERE id = ?", status, orderID)
	if err != nil {
		return err
	}

	// If delivered, set delivery_person.unavailable_until = NOW() + 30 minutes
	if status == "DELIVERED" {
		// find delivery_person_id for this order
		var dpID sql.NullInt64
		err = tx.QueryRow("SELECT delivery_person_id FROM orders WHERE id = ?", orderID).Scan(&dpID)
		if err != nil {
			return err
		}
		if dpID.Valid {
			_, err = tx.Exec("UPDATE delivery_person SET unavailable_until = DATE_ADD(NOW(), INTERVAL 30 MINUTE) WHERE id = ?", dpID.Int64)
			if err != nil {
				return err
			}
		}
	}

	// If failed, clear unavailable_until so they can be available immediately
	if status == "FAILED" {
		var dpID sql.NullInt64
		err = tx.QueryRow("SELECT delivery_person_id FROM orders WHERE id = ?", orderID).Scan(&dpID)
		if err != nil {
			return err
		}
		if dpID.Valid {
			_, err = tx.Exec("UPDATE delivery_person SET unavailable_until = NULL WHERE id = ?", dpID.Int64)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
