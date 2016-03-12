package goa_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPError", func() {
	const (
		id     = "goa.1"
		title  = "title"
		status = 400
		err    = "error"
	)

	var httpError *goa.HTTPError

	BeforeEach(func() {
		httpError = &goa.HTTPError{&goa.ErrorClass{id, title, status}, err}
	})

	It("serializes to JSON", func() {
		b, err := json.Marshal(httpError)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(string(b)).Should(Equal(`{"id":"goa.1","title":"title","err":"error"}`))
	})
})

var _ = Describe("InvalidParamTypeError", func() {
	var valErr error
	name := "param"
	val := 42
	expected := "43"

	JustBeforeEach(func() {
		valErr = goa.InvalidParamTypeError(name, val, expected)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(name))
		Ω(err.Err).Should(ContainSubstring("%d", val))
		Ω(err.Err).Should(ContainSubstring(expected))
	})
})

var _ = Describe("MissingParaerror", func() {
	var valErr error
	name := "param"

	JustBeforeEach(func() {
		valErr = goa.MissingParamError(name)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(name))
	})
})

var _ = Describe("InvalidAttributeTypeError", func() {
	var valErr error
	ctx := "ctx"
	val := 42
	expected := "43"

	JustBeforeEach(func() {
		valErr = goa.InvalidAttributeTypeError(ctx, val, expected)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(ctx))
		Ω(err.Err).Should(ContainSubstring("%d", val))
		Ω(err.Err).Should(ContainSubstring(expected))
	})
})

var _ = Describe("MissingAttributeError", func() {
	var valErr error
	ctx := "ctx"
	name := "param"

	JustBeforeEach(func() {
		valErr = goa.MissingAttributeError(ctx, name)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(ctx))
		Ω(err.Err).Should(ContainSubstring(name))
	})
})

var _ = Describe("MissingHeaderError", func() {
	var valErr error
	name := "param"

	JustBeforeEach(func() {
		valErr = goa.MissingHeaderError(name)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(name))
	})
})

var _ = Describe("InvalidEnumValueError", func() {
	var valErr error
	ctx := "ctx"
	val := 42
	allowed := []interface{}{"43", "44"}

	JustBeforeEach(func() {
		valErr = goa.InvalidEnumValueError(ctx, val, allowed)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(ctx))
		Ω(err.Err).Should(ContainSubstring("%d", val))
		Ω(err.Err).Should(ContainSubstring(`"43", "44"`))
	})
})

var _ = Describe("InvalidFormaerror", func() {
	var valErr error
	ctx := "ctx"
	target := "target"
	format := goa.FormatDateTime
	formatError := errors.New("boo")

	JustBeforeEach(func() {
		valErr = goa.InvalidFormatError(ctx, target, format, formatError)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(ctx))
		Ω(err.Err).Should(ContainSubstring(target))
		Ω(err.Err).Should(ContainSubstring("date-time"))
		Ω(err.Err).Should(ContainSubstring(formatError.Error()))
	})
})

var _ = Describe("InvalidPatternError", func() {
	var valErr error
	ctx := "ctx"
	target := "target"
	pattern := "pattern"

	JustBeforeEach(func() {
		valErr = goa.InvalidPatternError(ctx, target, pattern)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(ctx))
		Ω(err.Err).Should(ContainSubstring(target))
		Ω(err.Err).Should(ContainSubstring(pattern))
	})
})

var _ = Describe("InvalidRangeError", func() {
	var valErr error
	ctx := "ctx"
	target := "target"
	value := 42
	min := true

	JustBeforeEach(func() {
		valErr = goa.InvalidRangeError(ctx, target, value, min)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
		err := valErr.(*goa.HTTPError)
		Ω(err.Err).Should(ContainSubstring(ctx))
		Ω(err.Err).Should(ContainSubstring("greater or equal"))
		Ω(err.Err).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
		Ω(err.Err).Should(ContainSubstring(target))
	})
})

var _ = Describe("InvalidLengthError", func() {
	const ctx = "ctx"
	const value = 42
	const min = true

	var target interface{}
	var ln int

	var valErr error

	JustBeforeEach(func() {
		valErr = goa.InvalidLengthError(ctx, target, ln, value, min)
	})

	Context("on strings", func() {
		BeforeEach(func() {
			target = "target"
			ln = len("target")
		})

		It("creates a http error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
			err := valErr.(*goa.HTTPError)
			Ω(err.Err).Should(ContainSubstring(ctx))
			Ω(err.Err).Should(ContainSubstring("greater or equal"))
			Ω(err.Err).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(err.Err).Should(ContainSubstring(target.(string)))
		})
	})

	Context("on slices", func() {
		BeforeEach(func() {
			target = []string{"target1", "target2"}
			ln = 2
		})

		It("creates a http error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(&goa.HTTPError{}))
			err := valErr.(*goa.HTTPError)
			Ω(err.Err).Should(ContainSubstring(ctx))
			Ω(err.Err).Should(ContainSubstring("greater or equal"))
			Ω(err.Err).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(err.Err).Should(ContainSubstring(fmt.Sprintf("%#v", target)))
		})
	})
})

