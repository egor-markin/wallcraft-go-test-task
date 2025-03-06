package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/egor-markin/wallcraft-go-test-task/config"
	"github.com/egor-markin/wallcraft-go-test-task/database"
	"github.com/egor-markin/wallcraft-go-test-task/handlers"
	_ "github.com/lib/pq"
)

func main() {
	// Check if the DATABASE_URL environment variable is set
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Initialize the database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	// Test the database connection
	if _, err := db.Exec("SELECT 1"); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}

	// Initialize the query object
	queries := database.New(db)

	// Initialize handlers
	productHandler := &handlers.ProductHandler{Queries: queries}
	customerHandler := &handlers.CustomerHandler{Queries: queries}
	invoiceHandler := &handlers.InvoiceHandler{Queries: queries}

	// Routes
	http.HandleFunc(config.ProductsApiPrefix, productHandler.ProductsHandler)
	http.HandleFunc(config.ProductsApiPrefix+"/", productHandler.ProductHandler)
	http.HandleFunc(config.CustomersApiPrefix, customerHandler.CustomersHandler)
	http.HandleFunc(config.CustomersApiPrefix+"/", customerHandler.CustomerHandler)
	http.HandleFunc(config.InvoicesApiPrefix, invoiceHandler.InvoicesHandler)
	http.HandleFunc(config.InvoicesApiPrefix+"/", invoiceHandler.InvoiceHandler)

	// Health check endpoint
	http.HandleFunc(config.ApiPrefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database connection failed"))
			return
		}

		// If everything is fine, return 200 OK
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start the server
	log.Printf("The service is available at %s...", config.DefaultServiceBindingAddress)
	if err := http.ListenAndServe(config.DefaultServiceBindingAddress, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
