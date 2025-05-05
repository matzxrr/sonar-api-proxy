package sonar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	apiURL   string
	apiToken string
	client   *http.Client
}

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// General GraphQL response structure for error handling
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// CreateTicketResponse represents the response structure for the createPublicTicket mutation
type CreateTicketResponse struct {
	CreatePublicTicket struct {
		ID        string `json:"id"`
		Subject   string `json:"subject"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	} `json:"createPublicTicket"`
}

// ResendAutoreplyResponse represents the response structure for the resendAutoreply mutation
type ResendAutoreplyResponse struct {
	ResendAutoreply struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	} `json:"resendAutoreply"`
}

func NewClient(apiUrl, apiToken string) *Client {
	return &Client{
		apiURL:   apiUrl,
		apiToken: apiToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreateTicket(
	subject, description, priority string,
	ticketGroupID, inboundMailboxID int,
	recipientEmail, recipientName string,
) (int, error) {
	mutation := `
		mutation CreatePublicTicket($input: CreatePublicTicketMutationInput!) {
			createPublicTicket(input: $input) {
				id
				subject
				status
				created_at
			}
		}		
	`

	variables := map[string]any{
		"input": map[string]any{
			"subject":            subject,
			"description":        description,
			"status":             "OPEN",
			"priority":           priority,
			"inbound_mailbox_id": inboundMailboxID,
			"ticket_group_id":    ticketGroupID,
			"ticket_recipients": []map[string]string{
				{
					"email_address": recipientEmail,
					"name":          recipientName,
				},
			},
		},
	}

	req := GraphQLRequest{
		Query:     mutation,
		Variables: variables,
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}

	// First, extract any GraphQL errors
	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(response, &graphQLResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return 0, fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	// Then parse the actual data
	var createTicketResp CreateTicketResponse
	if err := json.Unmarshal(graphQLResp.Data, &createTicketResp); err != nil {
		return 0, fmt.Errorf("failed to parse ticket data: %w", err)
	}

	ticketID, err := strconv.Atoi(createTicketResp.CreatePublicTicket.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to convert ticket ID to integer: %w", err)
	}

	return ticketID, nil
}

// ResendAutoreply sends the autoreply notification for a ticket
func (c *Client) ResendAutoreply(ticketID int, toEmailAddress string) error {
	mutation := `
		mutation ResendAutoreply($input: ResendAutoreplyMutationInput!) {
			resendAutoreply(input: $input) {
				success
				message
			}
		}
	`

	variables := map[string]any{
		"input": map[string]any{
			"ticket_id":        ticketID,
			"to_email_address": toEmailAddress,
		},
	}

	req := GraphQLRequest{
		Query:     mutation,
		Variables: variables,
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send autoreply request: %w", err)
	}

	// First, extract any GraphQL errors
	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(response, &graphQLResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	// Then parse the actual data
	var resendAutoreplyResp ResendAutoreplyResponse
	if err := json.Unmarshal(graphQLResp.Data, &resendAutoreplyResp); err != nil {
		return fmt.Errorf("failed to parse autoreply data: %w", err)
	}

	if !resendAutoreplyResp.ResendAutoreply.Success {
		return fmt.Errorf("failed to resend autoreply: %s", resendAutoreplyResp.ResendAutoreply.Message)
	}

	return nil
}

func (c *Client) sendRequest(graphQLReq GraphQLRequest) ([]byte, error) {
	jsonBody, err := json.Marshal(graphQLReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
