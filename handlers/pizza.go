package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html_string, err := os.ReadFile("frontend/index.html")
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(w, string(html_string))
	})

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
