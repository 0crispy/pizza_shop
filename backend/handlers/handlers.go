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

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	success, _, _, err := isLoginOK(r)

	msg := ""
	if err != nil {
		msg = err.Error()
	}

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
