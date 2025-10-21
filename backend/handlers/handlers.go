package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	database "pizza_shop/backend/database"
	"sort"
	"strings"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/home.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		LoginPostHandler(w, r)
	case http.MethodGet:
		LoginGetHandler(w, r)
	}
}

func LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/login.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {

	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username != "" && password != "" {
		ok, _ := database.TryLogin(username, password)
		if !ok {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		role, err := database.GetUserRole(username)
		if err != nil || role != database.AdminRole.String() {
			http.Error(w, "Not an admin user", http.StatusForbidden)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "X-Username", Value: username, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "X-Password", Value: password, Path: "/"})
	} else {
		userCookie, _ := r.Cookie("X-Username")
		passCookie, _ := r.Cookie("X-Password")
		if userCookie == nil || passCookie == nil {
			html := `<html><head><title>Admin Login</title></head><body><center>
<h1>Admin Panel Login</h1>
<form method="GET" action="/admin">
<table>
<tr><td><b>Username:</b></td><td><input type="text" name="username" required></td></tr>
<tr><td><b>Password:</b></td><td><input type="password" name="password" required></td></tr>
<tr><td colspan="2"><input type="submit" value="Login"></td></tr>
</table>
</form>
</center></body></html>`
			fmt.Fprint(w, html)
			return
		}
		username = userCookie.Value
		password = passCookie.Value

		// Verify cookie credentials
		ok, _ := database.TryLogin(username, password)
		if !ok {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		role, err := database.GetUserRole(username)
		if err != nil || role != database.AdminRole.String() {
			http.Error(w, "Not an admin user", http.StatusForbidden)
			return
		}
	}
	users, _ := database.GetAllUsers()
	orders, _ := database.GetAllOrders()
	deliveryPersons, _ := database.GetAllDeliveryPersons()
	pizzas, _ := database.GetAllPizzasWithPrice()
	ingredients, _ := database.GetAllIngredients()

	extraItems := []map[string]interface{}{}
	rows, err := database.DATABASE.Query("SELECT id, name, category, price FROM extra_item ORDER BY category, name")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name, category string
			var price float64
			if rows.Scan(&id, &name, &category, &price) == nil {
				extraItems = append(extraItems, map[string]interface{}{"id": id, "name": name, "category": category, "price": price})
			}
		}
	}

	html := `<html><head><title>Admin Panel</title></head><body><center>
<h1>Admin Panel</h1>
<hr>
<button onclick="showTab('users-tab')">Users</button>
<button onclick="showTab('orders-tab')">Orders</button>
<button onclick="showTab('delivery-tab')">Delivery</button>
<button onclick="showTab('pizzas-tab')">Pizzas</button>
<button onclick="showTab('ingredients-tab')">Ingredients</button>
<button onclick="showTab('extras-tab')">Desserts & Drinks</button>
<button onclick="showTab('discounts-tab')">Discount Codes</button>
<button onclick="showTab('reports-tab')">Reports</button>
<hr>

<div id="users-tab" style="display:block;">
<h2>Users</h2>
<h3>Create User</h3>
<form method="POST" action="/admin/users/create">
<table><tr><td><b>Username:</b></td><td><input type="text" name="username" required></td></tr>
<tr><td><b>Password:</b></td><td><input type="password" name="password" required></td></tr>
<tr><td><b>Role:</b></td><td>
<select name="role"><option value="customer">Customer</option><option value="admin">Admin</option><option value="delivery_person">Delivery Person</option></select>
</td></tr>
<tr><td colspan="2"><input type="submit" value="Create User"></td></tr></table>
</form>
<hr>
<h3>All Users</h3>
<table border="1"><tr><th>ID</th><th>Username</th><th>Role</th><th>Actions</th></tr>`

	for _, u := range users {
		html += fmt.Sprintf(`<tr><td>%v</td><td>%v</td><td>%v</td><td>
<form method="POST" action="/admin/users/delete" style="display:inline;">
<input type="hidden" name="id" value="%v">
<input type="submit" value="Delete"></form></td></tr>`, u["id"], u["username"], u["role"], u["id"])
	}

	html += `</table></div>

<div id="orders-tab" style="display:none;">
<h2>Orders</h2>
<table border="1"><tr><th>ID</th><th>Username</th><th>Status</th><th>Delivery Address</th><th>Postal Code</th><th>Driver</th><th>Actions</th></tr>`

	for _, o := range orders {
		// Get order details (pizzas and extras)
		orderDetails, _ := database.GetOrderDetails(o.ID)
		var itemsHTML string

		if len(orderDetails.Pizzas) > 0 || len(orderDetails.ExtraItems) > 0 {
			itemsHTML = "<b>Pizzas:</b><br>"
			var pizzaTotal float64
			for _, p := range orderDetails.Pizzas {
				pizzaInfo, _ := database.GetPizzaInformation(p.PizzaName)
				priceFloat, _ := pizzaInfo.Cost.Float64()
				pizzaTotal += priceFloat * float64(p.Quantity)
				itemsHTML += fmt.Sprintf("- %s (x%d) @ $%.2f = $%.2f<br>", p.PizzaName, p.Quantity, priceFloat, priceFloat*float64(p.Quantity))
			}

			var extrasTotal float64
			if len(orderDetails.ExtraItems) > 0 {
				itemsHTML += "<br><b>Extras:</b><br>"
				for _, e := range orderDetails.ExtraItems {
					extrasTotal += e.Price * float64(e.Quantity)
					itemsHTML += fmt.Sprintf("- %s (x%d) @ $%.2f = $%.2f<br>", e.ExtraItemName, e.Quantity, e.Price, e.Price*float64(e.Quantity))
				}
			}

			itemsHTML += fmt.Sprintf("<br><b>Total: $%.2f</b>", pizzaTotal+extrasTotal)
		}

		// Prepare driver dropdown
		driverDropdown := `<form method="POST" action="/admin/orders/assign-delivery" style="display:inline;">
<input type="hidden" name="order_id" value="` + fmt.Sprintf("%d", o.ID) + `">
<select name="delivery_person_id">
<option value="">No Driver</option>`

		for _, dp := range deliveryPersons {
			selected := ""
			dpID, ok := dp["id"].(int)
			if !ok {
				// Try int64
				dpID64, ok := dp["id"].(int64)
				if ok {
					dpID = int(dpID64)
				}
			}
			if o.DeliveryPersonID != nil && *o.DeliveryPersonID == dpID {
				selected = "selected"
			}
			driverDropdown += fmt.Sprintf(`<option value="%v" %s>%v</option>`, dp["id"], selected, dp["username"])
		}
		driverDropdown += `</select><input type="submit" value="Assign"></form>`

		driverDisplay := "None"
		if o.DeliveryPersonName != nil {
			driverDisplay = *o.DeliveryPersonName
		}

		html += fmt.Sprintf(`<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s<br>%s</td><td>
<button type="button" onclick="document.getElementById('order-details-%d').style.display = document.getElementById('order-details-%d').style.display === 'none' ? 'table-row' : 'none'">View Details</button>
<form method="POST" action="/admin/orders/update-status" style="display:inline;">
<input type="hidden" name="id" value="%d">
<select name="status"><option value="IN_PROGRESS" %s>IN_PROGRESS</option><option value="DELIVERED" %s>DELIVERED</option><option value="FAILED" %s>FAILED</option></select>
<input type="submit" value="Update">
</form>
<form method="POST" action="/admin/orders/delete" style="display:inline;">
<input type="hidden" name="id" value="%d">
<input type="submit" value="Delete" onclick="return confirm('Delete this order?')"></form></td></tr>
<tr id="order-details-%d" style="display:none;"><td colspan="7" style="background:#f0f0f0;padding:10px;">%s</td></tr>`,
			o.ID, o.CustomerName, o.Status, o.DeliveryAddress, o.PostalCode, driverDisplay, driverDropdown,
			o.ID, o.ID,
			o.ID,
			func() string {
				if o.Status == "IN_PROGRESS" {
					return "selected"
				}
				return ""
			}(),
			func() string {
				if o.Status == "DELIVERED" {
					return "selected"
				}
				return ""
			}(),
			func() string {
				if o.Status == "FAILED" {
					return "selected"
				}
				return ""
			}(),
			o.ID, o.ID, itemsHTML)
	}

	html += `</table></div>

<div id="delivery-tab" style="display:none;">
<h2>Delivery Persons</h2>
<table border="1"><tr><th>ID</th><th>Username</th><th>Vehicle Type</th><th>Actions</th></tr>`

	for _, d := range deliveryPersons {
		html += fmt.Sprintf(`<tr><td>%v</td><td>%v</td><td>%v</td><td>
<form method="POST" action="/admin/delivery/delete" style="display:inline;">
<input type="hidden" name="id" value="%v">
<input type="submit" value="Delete"></form></td></tr>`, d["id"], d["username"], d["vehicle_type"], d["id"])
	}

	html += `</table></div>

<div id="pizzas-tab" style="display:none;">
<h2>Pizzas</h2>
<h3>Create Pizza</h3>
<form method="POST" action="/admin/pizza/create">
<table><tr><td><b>Name:</b></td><td><input type="text" name="name" required></td></tr>
<tr><td><b>Ingredients (comma-separated):</b></td><td><input type="text" name="ingredients" placeholder="cheese, tomato, basil" required></td></tr>
<tr><td colspan="2"><input type="submit" value="Create Pizza"></td></tr></table>
</form>
<hr>
<h3>All Pizzas</h3>
<table border="1"><tr><th>ID</th><th>Name</th><th>Ingredients</th><th>Price</th><th>Actions</th></tr>`

	for _, p := range pizzas {
		var ingredientNames []string
		for _, ingr := range p.Ingredients {
			ingredientNames = append(ingredientNames, ingr.Ingr.Name)
		}
		html += fmt.Sprintf(`<tr><td>%d</td><td>%s</td><td>%s</td><td>%.2f</td><td>
<form method="POST" action="/admin/pizza/delete" style="display:inline;">
<input type="hidden" name="id" value="%d">
<input type="submit" value="Delete"></form></td></tr>`,
			p.ID, p.Name, strings.Join(ingredientNames, ", "), p.Price, p.ID)
	}

	html += `</table></div>

<div id="ingredients-tab" style="display:none;">
<h2>Ingredients</h2>
<h3>Create Ingredient</h3>
<form method="POST" action="/admin/ingredient/create">
<table><tr><td><b>Name:</b></td><td><input type="text" name="name" required></td></tr>
<tr><td><b>Cost (cents):</b></td><td><input type="number" name="cost" required></td></tr>
<tr><td><b>Has Meat:</b></td><td><input type="checkbox" name="has_meat"></td></tr>
<tr><td><b>Has Animal Products:</b></td><td><input type="checkbox" name="has_animal"></td></tr>
<tr><td colspan="2"><input type="submit" value="Create Ingredient"></td></tr></table>
</form>
<hr>
<h3>All Ingredients</h3>
<table border="1"><tr><th>ID</th><th>Name</th><th>Cost (cents)</th><th>Has Meat</th><th>Has Animal Products</th><th>Actions</th></tr>`

	for _, i := range ingredients {
		costCents, _ := i.Ingr.Cost.Shift(2).Float64()
		meatChecked := ""
		animalChecked := ""
		if i.Ingr.HasMeat {
			meatChecked = "checked"
		}
		if i.Ingr.HasAnimalProducts {
			animalChecked = "checked"
		}
		html += fmt.Sprintf(`<tr>
<form method="POST" action="/admin/ingredient/update" style="display:inline;">
<td>%d<input type="hidden" name="id" value="%d"></td>
<td><input type="text" name="name" value="%s" required></td>
<td><input type="number" name="cost" value="%.0f" required style="width:80px;"></td>
<td><input type="checkbox" name="has_meat" %s></td>
<td><input type="checkbox" name="has_animal" %s></td>
<td><input type="submit" value="Update">
<button type="button" onclick="this.closest('tr').querySelector('form').reset()">Cancel</button>
</form>
<form method="POST" action="/admin/ingredient/delete" style="display:inline;">
<input type="hidden" name="id" value="%d">
<input type="submit" value="Delete" onclick="return confirm('Delete this ingredient?')"></form></td></tr>`,
			i.ID, i.ID, i.Ingr.Name, costCents, meatChecked, animalChecked, i.ID)
	}

	html += `</table></div>

<div id="extras-tab" style="display:none;">
<h2>Desserts & Drinks</h2>
<h3>Create Item</h3>
<form method="POST" action="/admin/extra-items/create">
<table><tr><td><b>Name:</b></td><td><input type="text" name="name" required></td></tr>
<tr><td><b>Category:</b></td><td><select name="category"><option value="dessert">Dessert</option><option value="drink">Drink</option></select></td></tr>
<tr><td><b>Price:</b></td><td><input type="number" name="price" step="0.01" required></td></tr>
<tr><td colspan="2"><input type="submit" value="Create Item"></td></tr></table>
</form>
<hr>
<h3>All Items</h3>
<table border="1"><tr><th>ID</th><th>Name</th><th>Category</th><th>Price</th><th>Actions</th></tr>`

	for _, e := range extraItems {
		html += fmt.Sprintf(`<tr>
<form method="POST" action="/admin/extra-items/update" style="display:inline;">
<td>%v<input type="hidden" name="id" value="%v"></td>
<td><input type="text" name="name" value="%v" required></td>
<td><select name="category"><option value="dessert" %s>Dessert</option><option value="drink" %s>Drink</option></select></td>
<td><input type="number" name="price" value="%.2f" step="0.01" required style="width:80px;"></td>
<td><input type="submit" value="Update">
<button type="button" onclick="this.closest('tr').querySelector('form').reset()">Cancel</button>
</form>
<form method="POST" action="/admin/extra-items/delete" style="display:inline;">
<input type="hidden" name="id" value="%v">
<input type="submit" value="Delete" onclick="return confirm('Delete this item?')"></form></td></tr>`,
			e["id"], e["id"], e["name"],
			func() string {
				if e["category"] == "dessert" {
					return "selected"
				}
				return ""
			}(),
			func() string {
				if e["category"] == "drink" {
					return "selected"
				}
				return ""
			}(),
			e["price"], e["id"])
	}

	html += `</table></div>

<div id="discounts-tab" style="display:none;">
<h2>Discount Codes</h2>
<h3>Create Discount Code</h3>
<form method="POST" action="/admin/discount/create">
<table><tr><td><b>Code:</b></td><td><input type="text" name="code" required></td></tr>
<tr><td><b>Discount %:</b></td><td><input type="number" name="percentage" min="1" max="100" required></td></tr>
<tr><td colspan="2"><input type="submit" value="Create Code"></td></tr></table>
</form>
<hr>
<h3>All Discount Codes</h3>
<table border="1"><tr><th>ID</th><th>Code</th><th>Discount %</th><th>Active</th><th>Actions</th></tr>`

	// Fetch discount codes
	discountRows, err := database.DATABASE.Query("SELECT id, code, discount_percentage, is_active FROM discount_code ORDER BY code")
	if err == nil {
		defer discountRows.Close()
		for discountRows.Next() {
			var id, percentage int
			var code string
			var isActive bool
			if discountRows.Scan(&id, &code, &percentage, &isActive) == nil {
				activeChecked := ""
				if isActive {
					activeChecked = "checked"
				}
				html += fmt.Sprintf(`<tr>
<form method="POST" action="/admin/discount/update" style="display:inline;">
<td>%d<input type="hidden" name="id" value="%d"></td>
<td><input type="text" name="code" value="%s" required></td>
<td><input type="number" name="percentage" value="%d" min="1" max="100" required style="width:60px;"></td>
<td><input type="checkbox" name="is_active" %s></td>
<td><input type="submit" value="Update">
<button type="button" onclick="this.closest('tr').querySelector('form').reset()">Cancel</button>
</form>
<form method="POST" action="/admin/discount/delete" style="display:inline;">
<input type="hidden" name="id" value="%d">
<input type="submit" value="Delete" onclick="return confirm('Delete this discount code?')"></form></td></tr>`,
					id, id, code, percentage, activeChecked, id)
			}
		}
	}

	html += `</table></div>

<div id="reports-tab" style="display:none;">
<h2>📊 Staff Reports</h2>

<h3>📦 Undelivered Orders</h3>
<table border="1"><tr><th>Order ID</th><th>Customer</th><th>Address</th><th>Status</th><th>Timestamp</th><th>Items</th></tr>`

	// Report 1: Undelivered Orders (IN_PROGRESS or OUT_FOR_DELIVERY)
	undeliveredQuery := `
		SELECT o.id, c.name, o.delivery_address, o.status, o.timestamp
		FROM orders o
		JOIN customer c ON o.customer_id = c.id
		WHERE o.status IN ('IN_PROGRESS', 'OUT_FOR_DELIVERY')
		ORDER BY o.timestamp DESC
	`
	undeliveredRows, err := database.DATABASE.Query(undeliveredQuery)
	if err == nil {
		defer undeliveredRows.Close()
		for undeliveredRows.Next() {
			var orderID int
			var customerName, address, status string
			var timestamp time.Time
			if undeliveredRows.Scan(&orderID, &customerName, &address, &status, &timestamp) == nil {
				// Get order items
				itemsQuery := `
					SELECT p.name, op.quantity 
					FROM order_pizza op 
					JOIN pizza p ON op.pizza_id = p.id 
					WHERE op.order_id = ?
				`
				itemRows, _ := database.DATABASE.Query(itemsQuery, orderID)
				var items []string
				if itemRows != nil {
					for itemRows.Next() {
						var pizzaName string
						var qty int
						if itemRows.Scan(&pizzaName, &qty) == nil {
							items = append(items, fmt.Sprintf("%s x%d", pizzaName, qty))
						}
					}
					itemRows.Close()
				}
				itemsList := strings.Join(items, ", ")
				if itemsList == "" {
					itemsList = "No items"
				}

				html += fmt.Sprintf(`<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					orderID, customerName, address, status, timestamp.Format("2006-01-02 15:04"), itemsList)
			}
		}
	}

	html += `</table><br><hr width="70%"><br>

<h3>🏆 Top 3 Pizzas (Last 30 Days)</h3>
<table border="1"><tr><th>Rank</th><th>Pizza</th><th>Total Sold</th></tr>`

	// Report 2: Top 3 Pizzas
	topPizzasQuery := `
		SELECT p.name, SUM(op.quantity) as total_sold
		FROM order_pizza op
		JOIN pizza p ON op.pizza_id = p.id
		JOIN orders o ON op.order_id = o.id
		WHERE o.timestamp >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY p.id, p.name
		ORDER BY total_sold DESC
		LIMIT 3
	`
	topPizzasRows, err := database.DATABASE.Query(topPizzasQuery)
	if err == nil {
		defer topPizzasRows.Close()
		rank := 1
		for topPizzasRows.Next() {
			var pizzaName string
			var totalSold int
			if topPizzasRows.Scan(&pizzaName, &totalSold) == nil {
				medal := "🥇"
				if rank == 2 {
					medal = "🥈"
				} else if rank == 3 {
					medal = "🥉"
				}
				html += fmt.Sprintf(`<tr><td>%s #%d</td><td><b>%s</b></td><td>%d</td></tr>`,
					medal, rank, pizzaName, totalSold)
				rank++
			}
		}
	}

	html += `</table><br><hr width="70%"><br>

<h3>💰 Revenue by Customer Gender</h3>
<table border="1"><tr><th>Gender</th><th>Total Revenue</th><th>Orders</th><th>Avg Order Value</th></tr>`

	// Report 3: Revenue by Gender
	genderRevenueQuery := `
		SELECT c.gender, COUNT(DISTINCT o.id) as order_count,
		       SUM(
		           (SELECT SUM(
		               (SELECT ROUND(SUM(i.cost) * 1.4 * 1.09, 2)
		                FROM pizza_ingredient pi
		                JOIN ingredient i ON pi.ingredient_id = i.id
		                WHERE pi.pizza_id = op.pizza_id) * op.quantity
		           ) FROM order_pizza op WHERE op.order_id = o.id)
		       ) as total_revenue
		FROM orders o
		JOIN customer c ON o.customer_id = c.id
		GROUP BY c.gender
		ORDER BY total_revenue DESC
	`
	genderRevenueRows, err := database.DATABASE.Query(genderRevenueQuery)
	if err == nil {
		defer genderRevenueRows.Close()
		for genderRevenueRows.Next() {
			var gender string
			var orderCount int
			var totalRevenue sql.NullFloat64
			if genderRevenueRows.Scan(&gender, &orderCount, &totalRevenue) == nil {
				revenue := 0.0
				if totalRevenue.Valid {
					revenue = totalRevenue.Float64
				}
				avgOrder := 0.0
				if orderCount > 0 {
					avgOrder = revenue / float64(orderCount)
				}
				html += fmt.Sprintf(`<tr><td>%s</td><td>$%.2f</td><td>%d</td><td>$%.2f</td></tr>`,
					gender, revenue, orderCount, avgOrder)
			}
		}
	}

	html += `</table><br><hr width="70%"><br>

<h3>👥 Revenue by Age Group</h3>
<table border="1"><tr><th>Age Group</th><th>Total Revenue</th><th>Orders</th><th>Avg Order Value</th></tr>`

	// Report 4: Revenue by Age Group
	ageGroupRevenueQuery := `
		SELECT 
		    CASE 
		        WHEN TIMESTAMPDIFF(YEAR, c.birth_date, CURDATE()) < 25 THEN 'Under 25'
		        WHEN TIMESTAMPDIFF(YEAR, c.birth_date, CURDATE()) BETWEEN 25 AND 34 THEN '25-34'
		        WHEN TIMESTAMPDIFF(YEAR, c.birth_date, CURDATE()) BETWEEN 35 AND 44 THEN '35-44'
		        WHEN TIMESTAMPDIFF(YEAR, c.birth_date, CURDATE()) BETWEEN 45 AND 54 THEN '45-54'
		        WHEN TIMESTAMPDIFF(YEAR, c.birth_date, CURDATE()) >= 55 THEN '55+'
		        ELSE 'Unknown'
		    END as age_group,
		    COUNT(DISTINCT o.id) as order_count,
		    SUM(
		        (SELECT SUM(
		            (SELECT ROUND(SUM(i.cost) * 1.4 * 1.09, 2)
		             FROM pizza_ingredient pi
		             JOIN ingredient i ON pi.ingredient_id = i.id
		             WHERE pi.pizza_id = op.pizza_id) * op.quantity
		        ) FROM order_pizza op WHERE op.order_id = o.id)
		    ) as total_revenue
		FROM orders o
		JOIN customer c ON o.customer_id = c.id
		WHERE c.birth_date IS NOT NULL
		GROUP BY age_group
		ORDER BY 
		    CASE age_group
		        WHEN 'Under 25' THEN 1
		        WHEN '25-34' THEN 2
		        WHEN '35-44' THEN 3
		        WHEN '45-54' THEN 4
		        WHEN '55+' THEN 5
		        ELSE 6
		    END
	`
	ageGroupRevenueRows, err := database.DATABASE.Query(ageGroupRevenueQuery)
	if err == nil {
		defer ageGroupRevenueRows.Close()
		for ageGroupRevenueRows.Next() {
			var ageGroup string
			var orderCount int
			var totalRevenue sql.NullFloat64
			if ageGroupRevenueRows.Scan(&ageGroup, &orderCount, &totalRevenue) == nil {
				revenue := 0.0
				if totalRevenue.Valid {
					revenue = totalRevenue.Float64
				}
				avgOrder := 0.0
				if orderCount > 0 {
					avgOrder = revenue / float64(orderCount)
				}
				html += fmt.Sprintf(`<tr><td>%s</td><td>$%.2f</td><td>%d</td><td>$%.2f</td></tr>`,
					ageGroup, revenue, orderCount, avgOrder)
			}
		}
	}

	html += `</table><br><hr width="70%"><br>

<h3>📍 Revenue by Postal Code (Top 10)</h3>
<table border="1"><tr><th>Postal Code</th><th>Total Revenue</th><th>Orders</th><th>Avg Order Value</th></tr>`

	// Report 5: Revenue by Postal Code
	postalCodeRevenueQuery := `
		SELECT c.postal_code, COUNT(DISTINCT o.id) as order_count,
		       SUM(
		           (SELECT SUM(
		               (SELECT ROUND(SUM(i.cost) * 1.4 * 1.09, 2)
		                FROM pizza_ingredient pi
		                JOIN ingredient i ON pi.ingredient_id = i.id
		                WHERE pi.pizza_id = op.pizza_id) * op.quantity
		           ) FROM order_pizza op WHERE op.order_id = o.id)
		       ) as total_revenue
		FROM orders o
		JOIN customer c ON o.customer_id = c.id
		GROUP BY c.postal_code
		ORDER BY total_revenue DESC
		LIMIT 10
	`
	postalCodeRevenueRows, err := database.DATABASE.Query(postalCodeRevenueQuery)
	if err == nil {
		defer postalCodeRevenueRows.Close()
		for postalCodeRevenueRows.Next() {
			var postalCode string
			var orderCount int
			var totalRevenue sql.NullFloat64
			if postalCodeRevenueRows.Scan(&postalCode, &orderCount, &totalRevenue) == nil {
				revenue := 0.0
				if totalRevenue.Valid {
					revenue = totalRevenue.Float64
				}
				avgOrder := 0.0
				if orderCount > 0 {
					avgOrder = revenue / float64(orderCount)
				}
				html += fmt.Sprintf(`<tr><td>%s</td><td>$%.2f</td><td>%d</td><td>$%.2f</td></tr>`,
					postalCode, revenue, orderCount, avgOrder)
			}
		}
	}

	html += `</table>
</div>

<script>
function showTab(tabId) {
  const tabs = ['users-tab', 'orders-tab', 'delivery-tab', 'pizzas-tab', 'ingredients-tab', 'extras-tab', 'discounts-tab', 'reports-tab'];
  tabs.forEach(id => document.getElementById(id).style.display = (id === tabId) ? 'block' : 'none');
}
</script>
</center></body></html>`

	fmt.Fprint(w, html)
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	success, username, _, err := isLoginOK(r)

	msg := ""
	if err != nil {
		msg = err.Error()
		success = false
	}

	role, err := database.GetUserRole(username)
	if err != nil {
		msg = err.Error()
		success = false
	}

	type Msg struct {
		Ok   bool   `json:"ok"`
		Msg  string `json:"msg"`
		Role string `json:"role"`
	}
	sendMsg := Msg{success, msg, role}
	jsonMsg, err := json.Marshal(sendMsg)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(jsonMsg))
}

