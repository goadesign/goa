// These sources run beside generated services. Authentication callbacks record
// their inputs and the business methods record dispatch. The same error can
// arrive before or after a method runs, so transport retries cannot infer
// dispatch from the returned error, its name or its response status.
package generator

const authenticationRuntimeTest = `package gen_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"

	genprobe "generated.local/gen/auth_probe"
	genrpc "generated.local/gen/auth_rpc"
	genhttp "generated.local/gen/http/auth_probe/server"
	genrpcsrv "generated.local/gen/jsonrpc/auth_rpc/server"
	goagrpc "goa.design/goa/v3/grpc"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

type (
	authContextKey struct{}
	authProbe      struct {
		failures      map[string]error
		calls         []string
		credentials   []string
		scopes        [][]string
		dispatches    int
		methodError   error
		methodContext context.Context
	}
	rpcProbe    struct{ authProbe }
	nestedProbe struct {
		authProbe
		inner *authProbe
	}
)

func (s *authProbe) authenticate(ctx context.Context, name, credential string, scopes []string) (context.Context, error) {
	s.calls = append(s.calls, name)
	s.credentials = append(s.credentials, credential)
	s.scopes = append(s.scopes, scopes)
	return context.WithValue(ctx, authContextKey{}, name), s.failures[name]
}

func (s *authProbe) BasicAuth(ctx context.Context, user, pass string, scheme *security.BasicScheme) (context.Context, error) {
	return s.authenticate(ctx, scheme.Name, user+"/"+pass, scheme.RequiredScopes)
}
func (s *authProbe) APIKeyAuth(ctx context.Context, key string, scheme *security.APIKeyScheme) (context.Context, error) {
	return s.authenticate(ctx, scheme.Name, key, scheme.RequiredScopes)
}
func (s *authProbe) BearerAuth(ctx context.Context, token string, scheme *security.BearerScheme) (context.Context, error) {
	return s.authenticate(ctx, scheme.Name, token, scheme.RequiredScopes)
}
func (s *authProbe) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	return s.authenticate(ctx, scheme.Name, token, scheme.RequiredScopes)
}
func (s *authProbe) OAuth2Auth(ctx context.Context, token string, scheme *security.OAuth2Scheme) (context.Context, error) {
	return s.authenticate(ctx, scheme.Name, token, scheme.RequiredScopes)
}

func (s *authProbe) finish(ctx context.Context) (string, error) {
	s.dispatches++
	s.methodContext = ctx
	return "done", s.methodError
}
func (s *authProbe) Basic(ctx context.Context, _ *genprobe.BasicPayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) Key(ctx context.Context, _ *genprobe.KeyPayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) Bearer(ctx context.Context, _ *genprobe.BearerPayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) JWT(ctx context.Context, _ *genprobe.JWTPayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) Oauth(ctx context.Context, _ *genprobe.OauthPayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) Combined(ctx context.Context, _ *genprobe.CombinedPayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) Alternative(ctx context.Context, _ *genprobe.AlternativePayload) (string, error) {
	return s.finish(ctx)
}
func (s *authProbe) Open(ctx context.Context) (string, error) {
	return s.finish(ctx)
}
func (s *rpcProbe) Run(ctx context.Context, _ *genrpc.RunPayload) (string, error) {
	return s.finish(ctx)
}

func (s *nestedProbe) Basic(ctx context.Context, _ *genprobe.BasicPayload) (string, error) {
	s.dispatches++
	_, err := genprobe.NewEndpoints(s.inner).Oauth(ctx, &genprobe.OauthPayload{Token: "nested-token"})
	return "", err
}

func TestNestedAuthenticationPreservesOriginalError(t *testing.T) {
	failure := goa.PermanentError("denied", "inner credential rejected")
	inner := &authProbe{failures: map[string]error{"oauth": failure}}
	outer := &nestedProbe{inner: inner}
	_, err := genprobe.NewEndpoints(outer).Basic(t.Context(), &genprobe.BasicPayload{User: "user", Pass: "pass"})
	require.Equal(t, 1, outer.dispatches)
	require.Zero(t, inner.dispatches)
	require.ErrorIs(t, err, failure)
	require.Same(t, failure, err)
}

func TestMiddlewareAuthenticationPreservesOriginalError(t *testing.T) {
	failure := goa.PermanentError("denied", "inner credential rejected")
	inner := &authProbe{failures: map[string]error{"oauth": failure}}
	outer := &authProbe{}
	endpoints := genprobe.NewEndpoints(outer)
	endpoints.Use(func(next goa.Endpoint) goa.Endpoint {
		return func(ctx context.Context, payload any) (any, error) {
			result, err := next(ctx, payload)
			if err != nil {
				return result, err
			}
			_, err = genprobe.NewEndpoints(inner).Oauth(ctx, &genprobe.OauthPayload{Token: "nested-token"})
			return result, err
		}
	})
	_, err := endpoints.Basic(t.Context(), &genprobe.BasicPayload{User: "user", Pass: "pass"})
	require.Equal(t, 1, outer.dispatches)
	require.Zero(t, inner.dispatches)
	require.ErrorIs(t, err, failure)
	require.Same(t, failure, err)
}

func TestAuthenticationErrorsAndDispatch(t *testing.T) {
	failure := goa.PermanentError("denied", "credential rejected")
	for _, test := range []struct {
		name        string
		endpoint    func(*genprobe.Endpoints) goa.Endpoint
		payload     any
		calls       []string
		credentials []string
	}{
		{"basic", func(e *genprobe.Endpoints) goa.Endpoint {
			return e.Basic
		}, &genprobe.BasicPayload{User: "user", Pass: "pass"}, []string{"basic"}, []string{"user/pass"}},
		{"key", func(e *genprobe.Endpoints) goa.Endpoint {
			return e.Key
		}, &genprobe.KeyPayload{Key: "key-value"}, []string{"key"}, []string{"key-value"}},
		{"bearer", func(e *genprobe.Endpoints) goa.Endpoint {
			return e.Bearer
		}, &genprobe.BearerPayload{Token: "bearer-value"}, []string{"bearer"}, []string{"bearer-value"}},
		{"jwt", func(e *genprobe.Endpoints) goa.Endpoint {
			return e.JWT
		}, &genprobe.JWTPayload{Token: "jwt-value"}, []string{"jwt"}, []string{"jwt-value"}},
		{"oauth", func(e *genprobe.Endpoints) goa.Endpoint {
			return e.Oauth
		}, &genprobe.OauthPayload{Token: "access-value"}, []string{"oauth"}, []string{"access-value"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, cause := range []error{failure, fmt.Errorf("wrapped: %w", failure), errors.Join(failure)} {
				s := &authProbe{failures: map[string]error{test.name: cause}}
				result, err := test.endpoint(genprobe.NewEndpoints(s))(t.Context(), test.payload)
				require.Nil(t, result)
				require.Same(t, cause, err)
				require.ErrorIs(t, err, failure)
				require.Equal(t, cause.Error(), err.Error())
				require.Zero(t, s.dispatches)
				require.Equal(t, test.calls, s.calls)
				require.Equal(t, test.credentials, s.credentials)
				require.Equal(t, [][]string{{"read"}}, s.scopes)
				require.Equal(t, status.Code(goagrpc.EncodeError(cause)), status.Code(goagrpc.EncodeError(err)))
				s = &authProbe{methodError: cause}
				_, err = test.endpoint(genprobe.NewEndpoints(s))(t.Context(), test.payload)
				require.Same(t, cause, err)
				require.Equal(t, 1, s.dispatches)
				require.Equal(t, test.name, s.methodContext.Value(authContextKey{}))
			}
		})
	}
}

func TestRequirementsAndAlternatives(t *testing.T) {
	basicFailure := errors.New("basic rejected")
	jwtFailure := errors.New("jwt rejected")
	for _, test := range []struct {
		name         string
		combined     bool
		failures     map[string]error
		calls        []string
		cause        error
		contextValue string
	}{
		{"combined success", true, nil, []string{"basic", "jwt"}, nil, "jwt"},
		{"combined first rejection", true, map[string]error{"basic": basicFailure}, []string{"basic"}, basicFailure, ""},
		{"combined second rejection", true, map[string]error{"jwt": jwtFailure}, []string{"basic", "jwt"}, jwtFailure, ""},
		{"first alternative succeeds", false, nil, []string{"basic"}, nil, "basic"},
		{"second alternative succeeds", false, map[string]error{"basic": basicFailure}, []string{"basic", "jwt"}, nil, "jwt"},
		{"all alternatives reject", false, map[string]error{"basic": basicFailure, "jwt": jwtFailure}, []string{"basic", "jwt"}, jwtFailure, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := &authProbe{failures: test.failures}
			e := genprobe.NewEndpoints(s)
			endpoint := e.Alternative
			var payload any = &genprobe.AlternativePayload{User: "user", Pass: "pass", Token: "jwt-value"}
			if test.combined {
				endpoint = e.Combined
				payload = &genprobe.CombinedPayload{User: "user", Pass: "pass", Token: "jwt-value"}
			}
			result, err := endpoint(t.Context(), payload)
			require.Equal(t, test.calls, s.calls)
			for i, scopes := range s.scopes {
				want := []string{"read"}
				if !test.combined && s.calls[i] == "jwt" {
					want = []string{"write"}
				}
				require.Equal(t, want, scopes)
			}
			if test.cause != nil {
				require.ErrorIs(t, err, test.cause)
				require.Nil(t, result)
				require.Zero(t, s.dispatches)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "done", result)
			require.Equal(t, 1, s.dispatches)
			require.Equal(t, test.contextValue, s.methodContext.Value(authContextKey{}))
		})
	}
	s := &authProbe{failures: map[string]error{"basic": basicFailure}}
	result, err := genprobe.NewEndpoints(s).Open(t.Context(), nil)
	require.NoError(t, err)
	require.Equal(t, "done", result)
	require.Empty(t, s.calls)
	require.Equal(t, 1, s.dispatches)
}

func TestConfiguredAuthenticationCallback(t *testing.T) {
	failure := errors.New("configured callback rejected")
	for _, reject := range []bool{true, false} {
		s := &authProbe{}
		calls := 0
		endpoint := genprobe.NewOauthEndpoint(s, func(ctx context.Context, token string, scheme *security.OAuth2Scheme) (context.Context, error) {
			calls++
			require.Equal(t, "access-value", token)
			require.Equal(t, "oauth", scheme.Name)
			require.Equal(t, []string{"read"}, scheme.RequiredScopes)
			ctx = context.WithValue(ctx, authContextKey{}, "configured")
			if reject {
				return ctx, failure
			}
			return ctx, nil
		})
		result, err := endpoint(t.Context(), &genprobe.OauthPayload{Token: "access-value"})
		require.Equal(t, 1, calls)
		require.Empty(t, s.calls)
		if reject {
			require.ErrorIs(t, err, failure)
			require.Nil(t, result)
			require.Zero(t, s.dispatches)
		} else {
			require.NoError(t, err)
			require.Equal(t, "done", result)
			require.Equal(t, 1, s.dispatches)
			require.Equal(t, "configured", s.methodContext.Value(authContextKey{}))
		}
	}
}

func TestHTTPDeclaredErrorMapping(t *testing.T) {
	failure := goa.PermanentError("denied", "credential rejected")
	for _, authenticate := range []bool{true, false} {
		s := &authProbe{}
		if authenticate {
			s.failures = map[string]error{"basic": failure}
		} else {
			s.methodError = failure
		}
		_, err := genprobe.NewEndpoints(s).Basic(t.Context(), &genprobe.BasicPayload{User: "user", Pass: "pass"})
		require.ErrorIs(t, err, failure)
		response := httptest.NewRecorder()
		require.NoError(t, genhttp.EncodeBasicError(goahttp.ResponseEncoder, nil)(t.Context(), response, err))
		require.Equal(t, http.StatusUnauthorized, response.Code)
		var body goahttp.ErrorResponse
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
		require.Equal(t, failure.Name, body.Name)
		require.Equal(t, failure.ID, body.ID)
		require.Equal(t, failure.Message, body.Message)
	}
}

func TestJSONRPCDeclaredErrorMapping(t *testing.T) {
	failure := goa.PermanentError("denied", "credential rejected")
	for _, authenticate := range []bool{true, false} {
		s := &rpcProbe{}
		if authenticate {
			s.failures = map[string]error{"oauth": failure}
		} else {
			s.methodError = failure
		}
		server := genrpcsrv.New(genrpc.NewEndpoints(s), goahttp.NewMuxer(), goahttp.RequestDecoder, goahttp.ResponseEncoder, func(_ context.Context, _ http.ResponseWriter, err error) {
			t.Error(err)
		})
		request := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":\"auth-1\",\"method\":\"run\",\"params\":{\"access\":\"value\"}}"))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer value")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		var body struct {
			ID    string
			Error *jsonrpc.RawErrorResponse
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
		require.Equal(t, "auth-1", body.ID)
		require.NotNil(t, body.Error)
		require.Equal(t, -32043, body.Error.Code)
		require.Equal(t, failure.Message, body.Error.Message)
		name, encoded, ok := jsonrpc.DecodeServiceErrorData(body.Error.Data)
		require.True(t, ok)
		require.Equal(t, failure.Name, name)
		var original goa.ServiceError
		require.NoError(t, json.Unmarshal(encoded, &original))
		require.Equal(t, failure.ID, original.ID)
		if authenticate {
			require.Zero(t, s.dispatches)
		} else {
			require.Equal(t, 1, s.dispatches)
		}
	}
}
`
