package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	database "pizza_shop/backend/database"
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
	type User struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	success, msg := database.TryLogin(user.Username, user.Password)

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
	log.Println(success, msg)

	type Msg struct {
		Ok  bool   `json:"ok"`
		Msg string `json:"msg"`
	}

}

func PizzaHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/pizza.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}
