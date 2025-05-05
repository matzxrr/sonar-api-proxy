package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"kellin/sonar-api-proxy/internal/models"
	"kellin/sonar-api-proxy/internal/sonar"
)

type OutageHandler struct {
	sonarClient *sonar.Client
}

func NewOutageHandler(client *sonar.Client) *OutageHandler {
	return &OutageHandler{
		sonarClient: client,
	}
}


const OutageTemplate = `
<p>
	<strong>First Name:</strong> %s<br>
	<strong>Last Name:</strong> %s<br>
	<strong>Phone:</strong> %s<br>
	<strong>Email:</strong> %s<br>
	<strong>Street Address:</strong> %s<br>
	<strong>City:</strong> %s<br>
	<strong>State:</strong> %s<br>
	<strong>Reason:</strong> %s<br>
	<strong>Message:</strong><br>
	%s
</p>
`

func (h *OutageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form models.OutageReport
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	
	// Build ticket content
	subject := fmt.Sprintf("Outage Form: %s", form.Reason)
	description := fmt.Sprintf(
		DescriptionTemplate,
		form.FirstName,
		form.LastName,
		form.Phone,
		form.Email,
		form.StreetAddress,
		form.City,
		form.State,
		form.Reason,
		form.Message,
	)
	fullName := fmt.Sprintf("%s %s", form.FirstName, form.LastName)

	// Create ticket with HIGH priority
	ticketID, err := h.sonarClient.CreateTicket(
		subject,
		description,
		"HIGH",
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
		"message": "Outage report submitted successfully",
		"ticket_id": ticketID,
	})
}
