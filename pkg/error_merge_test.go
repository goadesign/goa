// These tests exercise merges through the runtime API. They verify unchanged
// inputs, original contribution values, and finite traversal of existing causes
// when inputs share or refer to one another.
package goa

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	// mergeSelectionError supplies a selected ServiceError through a custom As
	// method without putting that error in an Unwrap chain.
	mergeSelectionError struct {
		selected *ServiceError
	}

	// mergeOpaqueCause reads its original input's fields from shallow Error,
	// Is and As methods. Merging that input must preserve this concrete object
	// and the field values those methods observe.
	mergeOpaqueCause struct {
		input *ServiceError
		match error
		view  *mergeCauseView
	}

	mergeCauseView struct {
		message string
	}

	mergeMissingCause struct{}
)

func TestMergeErrorsNilPassThrough(t *testing.T) {
	service := NewServiceError(context.Canceled, "canceled", false, false, false)
	wrapped := fmt.Errorf("wrapper: %w", service)
	raw := errors.New("raw")
	for _, input := range []error{nil, service, wrapped, raw} {
		require.True(t, input == MergeErrors(nil, input)) //nolint:errorlint // Nil operands must return the exact input object.
		require.True(t, input == MergeErrors(input, nil)) //nolint:errorlint // Nil operands must return the exact input object.
	}
	require.Empty(t, service.history)
}

func TestMergeErrorsWholeValuesAndOriginalContributions(t *testing.T) {
	for _, name := range []string{"left", "error", "same"} {
		for leftTraits := 0; leftTraits < 8; leftTraits++ {
			for rightTraits := 0; rightTraits < 8; rightTraits++ {
				t.Run(fmt.Sprintf("%s/%d/%d", name, leftTraits, rightTraits), func(t *testing.T) {
					leftCause := errors.New("same message")
					rightCause := errors.New("same message")
					left := NewServiceError(leftCause, name, leftTraits&1 != 0, leftTraits&2 != 0, leftTraits&4 != 0)
					right := NewServiceError(rightCause, "same", rightTraits&1 != 0, rightTraits&2 != 0, rightTraits&4 != 0)
					field := "assigned after construction"
					left.Field = &field
					leftBefore, rightBefore := *left, *right

					merged := requireMergedError(t, MergeErrors(left, right))
					require.NotSame(t, left, merged)
					require.NotSame(t, right, merged)
					require.Equal(t, leftBefore, *left)
					require.Equal(t, rightBefore, *right)
					expectedName := name
					if name == "error" {
						expectedName = right.Name
					}
					require.Equal(t, expectedName, merged.Name)
					require.Equal(t, left.ID, merged.ID)
					require.Same(t, left.Field, merged.Field)
					require.Equal(t, "same message; same message", merged.Message)
					require.Equal(t, left.Timeout && right.Timeout, merged.Timeout)
					require.Equal(t, left.Temporary && right.Temporary, merged.Temporary)
					require.Equal(t, left.Fault && right.Fault, merged.Fault)
					joined, ok := merged.Unwrap().(interface{ Unwrap() []error })
					require.True(t, ok)
					children := joined.Unwrap()
					require.Len(t, children, 2)
					require.Same(t, leftCause, children[0])
					require.Same(t, rightCause, children[1])
					require.ErrorIs(t, merged, leftCause)
					require.ErrorIs(t, merged, rightCause)
					require.NotErrorIs(t, merged, left)
					require.NotErrorIs(t, merged, right)
					var selected *ServiceError
					require.ErrorAs(t, merged, &selected)
					require.Same(t, merged, selected)

					history := merged.History()
					require.Len(t, history, 2)
					requireOriginalContribution(t, leftBefore, history[0])
					requireOriginalContribution(t, rightBefore, history[1])
					require.NotSame(t, left.Field, history[0].Field)
					require.Nil(t, history[1].Field)
				})
			}
		}
	}
}

