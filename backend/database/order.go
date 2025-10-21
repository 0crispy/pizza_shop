package database

import (
	"database/sql"
	"time"
)

type Order struct {
	ID              int       `json:"id"`
	CustomerID      int       `json:"customer_id"`
	CustomerName    string    `json:"customer_name"`
	Timestamp       time.Time `json:"timestamp"`
	Status          string    `json:"status"`
	PostalCode      string    `json:"postal_code"`
	DeliveryAddress string    `json:"delivery_address"`
}

type OrderPizza struct {
	ID        int    `json:"id"`
	OrderID   int    `json:"order_id"`
	PizzaID   int    `json:"pizza_id"`
	PizzaName string `json:"pizza_name"`
	Quantity  int    `json:"quantity"`
}

type OrderExtraItem struct {
	ID            int     `json:"id"`
	OrderID       int     `json:"order_id"`
	ExtraItemID   int     `json:"extra_item_id"`
	ExtraItemName string  `json:"extra_item_name"`
	Category      string  `json:"category"`
	Price         float64 `json:"price"`
	Quantity      int     `json:"quantity"`
}

type OrderDetails struct {
	Order      Order            `json:"order"`
	Pizzas     []OrderPizza     `json:"pizzas"`
	ExtraItems []OrderExtraItem `json:"extra_items"`
	TotalPrice float64          `json:"total_price"`
}

