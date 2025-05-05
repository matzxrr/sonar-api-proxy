package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"kellin/sonar-api-proxy/internal/models"
	"kellin/sonar-api-proxy/internal/sonar"
)

type SignupHandler struct {
	sonarClient *sonar.Client
}

func NewSignupHandler(client *sonar.Client) *SignupHandler {
	return &SignupHandler{
		sonarClient: client,
	}
}

const SignupTemplate = `
<p>
	<strong>First Name:</strong> %s<br>
	<strong>Last Name:</strong> %s<br>
	<strong>Phone:</strong> %s<br>
	<strong>Email:</strong> %s<br>
	<strong>Street Address:</strong> %s<br>
	<strong>Service:</strong> %s<br>
	<strong>Message:</strong><br>
	%s
</p>
`

func (h *SignupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build ticket content
	subject := fmt.Sprintf("Signup Form: %s", form.Service)
	description := fmt.Sprintf(
		SignupTemplate,
		form.FirstName,
		form.LastName,
		form.Phone,
		form.Email,
		form.StreetAddress,
		form.Service,
		form.Message,
	)
	fullName := fmt.Sprintf("%s %s", form.FirstName, form.LastName)

	// Create ticket with HIGH priority (new business!)
	err := h.sonarClient.CreateTicket(
		subject,
		description,
		"HIGH",
		2,  // Sales ticket group ID
		1,  // Support inbound mailbox ID
		form.Email,
		fullName,
	)
	if err != nil {
		http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Signup request submitted successfully",
	})
}