func TestMergeErrorsExistingSelection(t *testing.T) {
	for _, mode := range []string{"direct", "wrapped", "custom As"} {
		t.Run(mode, func(t *testing.T) {
			left := NewServiceError(context.Canceled, "selected", true, false, false)
			right := NewServiceError(context.DeadlineExceeded, "right", false, true, true)
			var input error = left
			switch mode {
			case "wrapped":
				input = fmt.Errorf("outer text is not selected: %w", left)
			case "custom As":
				input = &mergeSelectionError{selected: left}
			}
			require.Same(t, left, asError(input))
			merged := requireMergedError(t, MergeErrors(input, fmt.Errorf("other outer: %w", right)))
			require.Equal(t, "selected", merged.Name)
			require.Equal(t, left.ID, merged.ID)
			require.Equal(t, "context canceled; context deadline exceeded", merged.Message)
			requireOriginalContribution(t, *left, merged.History()[0])
			requireOriginalContribution(t, *right, merged.History()[1])
			require.Same(t, context.Canceled, left.Unwrap())
			require.Equal(t, "context canceled", left.Message)
		})
	}
	t.Run("raw inputs", func(t *testing.T) {
		left, right := errors.New("raw left"), errors.New("raw right")
		merged := requireMergedError(t, MergeErrors(left, right))
		require.Equal(t, "error", merged.Name)
		require.Equal(t, "raw left; raw right", merged.Message)
		require.True(t, merged.Fault)
		history := merged.History()
		require.Len(t, history, 2)
		require.Equal(t, merged.ID, history[0].ID)
		require.NotEmpty(t, history[0].ID)
		require.NotEmpty(t, history[1].ID)
		require.NotEqual(t, history[0].ID, history[1].ID)
		require.Same(t, left, history[0].Unwrap())
		require.Same(t, right, history[1].Unwrap())
		require.Equal(t, "raw left", history[0].Message)
		require.Equal(t, "raw right", history[1].Message)
	})
	for _, rawLeft := range []bool{false, true} {
		t.Run(fmt.Sprintf("raw left=%t", rawLeft), func(t *testing.T) {
			raw := errors.New("raw")
			service := NewServiceError(context.Canceled, "selected", false, false, false)
			var left, right error = service, raw
			if rawLeft {
				left, right = right, left
			}
			merged := requireMergedError(t, MergeErrors(left, right))
			require.Equal(t, service.Name, merged.Name)
			history := merged.History()
			require.Len(t, history, 2)
			rawIndex := 1
			if rawLeft {
				rawIndex = 0
			}
			require.Equal(t, "error", history[rawIndex].Name)
			require.Equal(t, "raw", history[rawIndex].Message)
			require.True(t, history[rawIndex].Fault)
			require.Same(t, raw, history[rawIndex].Unwrap())
			require.False(t, merged.Fault)
		})
	}
}

func TestMergeErrorsAbsentCauses(t *testing.T) {
	for _, test := range []struct {
		name       string
		leftCause  error
		rightCause error
	}{
		{"neither", nil, nil},
		{"left", context.Canceled, nil},
		{"right", nil, context.Canceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			left, right := Fault("left"), Fault("right")
			if test.leftCause != nil {
				left = NewServiceError(test.leftCause, "left", false, false, true)
			}
			if test.rightCause != nil {
				right = NewServiceError(test.rightCause, "right", false, false, true)
			}
			merged := requireMergedError(t, MergeErrors(left, right))
			if test.leftCause == nil && test.rightCause == nil {
				require.Nil(t, merged.Unwrap())
			} else {
				joined, ok := merged.Unwrap().(interface{ Unwrap() []error })
				require.True(t, ok)
				require.Len(t, joined.Unwrap(), 1)
				require.Same(t, context.Canceled, joined.Unwrap()[0])
			}
			requireOriginalContribution(t, *left, merged.History()[0])
			requireOriginalContribution(t, *right, merged.History()[1])
		})
	}
}

