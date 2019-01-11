package xray

import (
	"fmt"
	"os"
	"strings"
	"sync"

	pkgerrors "github.com/pkg/errors"
)

type (
	// TestClientExpectation is a generic mock.
	TestClientExpectation struct {
		mu         sync.Mutex
		expected   expectations
		unexpected []string
	}

	// expectations is the data structure used to record expected function
	// calls and the corresponding behavior.
	expectations map[string][]interface{}
)

// NewTestClientExpectation creates a new *TestClientExpectation
func NewTestClientExpectation() *TestClientExpectation {
	return &TestClientExpectation{
		mu:       sync.Mutex{},
		expected: make(expectations),
	}
}

// Expect records the request handler in the list of expected request calls.
func (c *TestClientExpectation) Expect(fn string, e interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expected[fn] = append(c.expected[fn], e)
}

// ExpectNTimes records the request handler n times in the list of expected request calls.
func (c *TestClientExpectation) ExpectNTimes(n int, fn string, e interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < n; i++ {
		c.expected[fn] = append(c.expected[fn], e)
	}
}

// Expectation removes the expectation for the function with the given name from the expected calls
// if there is one and returns it. If there is no (more) expectations for the function,
// it prints a warning to stderr and returns nil.
func (c *TestClientExpectation) Expectation(fn string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	es, ok := c.expected[fn]
	if !ok {
		err := pkgerrors.New("!!! Expectation not found for: " + fn)
		fmt.Fprintf(os.Stderr, "\n%+v\n", err)
		c.unexpected = append(c.unexpected, fn)
		return nil
	}
	e := es[0]
	if len(es) == 1 {
		delete(c.expected, fn)
	} else {
		c.expected[fn] = c.expected[fn][1:]
	}
	return e
}

// MetExpectations returns nil if there no expectation left to be called and if there is no call
// that was made that did not match an expectation. It returns an error describing what is left to
// be called or what was called with no expectation otherwise.
func (c *TestClientExpectation) MetExpectations() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var msg string
	if len(c.unexpected) > 0 {
		msg = fmt.Sprintf("%s was called but wasn't expected.", strings.Join(c.unexpected, ", "))
	}
	if len(c.expected) > 0 {
		if len(msg) > 0 {
			msg += "\n"
		}
		i := 0
		keys := make([]string, len(c.expected))
		for e := range c.expected {
			keys[i] = e
			i++
		}
		msg += fmt.Sprintf("%s was expected to be called but wasn't.", strings.Join(keys, ", "))
	}
	if msg == "" {
		return nil
	}
	return fmt.Errorf(msg)
}
