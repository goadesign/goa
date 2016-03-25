package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/spf13/cobra"
)

type (
	// Signer is the common interface implemented by all signers.
	Signer interface {
		// Sign adds required headers, cookies etc.
		Sign(context.Context, *http.Request) error
		// RegisterFlags registers the command line flags that defines the values used to
		// initialize the signer.
		RegisterFlags(cmd *cobra.Command)
	}

	// BasicSigner implements basic auth.
	BasicSigner struct {
		// Username is the basic auth user.
		Username string
		// Password is err guess what? the basic auth password.
		Password string
	}

	// JWTSigner implements JSON Web Token auth.
	JWTSigner struct {
		// Header is the name of the HTTP header which contains the JWT.
		// The default is "Authentication"
		Header string
		// Format represents the format used to render the JWT.
		// The default is "Bearer %s"
		Format string

		// token stores the actual JWT.
		token string
	}

	// OAuth2Signer enables the use of OAuth2 refresh tokens. It takes care of creating access
	// tokens given a refresh token and a refresh URL as defined in RFC 6749.
	// Note that this signer does not concern itself with generating the initial refresh token,
	// this has to be done prior to using the client.
	// Also it assumes the response of the refresh request response is JSON encoded and of the
	// form:
	// 	{
	//		"access_token":"2YotnFZFEjr1zCsicMWpAA",
	// 		"expires_in":3600,
	// 		"refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA"
	// 	}
	// where the "expires_in" and "refresh_token" properties are optional and additional
	// properties are ignored. If the response contains a "expires_in" property then the signer
	// takes care of making refresh requests prior to the token expiration.
	OAuth2Signer struct {
		// RefreshURLFormat is a format that generates the refresh access token URL given a
		// refresh token.
		RefreshURLFormat string
		// RefreshToken contains the OAuth3 refresh token from which access tokens are
		// created.
		RefreshToken string

		// accessToken is the temporary access token.
		accessToken string
		// expiresAt specifies when to create a new access token.
		expiresAt time.Time
	}
)

// Sign adds the basic auth header to the request.
func (s *BasicSigner) Sign(ctx context.Context, req *http.Request) error {
	if s.Username != "" && s.Password != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	return nil
}

// RegisterFlags adds the "--user" and "--pass" flags to the client tool.
func (s *BasicSigner) RegisterFlags(app *cobra.Command) {
	app.Flags().StringVar(&s.Username, "user", "", "Basic Auth username")
	app.Flags().StringVar(&s.Password, "pass", "", "Basic Auth password")
}

// Sign adds the JWT auth header.
func (s *JWTSigner) Sign(ctx context.Context, req *http.Request) error {
	header := s.Header
	if header == "" {
		header = "Authorization"
	}
	format := s.Format
	if format == "" {
		format = "Bearer %s"
	}
	req.Header.Set(header, fmt.Sprintf(format, s.token))
	return nil
}

// RegisterFlags adds the "--jwt" flag to the client tool.
func (s *JWTSigner) RegisterFlags(app *cobra.Command) {
	app.Flags().StringVar(&s.token, "jwt", "", "JSON web token")
}

// Sign refreshes the access token if needed and adds the OAuth header.
func (s *OAuth2Signer) Sign(ctx context.Context, req *http.Request) error {
	if s.expiresAt.Before(time.Now()) {
		if err := s.Refresh(ctx); err != nil {
			return fmt.Errorf("failed to refresh OAuth token: %s", err)
		}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	return nil
}

// RegisterFlags adds the "--refreshURL" and "--refreshToken" flags to the client tool.
func (s *OAuth2Signer) RegisterFlags(app *cobra.Command) {
	app.Flags().StringVar(&s.RefreshURLFormat, "refreshURL", "", "OAuth2 refresh URL format, e.g. https://somewhere.com/token?grant_type=authorization_code&code=%s&client_id=xxx")
	app.Flags().StringVar(&s.RefreshToken, "refreshToken", "", "OAuth2 refresh token or authorization code")
}

// ouath2RefreshResponse is the data structure representing the interesting subset of a OAuth2
// refresh response.
type oauth2RefreshResponse struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	AccessToken  string `json:"access_token"`
}

// Refresh makes a OAuth2 refresh access token request.
func (s *OAuth2Signer) Refresh(ctx context.Context) error {
	url := fmt.Sprintf(s.RefreshURLFormat, s.RefreshToken)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	id := shortID()
	goa.LogInfo(ctx, "refresh", "id", id, "url", fmt.Sprintf(s.RefreshURLFormat, "*****"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		goa.LogError(ctx, "failed", "id", id, "err", err)
		return err
	}
	goa.LogInfo(ctx, "completed", "id", id, "status", resp.Status)
	defer resp.Body.Close()
	var r oauth2RefreshResponse
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}
	err = json.Unmarshal(respBody, &r)
	if err != nil {
		return fmt.Errorf("failed to decode refresh request response: %s", err)
	}
	s.accessToken = r.AccessToken
	if r.ExpiresIn > 0 {
		s.expiresAt = time.Now().Add(time.Duration(r.ExpiresIn) * time.Second)
		goa.LogInfo(ctx, "refreshed", "expires", s.expiresAt)
	}
	if r.RefreshToken != "" {
		s.RefreshToken = r.RefreshToken
	}
	return nil
}
