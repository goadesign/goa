package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/raphael/goa/examples/cellar/client"
	"gopkg.in/alecthomas/kingpin.v2"
)

type (
	// CreateAccountCommand is the command line data structure for the create action of account
	CreateAccountCommand struct {
		// Path is the HTTP request path.
		Path    string
		Payload string
	}
	// DeleteAccountCommand is the command line data structure for the delete action of account
	DeleteAccountCommand struct {
		// Path is the HTTP request path.
		Path string
	}
	// ShowAccountCommand is the command line data structure for the show action of account
	ShowAccountCommand struct {
		// Path is the HTTP request path.
		Path string
	}
	// UpdateAccountCommand is the command line data structure for the update action of account
	UpdateAccountCommand struct {
		// Path is the HTTP request path.
		Path    string
		Payload string
	}
	// CreateBottleCommand is the command line data structure for the create action of bottle
	CreateBottleCommand struct {
		// Path is the HTTP request path.
		Path    string
		Payload string
	}
	// DeleteBottleCommand is the command line data structure for the delete action of bottle
	DeleteBottleCommand struct {
		// Path is the HTTP request path.
		Path string
	}
	// ListBottleCommand is the command line data structure for the list action of bottle
	ListBottleCommand struct {
		// Path is the HTTP request path.
		Path string
		// Filter by years
		Years []int
	}
	// RateBottleCommand is the command line data structure for the rate action of bottle
	RateBottleCommand struct {
		// Path is the HTTP request path.
		Path    string
		Payload string
	}
	// ShowBottleCommand is the command line data structure for the show action of bottle
	ShowBottleCommand struct {
		// Path is the HTTP request path.
		Path string
	}
	// UpdateBottleCommand is the command line data structure for the update action of bottle
	UpdateBottleCommand struct {
		// Path is the HTTP request path.
		Path    string
		Payload string
	}
)

// Run makes the HTTP request corresponding to the CreateAccountCommand command.
func (cmd *CreateAccountCommand) Run(c *client.Client) (*http.Response, error) {
	var payload client.CreateAccountPayload
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize payload: %s", err)
		}
	}
	return c.CreateAccount(cmd.Path, &payload)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *CreateAccountCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, default is "/cellar/accounts"`).Default("/cellar/accounts").StringVar(&cmd.Path)
	cc.Flag("payload", "Request JSON body").StringVar(&cmd.Payload)
}

// Run makes the HTTP request corresponding to the DeleteAccountCommand command.
func (cmd *DeleteAccountCommand) Run(c *client.Client) (*http.Response, error) {
	return c.DeleteAccount(cmd.Path)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *DeleteAccountCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID`).Required().StringVar(&cmd.Path)
}

// Run makes the HTTP request corresponding to the ShowAccountCommand command.
func (cmd *ShowAccountCommand) Run(c *client.Client) (*http.Response, error) {
	return c.ShowAccount(cmd.Path)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *ShowAccountCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID`).Required().StringVar(&cmd.Path)
}

// Run makes the HTTP request corresponding to the UpdateAccountCommand command.
func (cmd *UpdateAccountCommand) Run(c *client.Client) (*http.Response, error) {
	var payload client.UpdateAccountPayload
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize payload: %s", err)
		}
	}
	return c.UpdateAccount(cmd.Path, &payload)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *UpdateAccountCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID`).Required().StringVar(&cmd.Path)
	cc.Flag("payload", "Request JSON body").StringVar(&cmd.Payload)
}

// Run makes the HTTP request corresponding to the CreateBottleCommand command.
func (cmd *CreateBottleCommand) Run(c *client.Client) (*http.Response, error) {
	var payload client.CreateBottlePayload
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize payload: %s", err)
		}
	}
	return c.CreateBottle(cmd.Path, &payload)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *CreateBottleCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID/bottles`).Required().StringVar(&cmd.Path)
	cc.Flag("payload", "Request JSON body").StringVar(&cmd.Payload)
}

// Run makes the HTTP request corresponding to the DeleteBottleCommand command.
func (cmd *DeleteBottleCommand) Run(c *client.Client) (*http.Response, error) {
	return c.DeleteBottle(cmd.Path)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *DeleteBottleCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID/bottles/:bottleID`).Required().StringVar(&cmd.Path)
}

// Run makes the HTTP request corresponding to the ListBottleCommand command.
func (cmd *ListBottleCommand) Run(c *client.Client) (*http.Response, error) {
	return c.ListBottle(cmd.Path, cmd.Years)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *ListBottleCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID/bottles`).Required().StringVar(&cmd.Path)
	cc.Flag("years", "Filter by years").IntsVar(&cmd.Years)
}

// Run makes the HTTP request corresponding to the RateBottleCommand command.
func (cmd *RateBottleCommand) Run(c *client.Client) (*http.Response, error) {
	var payload client.RateBottlePayload
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize payload: %s", err)
		}
	}
	return c.RateBottle(cmd.Path, &payload)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *RateBottleCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID/bottles/:bottleID/actions/rate`).Required().StringVar(&cmd.Path)
	cc.Flag("payload", "Request JSON body").StringVar(&cmd.Payload)
}

// Run makes the HTTP request corresponding to the ShowBottleCommand command.
func (cmd *ShowBottleCommand) Run(c *client.Client) (*http.Response, error) {
	return c.ShowBottle(cmd.Path)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *ShowBottleCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID/bottles/:bottleID`).Required().StringVar(&cmd.Path)
}

// Run makes the HTTP request corresponding to the UpdateBottleCommand command.
func (cmd *UpdateBottleCommand) Run(c *client.Client) (*http.Response, error) {
	var payload client.UpdateBottlePayload
	if cmd.Payload != "" {
		err := json.Unmarshal([]byte(cmd.Payload), &payload)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize payload: %s", err)
		}
	}
	return c.UpdateBottle(cmd.Path, &payload)
}

// RegisterFlags registers the command flags with the command line.
func (cmd *UpdateBottleCommand) RegisterFlags(cc *kingpin.CmdClause) {
	cc.Arg("path", `Request path, format is /cellar/accounts/:accountID/bottles/:bottleID`).Required().StringVar(&cmd.Path)
	cc.Flag("payload", "Request JSON body").StringVar(&cmd.Payload)
}