func isLoginOK(r *http.Request) (bool, string, string, error) {
	type User struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return false, "", "", errors.New("invalid json")
	}
	success, msg := database.TryLogin(user.Username, user.Password)

	if !success {
		return false, "", "", errors.New(msg)
	} else {
		return true, user.Username, user.Password, nil
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		RegisterPostHandler(w, r)
	case http.MethodGet:
		RegisterGetHandler(w, r)
	}
}

func RegisterGetHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/register.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}
func RegisterPostHandler(w http.ResponseWriter, r *http.Request) {

	var customer database.Customer
	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	success, msg := database.TryAddCustomer(customer)

	type Msg struct {
		Ok  bool   `json:"ok"`
		Msg string `json:"msg"`
	}

	sendMsg := Msg{success, msg}
	jsonMsg, err := json.Marshal(sendMsg)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(jsonMsg))
}

func PizzaHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/pizza.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}

func MenuHandler(w http.ResponseWriter, r *http.Request) {
	pizzas, err := database.GetAllPizzas()
	if err != nil {
		http.Error(w, "Failed to load pizzas", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "<h1>Pizza Menu</h1>")
	fmt.Fprintln(w, `<table border="1" cellpadding="5" cellspacing="0">`)
	fmt.Fprintln(w, "<tr><th>Pizza Name</th><th>Cost</th><th>Ingredients</th><th>Diet Info</th></tr>")

	// sort.Slice(pizzas, func(i, j int) bool {
	// 	// Sort ascending by price
	// 	return pizzas[i].Price.LessThan(pizzas[j].Price)
	// })

	pizzaInfos := []database.PizzaInformation{}

	for _, pizza := range pizzas {
		info, err := database.GetPizzaInformation(pizza.Name)
		if err != nil {
			fmt.Fprintf(w, "<tr><td colspan='4'>Failed to load info for %s</td></tr>", pizza.Name)
			continue
		}
		pizzaInfos = append(pizzaInfos, info)
	}

	indexes := make([]int, len(pizzas))
	for i := range indexes {
		indexes[i] = i
	}

	sort.Slice(indexes, func(i, j int) bool {
		return pizzaInfos[indexes[i]].Cost.LessThan(pizzaInfos[indexes[j]].Cost)
	})

	for _, id := range indexes {
		pizza := pizzas[id]
		info := pizzaInfos[id]

		var ingredientNames []string
		for _, ingr := range pizza.Ingredients {
			ingredientNames = append(ingredientNames, ingr.Ingr.Name)
		}

		diet := "Omnivore"
		if info.IsVegan {
			diet = "Vegan"
		} else if info.IsVegetarian {
			diet = "Vegetarian"
		}

		fmt.Fprintf(w,
			"<tr><td>%s</td><td>%s €</td><td>%s</td><td>%s</td></tr>",
			pizza.Name,
			info.Cost.StringFixed(2),
			strings.Join(ingredientNames, ", "),
			diet,
		)
	}

	fmt.Fprintln(w, "</table></body></html>")
}

func AccountHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/account.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}

