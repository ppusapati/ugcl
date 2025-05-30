package main

import (
	"log"
	"net/http"
	"os"

	"p9e.in/ugcl/config"
	"p9e.in/ugcl/routes"
)

func main() {
	config.Connect()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := config.Migrations(config.DB); err != nil {
		log.Fatalf("could not run migrations: %v", err)
	}
	handler := routes.RegisterRoutes()
	log.Println("Server starting at port", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
