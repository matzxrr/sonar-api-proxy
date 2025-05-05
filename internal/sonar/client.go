package sonar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type GraphQLResponse struct {
	Data   any `json:"data,omitempty"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

func NewClient(apiUrl, apiToken string) *Client {
	return &Client{
		apiURL: apiUrl,
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
) error {
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

	variables := map[string]any{
		"input": map[string]any{
			"subject": subject,
			"description": description,
			"status": "OPEN",
			"priority": priority,
			"inbound_mailbox_id": inboundMailboxID,
			"ticket_group_id": ticketGroupID,
			"ticket_recipients": []map[string]string {
				{
					"email_address": recipientEmail,
					"name": recipientName,
				},
			},
		},
	}

	req := GraphQLRequest{
		Query: mutation,
		Variables: variables,
	}

	response, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(response, &graphQLResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return fmt.Errorf("graphql error: %v", graphQLResp.Errors)
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
