package goa

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceErrorUnwrap(t *testing.T) {
	var (
		errFoo          = errors.New("foo")
		errBar          = errors.New("bar")
		serviceErrorFoo = NewServiceError(errFoo, "foo", false, false, false)
		serviceErrorBar = NewServiceError(errBar, "bar", false, false, false)
	)
	cases := map[string]struct {
		err  error
		want error
	}{
		"service error": {
			err:  serviceErrorFoo,
			want: errFoo,
		},
		"merged service error": {
			err:  MergeErrors(serviceErrorFoo, serviceErrorBar),
			want: errors.Join(errFoo, errBar),
		},
	}
	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			got := errors.Unwrap(tc.err)
			if errs, ok := tc.want.(interface{ Unwrap() []error }); ok {
				for _, e := range errs.Unwrap() {
					if !errors.Is(got, e) {
						t.Errorf("got %#v, want %#v", got, tc.want)
					}
				}
			} else if !errors.Is(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestAsError(t *testing.T) {
	err := MissingFieldError("foo", "bar")
	se := asError(err)
	if !errors.Is(err, se) {
		t.Errorf("got %#v, want %#v", se, err)
	}
}

func TestInvalidLengthErrorDoesNotIncludeRejectedValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target any
		length int
		bound  int
		min    bool
		want   string
		secret string
	}{
		{
			name:   "string",
			target: strings.Repeat("string-secret", 1000),
			length: 13000,
			bound:  8,
			want:   "length of payload must be at most 8 but got 13000",
			secret: "string-secret",
		},
		{
			name:   "bytes",
			target: bytes.Repeat([]byte{0xab}, 1<<20),
			length: 1 << 20,
			bound:  1024,
			want:   "length of payload must be at most 1024 but got 1048576",
			secret: "0xab",
		},
		{
			name:   "array",
			target: []string{"array-secret", "array-secret", "array-secret"},
			length: 3,
			bound:  4,
			min:    true,
			want:   "length of payload must be at least 4 but got 3",
			secret: "array-secret",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := InvalidLengthError("payload", test.target, test.length, test.bound, test.min)
			require.EqualError(t, err, test.want)
			require.NotContains(t, err.Error(), test.secret)
			require.Less(t, len(err.Error()), 128)
			var serviceErr *ServiceError
			require.ErrorAs(t, err, &serviceErr)
			require.Equal(t, InvalidLength, serviceErr.Name)
			require.NotNil(t, serviceErr.Field)
			require.Equal(t, "payload", *serviceErr.Field)
		})
	}
}
