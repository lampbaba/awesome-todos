package main

import (
	"awesome/todos/app"
	"log"
	"net/http"
)

func main() {
	appHandler := app.MakeHandler("./test.db")
	defer appHandler.Close()

	log.Println("Started App")

	err := http.ListenAndServe(":3000", appHandler)
	if err != nil {
		panic(err)
	}
}
