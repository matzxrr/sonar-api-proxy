package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ContactFormSubmission represents the data from the Gravity Forms contact form
type ContactFormSubmission struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// config holds the service configuration
type config struct {
	sonarAPIURL      string
	sonarAPIToken    string
	port             string
	inboundMailboxID int
	ticketGroupID    int
	allowedOrigins   []string
}

func goDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Load .env file
	goDotEnv()

	// Load configuration
	cfg := loadConfig()

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/contact-form", corsMiddleware(cfg.allowedOrigins, createTicketHandler(cfg)))

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting proxy service on port %s...\n", cfg.port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func loadConfig() config {
	cfg := config{
		sonarAPIURL:   getEnvOrDefault("SONAR_API_URL", "https://kellin.sonar.software/api/graphql"),
		sonarAPIToken: getEnvOrDefault("SONAR_API_TOKEN", ""),
		port:          getEnvOrDefault("PORT", "8080"),
	}

	// Parse allowed origins
	allowedOrigins := getEnvOrDefault("ALLOWED_ORIGINS", "*")
	cfg.allowedOrigins = strings.Split(allowedOrigins, ",")

	// Parse inbound mailbox ID
	mailboxID := getEnvOrDefault("INBOUND_MAILBOX_ID", "1")
	if id, err := strconv.Atoi(mailboxID); err == nil {
		cfg.inboundMailboxID = id
	} else {
		log.Printf("Invalid INBOUND_MAILBOX_ID, using default: 1")
		cfg.inboundMailboxID = 1
	}

	// Parse ticket group ID
	ticketGroupID := getEnvOrDefault("TICKET_GROUP_ID", "3")
	if id, err := strconv.Atoi(ticketGroupID); err == nil {
		cfg.ticketGroupID = id
	} else {
		log.Printf("Invalid TICKET_GROUP_ID, using default: 3")
		cfg.ticketGroupID = 3
	}

	// Validate required configuration
	if cfg.sonarAPIToken == "" {
		log.Fatal("SONAR_API_TOKEN environment variable is required")
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	if err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func createTicketHandler(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the form submission
		var submission ContactFormSubmission
		if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
			log.Printf("Failed to decode request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if submission.Name == "" || submission.Email == "" || submission.Subject == "" || submission.Message == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Create the GraphQL mutation
		mutation := `
            mutation CreatePublicTicket($input: CreatePublicTicketMutationInput!) {
                createPublicTicket(input: $input) {
                    id
                    subject
                    description
                    status
                    created_at
                }
            }
        `

		// Build the description with all form data
		description := fmt.Sprintf(`Contact Form Submission

Name: %s
Email: %s
Phone: %s

Message:
%s`, submission.Name, submission.Email, submission.Phone, submission.Message)

		// Create variables for the mutation
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"subject":            submission.Subject,
				"description":        description,
				"status":             "OPEN",
				"priority":           "MEDIUM",
				"inbound_mailbox_id": cfg.inboundMailboxID,
				"ticket_group_id":    cfg.ticketGroupID,
				"ticket_recipients": []map[string]string{
					{
						"email_address": submission.Email,
						"name":          submission.Name,
					},
				},
				// Uncomment and set these if you want to associate with specific entities
				// "ticketable_type": "Account",
				// "ticketable_id": 123,
			},
		}

		graphQLReq := GraphQLRequest{
			Query:     mutation,
			Variables: variables,
		}

		// Send to Sonar API
		response, err := forwardToSonar(cfg, graphQLReq)
		if err != nil {
			log.Printf("Failed to create ticket in Sonar: %v", err)
			http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
			return
		}

		// Parse the response
		var graphQLResp GraphQLResponse
		if err := json.Unmarshal(response, &graphQLResp); err != nil {
			log.Printf("Failed to parse Sonar response: %v", err)
			http.Error(w, "Invalid response from Sonar", http.StatusInternalServerError)
			return
		}

		// Check for GraphQL errors
		if len(graphQLResp.Errors) > 0 {
			log.Printf("GraphQL errors: %v", graphQLResp.Errors)
			http.Error(w, "Failed to create ticket: "+graphQLResp.Errors[0].Message, http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Ticket created successfully",
			"data":    graphQLResp.Data,
		})
		if encodeErr != nil {
			log.Printf("Failed to write response body: %v", encodeErr)
		}
	}
}

func forwardToSonar(cfg config, graphQLReq GraphQLRequest) ([]byte, error) {
	jsonBody, err := json.Marshal(graphQLReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", cfg.sonarAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.sonarAPIToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sonar API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Middleware functions
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(allowedOrigins []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false

		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