func CreateOrderWithTransaction(customerID int, deliveryAddress, postalCode string, pizzaItems []struct {
	PizzaID  int
	Quantity int
}, extraItems []struct {
	ExtraItemID int
	Quantity    int
}) (int, error) {
	tx, err := DATABASE.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO orders (customer_id, delivery_address, postal_code, status, timestamp)
		VALUES (?, ?, ?, 'IN_PROGRESS', NOW())
	`
	result, err := tx.Exec(query, customerID, deliveryAddress, postalCode)
	if err != nil {
		return 0, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	pizzaQuery := `INSERT INTO order_pizza (order_id, pizza_id, quantity) VALUES (?, ?, ?)`
	for _, item := range pizzaItems {
		_, err = tx.Exec(pizzaQuery, orderID, item.PizzaID, item.Quantity)
		if err != nil {
			return 0, err
		}
	}

	if len(extraItems) > 0 {
		extraQuery := `INSERT INTO order_extra_item (order_id, extra_item_id, quantity) VALUES (?, ?, ?)`
		for _, item := range extraItems {
			_, err = tx.Exec(extraQuery, orderID, item.ExtraItemID, item.Quantity)
			if err != nil {
				return 0, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return int(orderID), nil
}

func CreateOrder(customerID int, deliveryAddress, postalCode string) (int, error) {
	query := `
		INSERT INTO ` + "`order`" + ` (customer_id, delivery_address, postal_code, status, timestamp)
		VALUES (?, ?, ?, 'pending', NOW())
	`
	result, err := DATABASE.Exec(query, customerID, deliveryAddress, postalCode)
	if err != nil {
		return 0, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(orderID), nil
}

func AddPizzaToOrder(orderID, pizzaID, quantity int) error {
	query := `INSERT INTO order_pizza (order_id, pizza_id, quantity) VALUES (?, ?, ?)`
	_, err := DATABASE.Exec(query, orderID, pizzaID, quantity)
	return err
}

func AddExtraItemToOrder(orderID, extraItemID, quantity int) error {
	query := `INSERT INTO order_extra_item (order_id, extra_item_id, quantity) VALUES (?, ?, ?)`
	_, err := DATABASE.Exec(query, orderID, extraItemID, quantity)
	return err
}

func GetOrdersByCustomer(customerID int) ([]Order, error) {
	query := `
		SELECT id, customer_id, timestamp, status, postal_code, delivery_address
		FROM orders
		WHERE customer_id = ?
		ORDER BY timestamp DESC
	`
	rows, err := DATABASE.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.Timestamp, &order.Status, &order.PostalCode, &order.DeliveryAddress)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func GetOrderByID(orderID int) (*Order, error) {
	var order Order
	query := `
		SELECT id, customer_id, timestamp, status, postal_code, delivery_address
		FROM orders
		WHERE id = ?
	`
	err := DATABASE.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Timestamp,
		&order.Status,
		&order.PostalCode,
		&order.DeliveryAddress,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func GetOrderDetails(orderID int) (*OrderDetails, error) {
	var details OrderDetails

	query := `
		SELECT o.id, o.customer_id, c.name, o.timestamp, o.status, o.postal_code, o.delivery_address
		FROM orders o
		LEFT JOIN customer c ON o.customer_id = c.id
		WHERE o.id = ?
	`
	var customerName sql.NullString
	err := DATABASE.QueryRow(query, orderID).Scan(
		&details.Order.ID,
		&details.Order.CustomerID,
		&customerName,
		&details.Order.Timestamp,
		&details.Order.Status,
		&details.Order.PostalCode,
		&details.Order.DeliveryAddress,
	)
	if err != nil {
		return nil, err
	}

	if customerName.Valid {
		details.Order.CustomerName = customerName.String
	} else {
		details.Order.CustomerName = "Unknown"
	}

	pizzaQuery := `
		SELECT op.id, op.order_id, op.pizza_id, p.name, op.quantity 
		FROM order_pizza op
		JOIN pizza p ON op.pizza_id = p.id
		WHERE op.order_id = ?
	`
	pizzaRows, err := DATABASE.Query(pizzaQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer pizzaRows.Close()

	for pizzaRows.Next() {
		var op OrderPizza
		err := pizzaRows.Scan(&op.ID, &op.OrderID, &op.PizzaID, &op.PizzaName, &op.Quantity)
		if err != nil {
			return nil, err
		}
		details.Pizzas = append(details.Pizzas, op)
	}

	extraQuery := `
		SELECT oei.id, oei.order_id, oei.extra_item_id, ei.name, ei.category, ei.price, oei.quantity 
		FROM order_extra_item oei
		JOIN extra_item ei ON oei.extra_item_id = ei.id
		WHERE oei.order_id = ?
	`
	extraRows, err := DATABASE.Query(extraQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer extraRows.Close()

	for extraRows.Next() {
		var oe OrderExtraItem
		err := extraRows.Scan(&oe.ID, &oe.OrderID, &oe.ExtraItemID, &oe.ExtraItemName, &oe.Category, &oe.Price, &oe.Quantity)
		if err != nil {
			return nil, err
		}
		details.ExtraItems = append(details.ExtraItems, oe)
	}

	details.TotalPrice, err = calculateOrderTotal(&details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

func calculateOrderTotal(details *OrderDetails) (float64, error) {
	total := 0.0

	for _, op := range details.Pizzas {
		pizza, err := GetPizzaByID(op.PizzaID)
		if err != nil {
			return 0, err
		}
		info, err := GetPizzaInformation(pizza.Name)
		if err != nil {
			return 0, err
		}
		priceFloat, _ := info.Cost.Float64()
		total += priceFloat * float64(op.Quantity)
	}

	for _, oe := range details.ExtraItems {
		total += oe.Price * float64(oe.Quantity)
	}

	return total, nil
}

func UpdateOrderStatus(orderID int, status string) error {
	query := `UPDATE orders SET status = ? WHERE id = ?`
	_, err := DATABASE.Exec(query, status, orderID)
	return err
}

func DeleteOrder(orderID int) error {
	query := `DELETE FROM orders WHERE id = ?`
	_, err := DATABASE.Exec(query, orderID)
	return err
}

func GetAllOrders() ([]Order, error) {
	query := `
		SELECT o.id, o.customer_id, c.name as customer_name, o.timestamp, o.status, o.postal_code, o.delivery_address
		FROM orders o
		LEFT JOIN customer c ON o.customer_id = c.id
		ORDER BY o.timestamp DESC
	`
	rows, err := DATABASE.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		var customerName sql.NullString
		err := rows.Scan(&order.ID, &order.CustomerID, &customerName, &order.Timestamp, &order.Status, &order.PostalCode, &order.DeliveryAddress)
		if err != nil {
			return nil, err
		}
		if customerName.Valid {
			order.CustomerName = customerName.String
		} else {
			order.CustomerName = "Unknown"
		}
		orders = append(orders, order)
	}

	return orders, nil
}
