package models

// ContactForm represents a contact us submission
type ContactForm struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Reason        string `json:"reason"`
	Message       string `json:"message"`
}

// OutageReport represents an outage report submission
type OutageReport struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Reason        string `json:"reason"`
	Message       string `json:"message"`
}

// SignupRequest represents a new service signup
type SignupRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	StreetAddress string `json:"street_address"`
	Service       string `json:"service"`
	Message       string `json:"message"`
}

// VoIPSupport represents a VoIP support request
type VoIPSupport struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Message       string `json:"message"`
}