func TestMergeErrorsFlatteningAndAliases(t *testing.T) {
	for _, scenario := range []string{"repeat", "both merged", "self", "shared", "mutual"} {
		t.Run(scenario, func(t *testing.T) {
			a := NewServiceError(context.Canceled, "a", true, true, false)
			b := NewServiceError(context.DeadlineExceeded, "b", false, true, true)
			c := NewServiceError(errors.New("independent"), "c", false, false, false)
			ab := requireMergedError(t, MergeErrors(a, b))
			abBefore := *ab
			var result *ServiceError
			var originals []*ServiceError
			switch scenario {
			case "repeat":
				result = requireMergedError(t, MergeErrors(ab, c))
				originals = []*ServiceError{a, b, c}
			case "both merged":
				bc := requireMergedError(t, MergeErrors(b, c))
				result = requireMergedError(t, MergeErrors(ab, bc))
				originals = []*ServiceError{a, b, b, c}
			case "self":
				result = requireMergedError(t, MergeErrors(ab, ab))
				originals = []*ServiceError{a, b, a, b}
			case "shared":
				ac := requireMergedError(t, MergeErrors(a, c))
				result = requireMergedError(t, MergeErrors(ab, ac))
				originals = []*ServiceError{a, b, a, c}
			case "mutual":
				nextB := requireMergedError(t, MergeErrors(b, ab))
				result = requireMergedError(t, MergeErrors(ab, nextB))
				originals = []*ServiceError{a, b, b, a, b}
			}
			require.Equal(t, abBefore, *ab)
			history := result.History()
			require.Len(t, history, len(originals))
			for i, original := range originals {
				requireOriginalContribution(t, *original, history[i])
			}
			require.NotErrorIs(t, result, errors.New("missing sentinel"))
			var missing *mergeMissingCause
			require.False(t, errors.As(result, &missing))
			require.ErrorIs(t, result, context.Canceled)
			require.ErrorIs(t, result, context.DeadlineExceeded)
		})
	}
	t.Run("unmerged self", func(t *testing.T) {
		a := NewServiceError(context.Canceled, "a", false, false, false)
		before := *a
		merged := requireMergedError(t, MergeErrors(a, a))
		require.Equal(t, before, *a)
		require.Equal(t, "context canceled; context canceled", merged.Message)
		require.Len(t, merged.History(), 2)
		for _, entry := range merged.History() {
			requireOriginalContribution(t, before, entry)
		}
	})
}

