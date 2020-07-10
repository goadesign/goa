package openapiv3

import (
	"testing"

	"goa.design/goa/v3/expr"
)

func TestBuildInfo(t *testing.T) {
	const (
		title        = "test title"
		description  = "test description"
		terms        = "test terms of service"
		version      = "test version"
		contactName  = "test contact name"
		contactEmail = "test contact email"
		contactURL   = "test contact URL"
		licenseName  = "test license name"
		licenseURL   = "test license URL"
	)
	cases := []struct {
		Name           string
		Title          string
		Description    string
		TermsOfService string
		Version        string
		ContactName    string
		ContactEmail   string
		ContactURL     string
		LicenseName    string
		LicenseURL     string
	}{{
		Name:           "simple",
		Title:          title,
		Description:    description,
		TermsOfService: terms,
		Version:        version,
		ContactName:    contactName,
		ContactEmail:   contactEmail,
		ContactURL:     contactURL,
		LicenseName:    licenseName,
		LicenseURL:     licenseURL,
	}, {
		Name:  "empty version",
		Title: title,
	}, {
		Name:    "empty title",
		Version: version,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			api := &expr.APIExpr{
				Name:           c.Name,
				Title:          c.Title,
				Description:    c.Description,
				TermsOfService: c.TermsOfService,
				Version:        c.Version,
				Contact:        &expr.ContactExpr{Name: contactName, Email: contactEmail, URL: contactURL},
				License:        &expr.LicenseExpr{Name: licenseName, URL: licenseURL},
			}

			info := buildInfo(api)

			expected := c.Title
			if api.Title == "" {
				expected = "Goa API"
			}
			if info.Title != expected {
				t.Errorf("got API title %q, expected %q", info.Title, expected)
			}

			if info.Description != c.Description {
				t.Errorf("got API description %q, expected %q", info.Description, c.Description)
			}

			if info.TermsOfService != c.TermsOfService {
				t.Errorf("got API terms of service %q, expected %q", info.TermsOfService, c.TermsOfService)
			}

			expectedVer := c.Version
			if api.Version == "" {
				expectedVer = "1.0"
			}
			if info.Version != expectedVer {
				t.Errorf("got API version %q, expected %q", info.Version, expectedVer)
			}
		})
	}
}
