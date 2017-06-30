// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// HTTP request path constructors for the account service.
//
// Command:
// $ goa server goa.design/goa.v2/examples/account/design

package transport

import (
	"fmt"
)

// CreateAccountPath returns the URL path to the account service create HTTP endpoint.
func CreateAccountPath(orgID uint) string {
	return fmt.Sprintf("/orgs/%v/accounts", orgID)
}

// ListAccountPath returns the URL path to the account service list HTTP endpoint.
func ListAccountPath(orgID uint) string {
	return fmt.Sprintf("/orgs/%v/accounts", orgID)
}

// ShowAccountPath returns the URL path to the account service show HTTP endpoint.
func ShowAccountPath(orgID uint, id string) string {
	return fmt.Sprintf("/orgs/%v/accounts/%v", orgID, id)
}

// DeleteAccountPath returns the URL path to the account service delete HTTP endpoint.
func DeleteAccountPath(orgID uint, id string) string {
	return fmt.Sprintf("/orgs/%v/accounts/%v", orgID, id)
}