func TestServiceErrorHistoryDetachedValues(t *testing.T) {
	for _, mergedInput := range []bool{false, true} {
		t.Run(fmt.Sprintf("merged=%t", mergedInput), func(t *testing.T) {
			field := "original field"
			a := NewServiceError(context.Canceled, "a", true, true, false)
			a.Field = &field
			b := NewServiceError(context.DeadlineExceeded, "b", false, false, true)
			input := a
			if mergedInput {
				input = requireMergedError(t, MergeErrors(a, b))
			}
			before := *input
			for i := 0; i < 3; i++ {
				history := input.History()
				fresh := input.History()
				require.NotSame(t, input, history[0])
				require.NotSame(t, history[0], fresh[0])
				require.NotSame(t, history[0].Field, fresh[0].Field)
				requireOriginalContribution(t, *a, history[0])
				*history[0].Field = "caller Field write"
				history[0].Name = "caller name"
				history[0].ID = "caller ID"
				history[0].Message = "caller message"
				history[0].Timeout = false
				history[0].Temporary = false
				history[0].Fault = true
				history[0].err = errors.New("caller cause")
				if mergedInput {
					history[1].Field = &field
					history[1].Message = "caller right write"
				}
				history[0] = b
				history = append(history, b)
				require.Len(t, history, len(fresh)+1)
				require.Equal(t, before, *input)
				requireOriginalContribution(t, *a, input.History()[0])
				require.Equal(t, "original field", field)
				if mergedInput {
					requireOriginalContribution(t, *b, input.History()[1])
				}
			}
		})
	}
	t.Run("inclusion captures original values", func(t *testing.T) {
		field := "before inclusion"
		a := NewServiceError(context.Canceled, "error", false, false, false)
		a.Field = &field
		b := NewServiceError(context.DeadlineExceeded, "right", false, false, false)
		merged := requireMergedError(t, MergeErrors(a, b))
		a.Name, a.Message = "later name", "later message"
		field = "later Field write"
		require.Same(t, a.Field, merged.Field)
		require.Equal(t, field, *merged.Field)
		require.Equal(t, "error", merged.History()[0].Name)
		require.Equal(t, "context canceled", merged.History()[0].Message)
		require.Equal(t, "before inclusion", *merged.History()[0].Field)
		require.Equal(t, "later name", a.History()[0].Name)
		require.Equal(t, field, *a.History()[0].Field)
		wholeField := "whole-only field"
		merged.Field = &wholeField
		next := requireMergedError(t, MergeErrors(merged, b))
		require.Same(t, merged.Field, next.Field)
		require.Equal(t, "before inclusion", *next.History()[0].Field)
		require.Equal(t, "right", merged.Name)
	})
	t.Run("absent original Field stays absent", func(t *testing.T) {
		a, b := Fault("a"), Fault("b")
		merged := requireMergedError(t, MergeErrors(a, b))
		field := "whole-only field"
		merged.Field = &field
		require.Nil(t, merged.History()[0].Field)
		require.Nil(t, merged.History()[1].Field)
	})
}

func TestMergeErrorsFiniteReferencedCauses(t *testing.T) {
	for _, opaque := range []bool{false, true} {
		t.Run(fmt.Sprintf("opaque=%t", opaque), func(t *testing.T) {
			a := NewServiceError(context.Canceled, "first", false, false, false)
			a.Message = "original"
			a.ID = "original-id"
			field := "original-field"
			a.Field = &field
			match := errors.New("custom match")
			view := &mergeCauseView{message: "original view"}
			wrapper := &mergeOpaqueCause{input: a, match: match, view: view}
			var cause error = a
			if opaque {
				cause = wrapper
			}
			b := NewServiceError(cause, "second", false, false, false)
			beforeA, beforeB := *a, *b
			beforeText := cause.Error()
			merged := requireMergedError(t, MergeErrors(a, b))
			require.Equal(t, beforeA, *a)
			require.Equal(t, beforeB, *b)
			require.Equal(t, beforeText, cause.Error())
			require.Same(t, cause, merged.History()[1].Unwrap())
			require.ErrorIs(t, merged, a, "the original input is genuinely a right-side cause")
			require.ErrorIs(t, merged, context.Canceled)
			require.NotErrorIs(t, merged, errors.New("absent"))
			var missing *mergeMissingCause
			require.False(t, errors.As(merged, &missing))
			if opaque {
				require.ErrorIs(t, merged, match)
				var gotWrapper *mergeOpaqueCause
				require.ErrorAs(t, merged, &gotWrapper)
				require.Same(t, wrapper, gotWrapper)
				var gotView *mergeCauseView
				require.ErrorAs(t, merged, &gotView)
				require.Same(t, view, gotView)
			}
		})
	}
}

func TestMergeErrorsDoesNotRepairExistingCycle(t *testing.T) {
	a := NewServiceError(context.Canceled, "first", false, false, false)
	b := NewServiceError(a, "second", false, false, false)
	// The old merge wrote this cause onto a. Inspect only the immediate edges;
	// traversing the missing target on this already cyclic input could hang.
	oldCause := errors.Join(a.err, b.err)
	a.err = oldCause
	next := requireMergedError(t, MergeErrors(a, Fault("independent")))
	joined, ok := next.Unwrap().(interface{ Unwrap() []error })
	require.True(t, ok)
	require.Same(t, oldCause, joined.Unwrap()[0])
	oldJoined, ok := oldCause.(interface{ Unwrap() []error })
	require.True(t, ok)
	require.Same(t, a, oldJoined.Unwrap()[1])
	require.Same(t, oldCause, a.Unwrap())
}

