package goa

import (
	"errors"
	"testing"
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