var _ = Describe("MultiError", func() {
	var err1, err2 error
	var multiError goa.MultiError

	JustBeforeEach(func() {
		multiError = goa.MultiError{err1, err2}
	})

	BeforeEach(func() {
		err1 = errors.New("error 1")
		err2 = errors.New("error 2")
	})

	It("lists all error messages in its error message", func() {
		errMsg := multiError.Error()
		Ω(errMsg).ShouldNot(BeEmpty())
		Ω(errMsg).Should(ContainSubstring(err1.Error()))
		Ω(errMsg).Should(ContainSubstring(err2.Error()))
	})
})

var _ = Describe("BuildError", func() {
	var err, err2 error
	var mErr error

	BeforeEach(func() {
		err = nil
		err2 = nil
	})

	JustBeforeEach(func() {
		mErr = goa.BuildError(err, err2)
	})

	Context("with two nil errors", func() {
		It("returns an empty error", func() {
			Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			Ω(mErr).Should(BeEmpty())
		})
	})

	Context("with the second error nil", func() {
		Context("with the first argument a MultiError", func() {
			BeforeEach(func() {
				err = goa.MultiError{errors.New("foo")}
			})

			It("returns it", func() {
				Ω(mErr).Should(Equal(err))
			})
		})

		Context("with the first argument not a MultiError", func() {
			BeforeEach(func() {
				err = errors.New("foo")
			})

			It("wraps it into a MultiError", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
				mmErr := mErr.(goa.MultiError)
				Ω(mmErr).Should(HaveLen(1))
				Ω(mmErr[0]).Should(Equal(err))
			})
		})

	})

	Context("with the first error nil", func() {
		Context("with the second argument a MultiError", func() {
			BeforeEach(func() {
				err2 = goa.MultiError{errors.New("foo")}
			})

			It("returns it", func() {
				Ω(mErr).Should(Equal(err2))
			})
		})

		Context("with the second argument not a MultiError", func() {
			BeforeEach(func() {
				err2 = errors.New("foo")
			})

			It("wraps it into a MultiError", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
				mmErr := mErr.(goa.MultiError)
				Ω(mmErr).Should(HaveLen(1))
				Ω(mmErr[0]).Should(Equal(err2))
			})
		})

	})

	Context("with the first error a MultiError", func() {
		var merr goa.MultiError

		BeforeEach(func() {
			merr = goa.MultiError{errors.New("foo")}
			err = merr
		})

		Context("with the second error a MultiError", func() {
			var merr2 goa.MultiError

			BeforeEach(func() {
				merr2 = goa.MultiError{errors.New("foo2")}
				err2 = merr2
			})

			It("concatenates both errors", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
				mmErr := mErr.(goa.MultiError)
				Ω(mmErr).Should(HaveLen(2))
				Ω(mmErr[0]).Should(Equal(merr[0]))
				Ω(mmErr[1]).Should(Equal(merr2[0]))
			})
		})

		Context("with the second not a MultiError", func() {
			BeforeEach(func() {
				err2 = errors.New("foo2")
			})

			It("concatenates both errors", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
				mmErr := mErr.(goa.MultiError)
				Ω(mmErr).Should(HaveLen(2))
				Ω(mmErr[0]).Should(Equal(merr[0]))
				Ω(mmErr[1]).Should(Equal(err2))
			})
		})
	})

	Context("with the first error not a MultiError", func() {
		BeforeEach(func() {
			err = errors.New("foo")
		})

		Context("with the second error a MultiError", func() {
			var merr2 goa.MultiError

			BeforeEach(func() {
				merr2 = goa.MultiError{errors.New("foo2")}
				err2 = merr2
			})

			It("concatenates both errors", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
				mmErr := mErr.(goa.MultiError)
				Ω(mmErr).Should(HaveLen(2))
				Ω(mmErr[0]).Should(Equal(err))
				Ω(mmErr[1]).Should(Equal(merr2[0]))
			})
		})

		Context("with the second not a MultiError", func() {
			BeforeEach(func() {
				err2 = errors.New("foo2")
			})

			It("concatenates both errors", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
				mmErr := mErr.(goa.MultiError)
				Ω(mmErr).Should(HaveLen(2))
				Ω(mmErr[0]).Should(Equal(err))
				Ω(mmErr[1]).Should(Equal(err2))
			})
		})
	})
})