func GetAccountDetailsHandler(w http.ResponseWriter, r *http.Request) {

	type CustomerResult struct {
		Ok       bool              `json:"ok"`
		Customer database.Customer `json:"customer"`
	}
	customerResult := CustomerResult{Ok: false}

	success, username, password, _ := isLoginOK(r)

	if !success {
		jsonMsg, _ := json.Marshal(customerResult)
		fmt.Fprint(w, string(jsonMsg))
		return
	}

	customer, err := database.GetCustomerDetails(username, password)
	if err != nil {
		jsonMsg, _ := json.Marshal(customerResult)
		fmt.Fprint(w, string(jsonMsg))
		return
	}
	jsonMsg, _ := json.Marshal(CustomerResult{Ok: true, Customer: customer})
	fmt.Fprint(w, string(jsonMsg))

}

// --- Admin APIs ---
// Simple auth: expects JSON body with {username,password} that must be an ADMIN user
type adminAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func isAdminFromHeaders(r *http.Request) bool {
	user := r.Header.Get("X-Username")
	pass := r.Header.Get("X-Password")

	if user == "" || pass == "" {
		userCookie, err1 := r.Cookie("X-Username")
		passCookie, err2 := r.Cookie("X-Password")
		if err1 == nil && err2 == nil {
			user = userCookie.Value
			pass = passCookie.Value
		}
	}

	if user == "" || pass == "" {
		return false
	}
	ok, _ := database.TryLogin(user, pass)
	if !ok {
		return false
	}
	role, err := database.GetUserRole(user)
	if err != nil || role != database.AdminRole.String() {
		return false
	}
	return true
}

func AdminCreateIngredientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var name string
	var costCents int64
	var hasMeat, hasAnimal bool

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		a := adminAuth{Username: payload["username"].(string), Password: payload["password"].(string)}
		ok, msg := func() (bool, string) {
			ok, _ := database.TryLogin(a.Username, a.Password)
			if !ok {
				return false, "invalid credentials"
			}
			role, err := database.GetUserRole(a.Username)
			if err != nil || role != database.AdminRole.String() {
				return false, "not admin"
			}
			return true, ""
		}()
		if !ok {
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}
		name, _ = payload["name"].(string)
		switch v := payload["costCents"].(type) {
		case float64:
			costCents = int64(v)
		case int64:
			costCents = v
		default:
			http.Error(w, "invalid costCents", http.StatusBadRequest)
			return
		}
		hasMeat, _ = payload["hasMeat"].(bool)
		hasAnimal, _ = payload["hasAnimalProducts"].(bool)
	} else {
		if !isAdminFromHeaders(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.ParseForm()
		name = r.FormValue("name")
		fmt.Sscanf(r.FormValue("cost"), "%d", &costCents)
		hasMeat = r.FormValue("has_meat") == "on"
		hasAnimal = r.FormValue("has_animal") == "on"
	}

	ingr := database.NewIngredient(name, costCents, hasMeat, hasAnimal)
	_, err := database.CreateIngredient(ingr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func AdminListIngredientsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ingredients, err := database.GetAllIngredients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ingredients)
}

func AdminDeleteIngredientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isAdminFromHeaders(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var id int
	if r.Method == http.MethodPost {
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
	} else {
		ids := r.URL.Query().Get("id")
		if ids == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		fmt.Sscanf(ids, "%d", &id)
	}

	if err := database.DeleteIngredient(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminUpdateIngredientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isAdminFromHeaders(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var id int
	var name string
	var costCents int64
	var hasMeat, hasAnimal bool

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
		name = r.FormValue("name")
		fmt.Sscanf(r.FormValue("cost"), "%d", &costCents)
		hasMeat = r.FormValue("has_meat") == "on"
		hasAnimal = r.FormValue("has_animal") == "on"
	} else {
		var req struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			CostCents int64  `json:"costCents"`
			HasMeat   bool   `json:"hasMeat"`
			HasAnimal bool   `json:"hasAnimalProducts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		id = req.ID
		name = req.Name
		costCents = req.CostCents
		hasMeat = req.HasMeat
		hasAnimal = req.HasAnimal
	}

	ingr := database.NewIngredient(name, costCents, hasMeat, hasAnimal)
	query := "UPDATE ingredient SET name = ?, cost = ?, has_meat = ?, has_animal_products = ? WHERE id = ?"
	_, err := database.DATABASE.Exec(query, ingr.Name, ingr.Cost.String(), ingr.HasMeat, ingr.HasAnimalProducts, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}
}

func AdminCreatePizzaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var name string
	var ingredients []string

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// JSON API
		var payload struct {
			Username    string   `json:"username"`
			Password    string   `json:"password"`
			Name        string   `json:"name"`
			Ingredients []string `json:"ingredients"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		ok, _ := database.TryLogin(payload.Username, payload.Password)
		if !ok {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		role, err := database.GetUserRole(payload.Username)
		if err != nil || role != database.AdminRole.String() {
			http.Error(w, "not admin", http.StatusUnauthorized)
			return
		}
		name = payload.Name
		ingredients = payload.Ingredients
	} else {
		// Form data
		if !isAdminFromHeaders(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.ParseForm()
		name = r.FormValue("name")
		ingrStr := r.FormValue("ingredients")
		for _, s := range strings.Split(ingrStr, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				ingredients = append(ingredients, s)
			}
		}
	}

	_, err := database.CreatePizza(name, ingredients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.Contains(contentType, "application/json") {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"status":"ok"}`)
	} else {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func AdminListPizzasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pizzas, err := database.GetAllPizzasWithPrice()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(pizzas)
}

func AdminDeletePizzaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isAdminFromHeaders(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var id int
	if r.Method == http.MethodPost {
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
	} else {
		ids := r.URL.Query().Get("id")
		if ids == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		fmt.Sscanf(ids, "%d", &id)
	}

	if err := database.DeletePizza(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		DeliveryAddress string `json:"delivery_address"`
		PostalCode      string `json:"postal_code"`
		DiscountCode    string `json:"discount_code"`
		CartItems       []struct {
			ID       int    `json:"id"`
			Quantity int    `json:"quantity"`
			Type     string `json:"type"` // "pizza" or "extra"
		} `json:"cart_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	success, _ := database.TryLogin(req.Username, req.Password)
	if !success {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Invalid credentials"})
		return
	}

	userID, err := database.GetUserIDFromUsername(req.Username)
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "User not found"})
		return
	}

	customerID, err := database.GetCustomerIDFromUserID(userID)
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Customer not found"})
		return
	}

	if len(req.CartItems) == 0 {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Cart is empty"})
		return
	}

	// Separate pizzas and extra items
	var pizzaItems []struct {
		PizzaID  int
		Quantity int
	}
	var extraItems []struct {
		ExtraItemID int
		Quantity    int
	}

	for _, item := range req.CartItems {
		if item.Type == "extra" {
			extraItems = append(extraItems, struct {
				ExtraItemID int
				Quantity    int
			}{
				ExtraItemID: item.ID,
				Quantity:    item.Quantity,
			})
		} else {
			// Default to pizza if type is missing or "pizza"
			pizzaItems = append(pizzaItems, struct {
				PizzaID  int
				Quantity int
			}{
				PizzaID:  item.ID,
				Quantity: item.Quantity,
			})
		}
	}

	orderID, err := database.CreateOrderWithTransaction(
		customerID,
		userID,
		req.DeliveryAddress,
		req.PostalCode,
		pizzaItems,
		extraItems,
		&req.DiscountCode,
	)
	if err != nil {
		fmt.Println(err)
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}

		// Check for specific errors
		errorMsg := "Failed to create order"
		if err.Error() == "discount code already used" {
			errorMsg = "You have already used this discount code"
		}

		json.NewEncoder(w).Encode(Msg{Ok: false, Error: errorMsg})
		return
	}

	type Msg struct {
		Ok      bool `json:"ok"`
		OrderID int  `json:"order_id"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true, OrderID: orderID})
}

func GetAvailableDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get credentials from cookies
	var username, password string
	userCookie, err := r.Cookie("user")
	passCookie, err2 := r.Cookie("pass")

	if err == nil && err2 == nil {
		username = userCookie.Value
		password = passCookie.Value
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Not authenticated",
		})
		return
	}

	success, _ := database.TryLogin(username, password)
	if !success {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Invalid credentials",
		})
		return
	}

	// Verify user is a delivery person
	role, err := database.GetUserRole(username)
	if err != nil || role != database.DeliveryRole.String() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Not authorized as delivery person",
		})
		return
	}

	// Get available deliveries
	deliveries, err := database.GetAvailableDeliveries()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Failed to get available deliveries",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":     true,
		"orders": deliveries,
	})
}

func GetAssignedDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get credentials from cookies
	var username, password string
	userCookie, err := r.Cookie("user")
	passCookie, err2 := r.Cookie("pass")

	if err == nil && err2 == nil {
		username = userCookie.Value
		password = passCookie.Value
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Not authenticated",
		})
		return
	}

	success, _ := database.TryLogin(username, password)
	if !success {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Invalid credentials",
		})
		return
	}

	// Verify user is a delivery person
	role, err := database.GetUserRole(username)
	if err != nil || role != database.DeliveryRole.String() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Not authorized as delivery person",
		})
		return
	}

	// Get user ID and delivery person ID
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "User not found",
		})
		return
	}

	deliveryPersonID, err := database.GetDeliveryPersonIDFromUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Delivery person not found",
		})
		return
	}

	// Get assigned deliveries
	deliveries, err := database.GetAssignedDeliveries(deliveryPersonID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "Failed to get assigned deliveries",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":     true,
		"orders": deliveries,
	})
}

func AssignDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify delivery person authentication - check cookies
	var username, password string
	userCookie, err := r.Cookie("user")
	passCookie, err2 := r.Cookie("pass")

	if err == nil && err2 == nil {
		username = userCookie.Value
		password = passCookie.Value
	} else {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	success, _ := database.TryLogin(username, password)
	if !success {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify user is a delivery person
	role, err := database.GetUserRole(username)
	if err != nil || role != database.DeliveryRole.String() {
		http.Error(w, "Not authorized as delivery person", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		OrderID int `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID and delivery person ID
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	deliveryPersonID, err := database.GetDeliveryPersonIDFromUserID(userID)
	if err != nil {
		http.Error(w, "Delivery person not found", http.StatusInternalServerError)
		return
	}

	// Assign delivery
	err = database.AssignDelivery(req.OrderID, deliveryPersonID)
	if err == database.ErrOrderNotAvailable {
		http.Error(w, "Order is not available", http.StatusBadRequest)
		return
	}
	if err == database.ErrOrderAlreadyAssigned {
		http.Error(w, "Order is already assigned", http.StatusConflict)
		return
	}
	if err == database.ErrDeliveryPersonUnavailable {
		http.Error(w, "You are currently unavailable (cooldown)", http.StatusConflict)
		return
	}
	if err == database.ErrDeliveryPersonBusy {
		http.Error(w, "You already have an active delivery", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "Failed to assign delivery", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func UpdateDeliveryStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify delivery person authentication - check cookies
	var username, password string
	userCookie, err := r.Cookie("user")
	passCookie, err2 := r.Cookie("pass")

	if err == nil && err2 == nil {
		username = userCookie.Value
		password = passCookie.Value
	} else {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	success, _ := database.TryLogin(username, password)
	if !success {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify user is a delivery person
	role, err := database.GetUserRole(username)
	if err != nil || role != database.DeliveryRole.String() {
		http.Error(w, "Not authorized as delivery person", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		OrderID int    `json:"order_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update delivery status
	err = database.UpdateDeliveryStatus(req.OrderID, req.Status)
	if err == database.ErrInvalidStatus {
		http.Error(w, "Invalid status. Must be 'DELIVERED' or 'FAILED'", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Failed to update delivery status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	success, _ := database.TryLogin(req.Username, req.Password)
	if !success {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Invalid credentials"})
		return
	}

	userID, err := database.GetUserIDFromUsername(req.Username)
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "User not found"})
		return
	}

	customerID, err := database.GetCustomerIDFromUserID(userID)
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Customer not found"})
		return
	}

	orders, err := database.GetOrdersByCustomer(customerID)
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Failed to get orders"})
		return
	}

	type Msg struct {
		Ok     bool             `json:"ok"`
		Orders []database.Order `json:"orders"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true, Orders: orders})
}

func GetOrderDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		OrderID  int    `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	success, role := database.TryLogin(req.Username, req.Password)
	if !success {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Invalid credentials"})
		return
	}

	details, err := database.GetOrderDetails(req.OrderID)
	if err != nil {
		fmt.Println("GetOrderDetails error:", err)
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Failed to get order details"})
		return
	}

	// Admins can view any order, customers can only view their own
	if role != "ADMIN" {
		userID, err := database.GetUserIDFromUsername(req.Username)
		if err != nil {
			type Msg struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}
			json.NewEncoder(w).Encode(Msg{Ok: false, Error: "User not found"})
			return
		}

		customerID, err := database.GetCustomerIDFromUserID(userID)
		if err != nil {
			type Msg struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}
			json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Customer not found"})
			return
		}

		if details.Order.CustomerID != customerID {
			type Msg struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}
			json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Unauthorized"})
			return
		}
	}

	type Msg struct {
		Ok    bool                   `json:"ok"`
		Order *database.OrderDetails `json:"order"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true, Order: details})
}

func DeliveryPerson(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/delivery_person.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}

func AdminGetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := database.GetAllUsers()
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Failed to get users"})
		return
	}

	type Msg struct {
		Ok    bool                     `json:"ok"`
		Users []map[string]interface{} `json:"users"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true, Users: users})
}

func AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var id int
	if r.Method == http.MethodPost {
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
	} else {
		userID := r.URL.Query().Get("id")
		if userID == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
		fmt.Sscanf(userID, "%d", &id)
	}

	err := database.DeleteUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	type Msg struct {
		Ok bool `json:"ok"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true})
}

func AdminCreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Form submission from server-side rendered page
		if !isAdminFromHeaders(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		role := r.FormValue("role")

		if role == "customer" {
			customer := database.Customer{
				Username: username,
				Password: password,
			}
			database.TryAddCustomer(customer)
		} else if role == "delivery_person" {
			deliveryPerson := database.DeliveryPerson{
				Username: username,
				Password: password,
			}
			database.TryAddDeliveryPerson(deliveryPerson)
		} else if role == "admin" {
			// Create admin user (you may need to add this function)
			customer := database.Customer{
				Username: username,
				Password: password,
			}
			database.TryAddCustomer(customer)
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// JSON API (backward compatibility)
	var req struct {
		AdminUsername string `json:"admin_username"`
		AdminPassword string `json:"admin_password"`
		UserType      string `json:"user_type"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		Name          string `json:"name"`
		Gender        string `json:"gender"`
		BirthDate     string `json:"birth_date"`
		NoBirthDate   bool   `json:"no_birth_date"`
		Address       string `json:"address"`
		PostalCode    string `json:"postal_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	success, role := database.TryLogin(req.AdminUsername, req.AdminPassword)
	if !success {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Invalid admin credentials"})
		return
	}
	if role != "ADMIN" {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "User is not an admin (role: " + role + ")"})
		return
	}

	if req.UserType == "customer" {
		customer := database.Customer{
			Username:    req.Username,
			Password:    req.Password,
			Name:        req.Name,
			Gender:      req.Gender,
			BirthDate:   req.BirthDate,
			NoBirthDate: req.NoBirthDate,
			Address:     req.Address,
			PostCode:    req.PostalCode,
		}
		ok, msg := database.TryAddCustomer(customer)
		type Msg struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message,omitempty"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: ok, Message: msg})
	} else if req.UserType == "delivery" {
		deliveryPerson := database.DeliveryPerson{
			Username: req.Username,
			Password: req.Password,
			Name:     req.Name,
		}
		ok, msg := database.TryAddDeliveryPerson(deliveryPerson)
		type Msg struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message,omitempty"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: ok, Message: msg})
	} else {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Invalid user type"})
	}
}

func AdminGetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orders, err := database.GetAllOrders()
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Failed to get orders"})
		return
	}

	type Msg struct {
		Ok     bool             `json:"ok"`
		Orders []database.Order `json:"orders"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true, Orders: orders})
}

func AdminDeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var id int
	if r.Method == http.MethodPost {
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
	} else {
		orderID := r.URL.Query().Get("id")
		if orderID == "" {
			http.Error(w, "Order ID required", http.StatusBadRequest)
			return
		}
		fmt.Sscanf(orderID, "%d", &id)
	}

	err := database.DeleteOrder(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	type Msg struct {
		Ok bool `json:"ok"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true})
}

func AdminUpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var orderID int
	var status string

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Form submission
		if !isAdminFromHeaders(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &orderID)
		status = r.FormValue("status")

		err := database.UpdateOrderStatus(orderID, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// JSON API
	var req struct {
		AdminUsername string `json:"admin_username"`
		AdminPassword string `json:"admin_password"`
		OrderID       int    `json:"order_id"`
		Status        string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	success, role := database.TryLogin(req.AdminUsername, req.AdminPassword)
	if !success || role != "ADMIN" {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Unauthorized"})
		return
	}

	err := database.UpdateOrderStatus(req.OrderID, req.Status)
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Failed to update order"})
		return
	}

	type Msg struct {
		Ok bool `json:"ok"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true})
}

func AdminGetAllDeliveryPersonsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deliveryPersons, err := database.GetAllDeliveryPersons()
	if err != nil {
		type Msg struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}
		json.NewEncoder(w).Encode(Msg{Ok: false, Error: "Failed to get delivery persons"})
		return
	}

	type Msg struct {
		Ok              bool                     `json:"ok"`
		DeliveryPersons []map[string]interface{} `json:"delivery_persons"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true, DeliveryPersons: deliveryPersons})
}

func AdminDeleteDeliveryPersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var id int
	if r.Method == http.MethodPost {
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
	} else {
		userID := r.URL.Query().Get("id")
		if userID == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
		fmt.Sscanf(userID, "%d", &id)
	}

	err := database.DeleteDeliveryPerson(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	type Msg struct {
		Ok bool `json:"ok"`
	}
	json.NewEncoder(w).Encode(Msg{Ok: true})
}

func ListExtraItemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `SELECT id, name, category, price FROM extra_item ORDER BY category, name`
	rows, err := database.DATABASE.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var name, category string
		var price float64
		err := rows.Scan(&id, &name, &category, &price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		items = append(items, map[string]interface{}{
			"id":       id,
			"name":     name,
			"category": category,
			"price":    price,
		})
	}

	json.NewEncoder(w).Encode(items)
}

func CreateExtraItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var name, category string
	var price float64

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Form submission
		r.ParseForm()
		name = r.FormValue("name")
		category = r.FormValue("category")
		fmt.Sscanf(r.FormValue("price"), "%f", &price)
	} else {
		// JSON API
		var req struct {
			Name     string  `json:"name"`
			Category string  `json:"category"`
			Price    float64 `json:"price"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		name = req.Name
		category = req.Category
		price = req.Price
	}

	if category != "dessert" && category != "drink" {
		http.Error(w, "Category must be 'dessert' or 'drink'", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO extra_item (name, category, price) VALUES (?, ?, ?)`
	result, err := database.DATABASE.Exec(query, name, category, price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	id, _ := result.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":       true,
		"id":       id,
		"name":     name,
		"category": category,
		"price":    price,
	})
}

func UpdateExtraItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var id int
	var name, category string
	var price float64

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Form submission
		r.ParseForm()
		fmt.Sscanf(r.FormValue("id"), "%d", &id)
		name = r.FormValue("name")
		category = r.FormValue("category")
		fmt.Sscanf(r.FormValue("price"), "%f", &price)
	} else {
		// JSON API
		var req struct {
			ID       int     `json:"id"`
			Name     string  `json:"name"`
			Category string  `json:"category"`
			Price    float64 `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		id = req.ID
		name = req.Name
		category = req.Category
		price = req.Price
	}

	if category != "dessert" && category != "drink" {
		http.Error(w, "Category must be 'dessert' or 'drink'", http.StatusBadRequest)
		return
	}

	query := `UPDATE extra_item SET name = ?, category = ?, price = ? WHERE id = ?`
	_, err := database.DATABASE.Exec(query, name, category, price, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

func DeleteExtraItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var id string
	if r.Method == http.MethodPost {
		r.ParseForm()
		id = r.FormValue("id")
	} else {
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM extra_item WHERE id = ?`
	_, err := database.DATABASE.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

// Discount code handlers
func ValidateDiscountCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"valid": false,
		})
		return
	}

	// Get user ID from cookie
	userCookie, err := r.Cookie("user")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"valid":   false,
			"message": "Please login to use discount codes",
		})
		return
	}
	username := userCookie.Value

	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    true,
			"valid": false,
		})
		return
	}

	var discountCodeID int
	var discountPercentage int
	var isActive bool
	err = database.DATABASE.QueryRow(`SELECT id, discount_percentage, is_active FROM discount_code WHERE code = ?`, code).Scan(&discountCodeID, &discountPercentage, &isActive)

	if err != nil || !isActive {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"valid":   false,
			"message": "Invalid or inactive discount code",
		})
		return
	}

	// Check if user has already used this discount code
	var usageCount int
	err = database.DATABASE.QueryRow(`SELECT COUNT(*) FROM discount_usage WHERE user_id = ? AND discount_code_id = ?`, userID, discountCodeID).Scan(&usageCount)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"valid": false,
		})
		return
	}

	if usageCount > 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"valid":   false,
			"message": "You have already used this discount code",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"ok":                  true,
		"valid":               true,
		"discount_percentage": discountPercentage,
		"discount_code_id":    discountCodeID,
	}

	// Special message for birthday discount
	if code == "BIRTHDAY" {
		response["message"] = "🎉 Birthday Special: Free cheapest pizza + free drink! 🎉"
		response["is_birthday"] = true
	}

	json.NewEncoder(w).Encode(response)
}

// CheckBirthdayDiscountHandler checks if it's the user's birthday
func CheckBirthdayDiscountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get username from cookie
	userCookie, err := r.Cookie("user")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          false,
			"is_birthday": false,
		})
		return
	}
	username := userCookie.Value

	// Get user ID from username
	userID, err := database.GetUserIDFromUsername(username)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          false,
			"is_birthday": false,
		})
		return
	}

	// Check if it's their birthday
	isBirthday, err := database.CheckCustomerBirthday(int64(userID))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          false,
			"is_birthday": false,
		})
		return
	}

	if !isBirthday {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"is_birthday": false,
		})
		return
	}

	// Check if they already used the birthday discount
	var discountCodeID int
	err = database.DATABASE.QueryRow(`SELECT id FROM discount_code WHERE code = 'BIRTHDAY'`).Scan(&discountCodeID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          false,
			"is_birthday": false,
		})
		return
	}

	var usageCount int
	err = database.DATABASE.QueryRow(`SELECT COUNT(*) FROM discount_usage WHERE user_id = ? AND discount_code_id = ?`, userID, discountCodeID).Scan(&usageCount)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          false,
			"is_birthday": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if usageCount > 0 {
		// Already used birthday discount
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":           true,
			"is_birthday":  true,
			"already_used": true,
			"message":      "You already used your birthday discount this year!",
		})
	} else {
		// Can use birthday discount
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":           true,
			"is_birthday":  true,
			"already_used": false,
			"code":         "BIRTHDAY",
			"message":      "🎉 Happy Birthday! Get your cheapest pizza + 1 free drink! 🎉",
		})
	}
}

func CreateDiscountCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	code := r.FormValue("code")
	var percentage int
	fmt.Sscanf(r.FormValue("percentage"), "%d", &percentage)

	if percentage < 1 || percentage > 100 {
		http.Error(w, "Percentage must be between 1 and 100", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO discount_code (code, discount_percentage, is_active) VALUES (?, ?, TRUE)`
	_, err := database.DATABASE.Exec(query, code, percentage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func UpdateDiscountCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	var id int
	fmt.Sscanf(r.FormValue("id"), "%d", &id)
	code := r.FormValue("code")
	var percentage int
	fmt.Sscanf(r.FormValue("percentage"), "%d", &percentage)
	isActive := r.FormValue("is_active") == "on"

	if percentage < 1 || percentage > 100 {
		http.Error(w, "Percentage must be between 1 and 100", http.StatusBadRequest)
		return
	}

	query := `UPDATE discount_code SET code = ?, discount_percentage = ?, is_active = ? WHERE id = ?`
	_, err := database.DATABASE.Exec(query, code, percentage, isActive, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func DeleteDiscountCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var id string
	if r.Method == http.MethodPost {
		r.ParseForm()
		id = r.FormValue("id")
	} else {
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM discount_code WHERE id = ?`
	_, err := database.DATABASE.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func AssignDeliveryPersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !isAdminFromHeaders(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	var orderID, deliveryPersonID int
	fmt.Sscanf(r.FormValue("order_id"), "%d", &orderID)
	fmt.Sscanf(r.FormValue("delivery_person_id"), "%d", &deliveryPersonID)

	query := `UPDATE orders SET delivery_person_id = ? WHERE id = ?`
	_, err := database.DATABASE.Exec(query, deliveryPersonID, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
