package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"kellin/sonar-api-proxy/internal/handlers"
	"kellin/sonar-api-proxy/internal/sonar"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Configuration
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	apiURL := os.Getenv("SONAR_API_URL")
	if apiURL == "" {
		apiURL = "https://kellin.sonar.software/api/graphql"
	}

	apiToken := os.Getenv("SONAR_API_TOKEN")
	if apiToken == "" {
		log.Fatal("SONAR_API_TOKEN environment variable is required")
	}

	// Create Sonar client
	sonarClient := sonar.NewClient(apiURL, apiToken)

	// Create handlers
	contactHandler := handlers.NewContactHandler(sonarClient)
	outageHandler := handlers.NewOutageHandler(sonarClient)
	signupHandler := handlers.NewSignupHandler(sonarClient)
	voipHandler := handlers.NewVoIPHandler(sonarClient)

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/contact-us", contactHandler.Handle)
	mux.HandleFunc("/api/v1/report-an-outage", outageHandler.Handle)
	mux.HandleFunc("/api/v1/sign-up", signupHandler.Handle)
	mux.HandleFunc("/api/v1/voip-support", voipHandler.Handle)

	// Start server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting server on port %s...\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Add panic recovery
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v", err)
                http.Error(w, "Internal server error", http.StatusInternalServerError)
            }
        }()
        
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
