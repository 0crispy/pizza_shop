package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	database "pizza_shop/backend/database"
	"sort"
	"strings"
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

// Admin UI page (basic panel)
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/admin.html")
	if err != nil {
		http.Error(w, "Admin UI not found", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(html_string))
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	success, username, _, err := isLoginOK(r)

	msg := ""
	if err != nil {
		msg = err.Error()
	}

	isAdmin := false
	if success {
		if role, rerr := database.GetUserRole(username); rerr == nil && role == database.AdminRole.String() {
			isAdmin = true
		}
	}

	type Msg struct {
		Ok      bool   `json:"ok"`
		Msg     string `json:"msg"`
		IsAdmin bool   `json:"isAdmin"`
	}
	sendMsg := Msg{success, msg, isAdmin}
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

// Admin auth via headers (used for DELETE where body is empty)
func isAdminFromHeaders(r *http.Request) bool {
	user := r.Header.Get("X-Username")
	pass := r.Header.Get("X-Password")
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
func isAdminRequest(r *http.Request) (bool, string) {
	var a adminAuth
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		return false, "invalid json"
	}
	ok, _ := database.TryLogin(a.Username, a.Password)
	if !ok {
		return false, "invalid credentials"
	}
	role, err := database.GetUserRole(a.Username)
	if err != nil || role != database.AdminRole.String() {
		return false, "not admin"
	}
	return true, ""
}

// POST /admin/ingredient/create {username,password,name,costCents,hasMeat,hasAnimalProducts}
func AdminCreateIngredientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// decode whole body once into map to reuse auth and payload
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// Re-marshal/re-decode to leverage existing adminAuth and payload structs is overkill; parse directly
	a := adminAuth{Username: payload["username"].(string), Password: payload["password"].(string)}
	ok, msg := func() (bool, string) { // inline admin check using payload values
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
	name, _ := payload["name"].(string)
	// cost as float or number -> convert to cents int64
	var costCents int64
	switch v := payload["costCents"].(type) {
	case float64:
		costCents = int64(v)
	case int64:
		costCents = v
	default:
		http.Error(w, "invalid costCents", http.StatusBadRequest)
		return
	}
	hasMeat, _ := payload["hasMeat"].(bool)
	hasAnimal, _ := payload["hasAnimalProducts"].(bool)
	ingr := database.NewIngredient(name, costCents, hasMeat, hasAnimal)
	created, err := database.CreateIngredient(ingr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(created)
}

// GET /admin/ingredient/list
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

// DELETE /admin/ingredient/delete?id=123
func AdminDeleteIngredientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isAdminFromHeaders(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Expect id in query param.
	ids := r.URL.Query().Get("id")
	if ids == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	var id int
	fmt.Sscanf(ids, "%d", &id)
	if err := database.DeleteIngredient(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /admin/pizza/create {username,password,name,ingredients:["Mozzarella","Tomato sauce"]}
func AdminCreatePizzaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
	pizza, err := database.CreatePizza(payload.Name, payload.Ingredients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(pizza)
}

// GET /admin/pizza/list
func AdminListPizzasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pizzas, err := database.GetAllPizzas()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(pizzas)
}

// DELETE /admin/pizza/delete?id=123
func AdminDeletePizzaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isAdminFromHeaders(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ids := r.URL.Query().Get("id")
	if ids == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	var id int
	fmt.Sscanf(ids, "%d", &id)
	if err := database.DeletePizza(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
