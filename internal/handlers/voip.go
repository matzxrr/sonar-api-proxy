package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"kellin/sonar-api-proxy/internal/models"
	"kellin/sonar-api-proxy/internal/sonar"
)

type VoIPHandler struct {
	sonarClient *sonar.Client
}

func NewVoIPHandler(client *sonar.Client) *VoIPHandler {
	return &VoIPHandler{
		sonarClient: client,
	}
}

const VoipTemplate = `
<p>
	<strong>First Name:</strong> %s<br>
	<strong>Last Name:</strong> %s<br>
	<strong>Email:</strong> %s<br>
	<strong>Street Address:</strong> %s<br>
	<strong>City:</strong> %s<br>
	<strong>State:</strong> %s<br>
	<strong>Message:</strong><br>
	%s
</p>
`

func (h *VoIPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form models.VoIPSupport
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build ticket content
	subject := "VoIP Form"
	description := fmt.Sprintf(
		VoipTemplate,
		form.FirstName,
		form.LastName,
		form.Email,
		form.StreetAddress,
		form.City,
		form.State,
		form.Message,
	)
	fullName := fmt.Sprintf("%s %s", form.FirstName, form.LastName)
	
	// Create ticket
	ticketID, err := h.sonarClient.CreateTicket(
		subject,
		description,
		"MEDIUM",
		3,  // Support ticket group ID
		1,  // Support inbound mailbox ID
		form.Email,
		fullName,
	)
	if err != nil {
		http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

    if err := h.sonarClient.ResendAutoreply(ticketID, form.Email); err != nil {
        // Log the error but don't fail the request
        log.Printf("Warning: Failed to send autoreply for ticket %d: %v", ticketID, err)
    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "VoIP support request submitted successfully",
		"ticket_id": ticketID,
	})
}
