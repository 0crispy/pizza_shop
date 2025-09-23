package handlers

import (
	"fmt"
	"net/http"
	"os"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/index.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}

func PizzaHandler(w http.ResponseWriter, r *http.Request) {
	html_string, err := os.ReadFile("frontend/pizza.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(html_string))
}
