package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateBottlePayload is the data structure used to initialize the bottle create request body.
type CreateBottlePayload struct {
	Color     string `json:"color"`
	Country   string `json:"country,omitempty"`
	Name      string `json:"name"`
	Region    string `json:"region,omitempty"`
	Review    string `json:"review,omitempty"`
	Sweetness int    `json:"sweetness,omitempty"`
	Varietal  string `json:"varietal"`
	Vineyard  string `json:"vineyard"`
	Vintage   int    `json:"vintage"`
}

// Record new bottle
func (c *Client) CreateBottle(path string, payload *CreateBottlePayload) (*http.Response, error) {
	var body io.Reader
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize body: %s", err)
	}
	body = bytes.NewBuffer(b)
	u := url.URL{Host: c.Host, Scheme: c.Scheme, Path: path}
	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// DeleteBottle makes a request to the delete action endpoint of the bottle resource
func (c *Client) DeleteBottle(path string) (*http.Response, error) {
	var body io.Reader
	u := url.URL{Host: c.Host, Scheme: c.Scheme, Path: path}
	req, err := http.NewRequest("DELETE", u.String(), body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// List all bottles in account optionally filtering by year
func (c *Client) ListBottle(path string, years []int) (*http.Response, error) {
	var body io.Reader
	u := url.URL{Host: c.Host, Scheme: c.Scheme, Path: path}
	values := u.Query()
	tmp12 := make([]string, len(years))
	for i, e := range years {
		tmp13 := strconv.Itoa(e)
		tmp12[i] = tmp13
	}
	tmp11 := strings.Join(tmp12, ",")
	values.Set("years", tmp11)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// RateBottlePayload is the data structure used to initialize the bottle rate request body.
type RateBottlePayload struct {
	// Rating of bottle between 1 and 5
	Rating int `json:"rating"`
}

// RateBottle makes a request to the rate action endpoint of the bottle resource
func (c *Client) RateBottle(path string, payload *RateBottlePayload) (*http.Response, error) {
	var body io.Reader
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize body: %s", err)
	}
	body = bytes.NewBuffer(b)
	u := url.URL{Host: c.Host, Scheme: c.Scheme, Path: path}
	req, err := http.NewRequest("PUT", u.String(), body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// Retrieve bottle with given id
func (c *Client) ShowBottle(path string) (*http.Response, error) {
	var body io.Reader
	u := url.URL{Host: c.Host, Scheme: c.Scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}

// UpdateBottlePayload is the data structure used to initialize the bottle update request body.
type UpdateBottlePayload struct {
	Color     string `json:"color,omitempty"`
	Country   string `json:"country,omitempty"`
	Name      string `json:"name,omitempty"`
	Region    string `json:"region,omitempty"`
	Review    string `json:"review,omitempty"`
	Sweetness int    `json:"sweetness,omitempty"`
	Varietal  string `json:"varietal,omitempty"`
	Vineyard  string `json:"vineyard,omitempty"`
	Vintage   int    `json:"vintage,omitempty"`
}

// UpdateBottle makes a request to the update action endpoint of the bottle resource
func (c *Client) UpdateBottle(path string, payload *UpdateBottlePayload) (*http.Response, error) {
	var body io.Reader
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize body: %s", err)
	}
	body = bytes.NewBuffer(b)
	u := url.URL{Host: c.Host, Scheme: c.Scheme, Path: path}
	req, err := http.NewRequest("PATCH", u.String(), body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	return c.Client.Do(req)
}
