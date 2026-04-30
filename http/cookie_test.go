package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieIntAttr(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want *int
	}{
		{"zero-yields-nil", 0, nil},
		{"positive-yields-pointer", 60, intPtr(60)},
		{"negative-yields-pointer", -1, intPtr(-1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CookieIntAttr(c.in)
			if c.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *c.want, *got)
		})
	}

	t.Run("works-for-int64", func(t *testing.T) {
		require.Nil(t, CookieIntAttr[int64](0))
		got := CookieIntAttr[int64](42)
		require.NotNil(t, got)
		require.EqualValues(t, 42, *got)
	})

	t.Run("works-for-uint", func(t *testing.T) {
		require.Nil(t, CookieIntAttr[uint](0))
		got := CookieIntAttr[uint](7)
		require.NotNil(t, got)
		require.EqualValues(t, 7, *got)
	})
}

func TestCookieStringAttr(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want *string
	}{
		{"empty-yields-nil", "", nil},
		{"non-empty-yields-pointer", "example.com", strPtr("example.com")},
		{"whitespace-yields-pointer", " ", strPtr(" ")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CookieStringAttr(c.in)
			if c.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *c.want, *got)
		})
	}
}

func TestCookieBoolAttr(t *testing.T) {
	cases := []struct {
		name string
		in   bool
		want *bool
	}{
		{"false-yields-nil", false, nil},
		{"true-yields-pointer", true, boolPtr(true)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CookieBoolAttr(c.in)
			if c.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *c.want, *got)
		})
	}
}

func TestCookieSameSiteAttr(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want *string
	}{
		{"empty-yields-nil", "", nil},
		{"default-yields-nil", "default", nil},
		{"strict-yields-pointer", "strict", strPtr("strict")},
		{"lax-yields-pointer", "lax", strPtr("lax")},
		{"none-yields-pointer", "none", strPtr("none")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CookieSameSiteAttr(c.in)
			if c.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *c.want, *got)
		})
	}
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
func boolPtr(v bool) *bool    { return &v }
