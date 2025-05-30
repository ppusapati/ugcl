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
	handlerWithCORS := enableCORS(handler)
	log.Println("Server starting at port", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Required CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Handle preflight (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