func TestServiceErrorHistoryConcurrentReaders(t *testing.T) {
	a := NewServiceError(context.Canceled, "a", false, false, false)
	field := "published field"
	a.Field = &field
	b := NewServiceError(context.DeadlineExceeded, "b", false, false, false)
	merged := requireMergedError(t, MergeErrors(a, b))
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go exerciseHistoryCopies(t, merged, &workers)
	}
	workers.Wait()
	requireOriginalContribution(t, *a, merged.History()[0])
	requireOriginalContribution(t, *b, merged.History()[1])
}

// requireMergedError checks the runtime result type so subsequent assertions
// inspect the returned whole error rather than a matching input cause.
func requireMergedError(t *testing.T, err error) *ServiceError {
	t.Helper()
	value, ok := err.(*ServiceError) //nolint:errorlint // Check the returned object itself.
	require.True(t, ok)
	return value
}

// requireOriginalContribution compares every saved field and exact cause to
// the original input and verifies that the entry has no nested merge history.
func requireOriginalContribution(t *testing.T, want ServiceError, got *ServiceError) {
	t.Helper()
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.Message, got.Message)
	require.Equal(t, want.Timeout, got.Timeout)
	require.Equal(t, want.Temporary, got.Temporary)
	require.Equal(t, want.Fault, got.Fault)
	require.Equal(t, want.Field, got.Field)
	if want.err == nil {
		require.Nil(t, got.Unwrap())
	} else {
		require.True(t, want.err == got.Unwrap()) //nolint:errorlint // Compare the exact retained cause, including value errors.
	}
	require.Empty(t, got.history)
}

// exerciseHistoryCopies reads a fully constructed error while editing only
// detached copies, so the race detector can check isolation between readers.
func exerciseHistoryCopies(t *testing.T, merged *ServiceError, workers *sync.WaitGroup) {
	defer workers.Done()
	for i := 0; i < 16; i++ {
		history := merged.History()
		if history[0].Name != "a" || *history[0].Field != "published field" {
			t.Error("a returned History write reached saved values")
		}
		history[0].Name = "reader edit"
		*history[0].Field = "reader edit"
		history[1] = history[0]
		next := MergeErrors(merged, merged)
		if !errors.Is(next, context.Canceled) || !errors.Is(next, context.DeadlineExceeded) {
			t.Error("concurrent merge lost an original cause")
		}
		var missing *mergeMissingCause
		if errors.As(next, &missing) {
			t.Error("concurrent merge acquired an absent cause")
		}
	}
}

func (e *mergeSelectionError) Error() string {
	return "outer text is not selected"
}

func (e *mergeSelectionError) As(target any) bool {
	if service, ok := target.(**ServiceError); ok {
		*service = e.selected
		return true
	}
	return false
}

func (e *mergeOpaqueCause) Error() string {
	return "wrapped: " + e.input.Error()
}

func (e *mergeOpaqueCause) Unwrap() error {
	return e.input
}

func (e *mergeOpaqueCause) Is(target error) bool {
	return target == e.match && e.input.Message == "original" && //nolint:errorlint // This shallow method owns exact sentinel matching.
		e.input.Name == "first" && e.input.ID == "original-id" &&
		*e.input.Field == "original-field" && !e.input.Fault
}

func (e *mergeOpaqueCause) As(target any) bool {
	if view, ok := target.(**mergeCauseView); ok && e.input.Message == "original" {
		*view = e.view
		return true
	}
	return false
}

func (e *mergeCauseView) Error() string {
	return e.message
}

func (e *mergeMissingCause) Error() string {
	return "missing cause"
}
