package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Order struct {
	ID                 int       `json:"id"`
	CustomerID         int       `json:"customer_id"`
	CustomerName       string    `json:"customer_name"`
	Timestamp          time.Time `json:"timestamp"`
	Status             string    `json:"status"`
	PostalCode         string    `json:"postal_code"`
	DeliveryAddress    string    `json:"delivery_address"`
	DiscountCodeID     *int      `json:"discount_code_id"`
	DiscountCode       *string   `json:"discount_code"`
	DiscountPercentage *int      `json:"discount_percentage"`
	DeliveryPersonID   *int      `json:"delivery_person_id"`
	DeliveryPersonName *string   `json:"delivery_person_name"`
}

type OrderPizza struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	PizzaID   int     `json:"pizza_id"`
	PizzaName string  `json:"pizza_name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
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

func CreateOrderWithTransaction(customerID int, userID int, deliveryAddress, postalCode string, pizzaItems []struct {
	PizzaID  int
	Quantity int
}, extraItems []struct {
	ExtraItemID int
	Quantity    int
}, discountCode *string) (int, error) {
	tx, err := DATABASE.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get discount code ID if provided
	var discountCodeID *int
	var isBirthdayDiscount bool
	if discountCode != nil && *discountCode != "" {
		var id int
		var isActive bool
		var code string
		err = tx.QueryRow(`SELECT id, code, is_active FROM discount_code WHERE code = ?`, *discountCode).Scan(&id, &code, &isActive)
		if err == nil && isActive {
			discountCodeID = &id
			isBirthdayDiscount = (code == "BIRTHDAY")

			// Check if user already used this discount
			var usageCount int
			err = tx.QueryRow(`SELECT COUNT(*) FROM discount_usage WHERE user_id = ? AND discount_code_id = ?`, userID, id).Scan(&usageCount)
			if err != nil {
				return 0, err
			}
			if usageCount > 0 {
				return 0, fmt.Errorf("discount code already used")
			}
		}
		// If discount code not found or inactive, we just ignore it (don't fail the order)
	}

	// Handle birthday discount: free cheapest pizza + 1 free drink
	if isBirthdayDiscount {
		// Find cheapest pizza in the order
		if len(pizzaItems) > 0 {
			cheapestIdx := -1
			var cheapestPrice float64

			for i, item := range pizzaItems {
				var price float64
				err = tx.QueryRow(`
					SELECT ROUND(SUM(i.cost) * 1.4 * 1.09, 4)
					FROM pizza_ingredient pi
					JOIN ingredient i ON pi.ingredient_id = i.id
					WHERE pi.pizza_id = ?
				`, item.PizzaID).Scan(&price)

				if err == nil {
					if cheapestIdx == -1 || price < cheapestPrice {
						cheapestIdx = i
						cheapestPrice = price
					}
				}
			}

			// Remove one quantity from cheapest pizza (make it free)
			if cheapestIdx != -1 && pizzaItems[cheapestIdx].Quantity > 0 {
				pizzaItems[cheapestIdx].Quantity--
				// If quantity becomes 0, we'll skip it when inserting
			}
		}

		// Add 1 free drink (find cheapest drink)
		var cheapestDrinkID int
		err = tx.QueryRow(`
			SELECT id FROM extra_item 
			WHERE category = 'drink' 
			ORDER BY price ASC 
			LIMIT 1
		`).Scan(&cheapestDrinkID)

		if err == nil {
			// Add free drink to extras
			extraItems = append(extraItems, struct {
				ExtraItemID int
				Quantity    int
			}{
				ExtraItemID: cheapestDrinkID,
				Quantity:    1,
			})
		}
	}

	query := `
		INSERT INTO orders (customer_id, delivery_address, postal_code, status, timestamp, discount_code_id)
		VALUES (?, ?, ?, 'IN_PROGRESS', NOW(), ?)
	`
	result, err := tx.Exec(query, customerID, deliveryAddress, postalCode, discountCodeID)
	if err != nil {
		return 0, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	pizzaQuery := `INSERT INTO order_pizza (order_id, pizza_id, quantity) VALUES (?, ?, ?)`
	for _, item := range pizzaItems {
		if item.Quantity > 0 { // Only insert if quantity > 0
			_, err = tx.Exec(pizzaQuery, orderID, item.PizzaID, item.Quantity)
			if err != nil {
				return 0, err
			}
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

	// Record discount usage
	if discountCodeID != nil {
		_, err = tx.Exec(`INSERT INTO discount_usage (user_id, discount_code_id, used_at) VALUES (?, ?, NOW())`, userID, *discountCodeID)
		if err != nil {
			return 0, err
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

		info, err := GetPizzaInformation(op.PizzaName)
		if err != nil {
			return nil, err
		}
		priceFloat, _ := info.Cost.Float64()
		op.Price = priceFloat

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
		SELECT o.id, o.customer_id, c.name as customer_name, o.timestamp, o.status, o.postal_code, o.delivery_address,
		       o.discount_code_id, dc.code, dc.discount_percentage, o.delivery_person_id, dp.name as delivery_person_name
		FROM orders o
		LEFT JOIN customer c ON o.customer_id = c.id
		LEFT JOIN discount_code dc ON o.discount_code_id = dc.id
		LEFT JOIN delivery_person dp ON o.delivery_person_id = dp.id
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
		var customerName, discountCode, deliveryPersonName sql.NullString
		var discountCodeID, discountPercentage, deliveryPersonID sql.NullInt64

		err := rows.Scan(&order.ID, &order.CustomerID, &customerName, &order.Timestamp, &order.Status,
			&order.PostalCode, &order.DeliveryAddress, &discountCodeID, &discountCode, &discountPercentage,
			&deliveryPersonID, &deliveryPersonName)
		if err != nil {
			return nil, err
		}

		if customerName.Valid {
			order.CustomerName = customerName.String
		} else {
			order.CustomerName = "Unknown"
		}

		if discountCodeID.Valid {
			id := int(discountCodeID.Int64)
			order.DiscountCodeID = &id
		}
		if discountCode.Valid {
			code := discountCode.String
			order.DiscountCode = &code
		}
		if discountPercentage.Valid {
			pct := int(discountPercentage.Int64)
			order.DiscountPercentage = &pct
		}
		if deliveryPersonID.Valid {
			dpID := int(deliveryPersonID.Int64)
			order.DeliveryPersonID = &dpID
		}
		if deliveryPersonName.Valid {
			dpName := deliveryPersonName.String
			order.DeliveryPersonName = &dpName
		}

		orders = append(orders, order)
	}

	return orders, nil
}
