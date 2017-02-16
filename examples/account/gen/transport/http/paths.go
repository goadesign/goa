package http

import "fmt"

// CreateAccountPath returns the URL path to the account service create HTTP
// endpoint.
func CreateAccountPath() string {
	return "/accounts"
}

// ListAccountPath returns the URL path to the account service list HTTP
// endpoint.
func ListAccountPath() string {
	return "/accounts"
}

// ShowAccountPath returns the URL path to the account service show HTTP
// endpoint.
func ShowAccountPath(id string) string {
	return fmt.Sprintf("/accounts/%v", id)
}

// DeleteAccountPath returns the URL path to the account service delete HTTP
// endpoint.
func DeleteAccountPath(id string) string {
	return fmt.Sprintf("/accounts/%v", id)
}
