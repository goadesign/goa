package goa_test

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
)

// allErrorKinds list all the existing goa.ErrorKind values.
var allErrorKinds = [10]goa.ErrorKind{
	goa.ErrInvalidParamType,
	goa.ErrMissingParam,
	goa.ErrInvalidAttributeType,
	goa.ErrMissingAttribute,
	goa.ErrInvalidEnumValue,
	goa.ErrMissingHeader,
	goa.ErrInvalidFormat,
	goa.ErrInvalidPattern,
	goa.ErrInvalidRange,
	goa.ErrInvalidLength,
}

var _ = Describe("ErrorKind", func() {
	for _, kind := range allErrorKinds {
		It(fmt.Sprintf("Kind %#v has a title", kind), func() {
			Ω(kind.Title()).ShouldNot(BeEmpty())
		})
	}
})

var _ = Describe("TypedError", func() {
	var kind goa.ErrorKind
	var msg string

	var typedError *goa.TypedError

	JustBeforeEach(func() {
		typedError = &goa.TypedError{Kind: kind, Mesg: msg}
	})

	Context("of kind ErrInvalidParamType", func() {
		BeforeEach(func() {
			kind = goa.ErrInvalidParamType
			msg = "some error message"
		})

		It("builds an error message that is valid JSON and contains the kind title and user message", func() {
			errMsg := typedError.Error()
			Ω(errMsg).ShouldNot(BeEmpty())
			var data map[string]interface{}
			err := json.Unmarshal([]byte(errMsg), &data)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveKey("kind"))
			Ω(data["kind"]).Should(Equal(float64(kind)))
			Ω(data).Should(HaveKey("title"))
			Ω(data["title"]).Should(Equal(kind.Title()))
			Ω(data).Should(HaveKey("msg"))
			Ω(data["msg"]).Should(Equal(msg))
		})
	})

	Context("in a multi-error", func() {
		var multiError goa.MultiError

		BeforeEach(func() {
			kind = goa.ErrInvalidParamType
			msg = "some error message"
		})

		JustBeforeEach(func() {
			multiError = append(multiError, typedError)
		})

		It("Error creates a slice of one element", func() {
			errMsg := multiError.Error()
			Ω(errMsg).ShouldNot(BeEmpty())
			var data []map[string]interface{}
			err := json.Unmarshal([]byte(errMsg), &data)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			Ω(data[0]).Should(HaveKey("kind"))
			Ω(data[0]["kind"]).Should(Equal(float64(kind)))
			Ω(data[0]).Should(HaveKey("title"))
			Ω(data[0]["title"]).Should(Equal(kind.Title()))
			Ω(data[0]).Should(HaveKey("msg"))
			Ω(data[0]["msg"]).Should(Equal(msg))
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
		var data []map[string]interface{}
		err := json.Unmarshal([]byte(errMsg), &data)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(data).Should(HaveLen(2))
		Ω(data[0]).Should(HaveKey("kind"))
		Ω(data[0]["kind"]).Should(Equal(float64(0)))
		Ω(data[0]).Should(HaveKey("title"))
		Ω(data[0]["title"]).Should(Equal("generic"))
		Ω(data[0]).Should(HaveKey("msg"))
		Ω(data[0]["msg"]).Should(Equal(err1.Error()))
		Ω(data[0]).Should(HaveKey("kind"))
		Ω(data[1]["kind"]).Should(Equal(float64(0)))
		Ω(data[1]).Should(HaveKey("title"))
		Ω(data[1]["title"]).Should(Equal("generic"))
		Ω(data[1]).Should(HaveKey("msg"))
		Ω(data[1]["msg"]).Should(Equal(err2.Error()))
	})
})

var _ = Describe("BadRequestError", func() {
	var rawErr error
	var badReq *goa.BadRequestError

	BeforeEach(func() {
		rawErr = errors.New("raw error")
	})

	JustBeforeEach(func() {
		badReq = goa.NewBadRequestError(rawErr)
	})

	It("builds a bad request error that reports the inner error message", func() {
		Ω(badReq).ShouldNot(BeNil())
		Ω(badReq.Error()).Should(Equal(rawErr.Error()))
	})
})

var _ = Describe("InvalidParamTypeError", func() {
	var valErr, err error
	name := "param"
	val := 42
	expected := "43"

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidParamTypeError(name, val, expected, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidParamType))))
		Ω(tErr.Mesg).Should(ContainSubstring(name))
		Ω(tErr.Mesg).Should(ContainSubstring("%d", val))
		Ω(tErr.Mesg).Should(ContainSubstring(expected))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidParamType))))
			Ω(tErr.Mesg).Should(ContainSubstring(name))
			Ω(tErr.Mesg).Should(ContainSubstring("%d", val))
			Ω(tErr.Mesg).Should(ContainSubstring(expected))
		})
	})
})

var _ = Describe("MissingParamError", func() {
	var valErr, err error
	name := "param"

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.MissingParamError(name, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrMissingParam))))
		Ω(tErr.Mesg).Should(ContainSubstring(name))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrMissingParam))))
			Ω(tErr.Mesg).Should(ContainSubstring(name))
		})
	})
})

var _ = Describe("InvalidAttributeTypeError", func() {
	var valErr, err error
	ctx := "ctx"
	val := 42
	expected := "43"

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidAttributeTypeError(ctx, val, expected, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidAttributeType))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring("%d", val))
		Ω(tErr.Mesg).Should(ContainSubstring(expected))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidAttributeType))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring("%d", val))
			Ω(tErr.Mesg).Should(ContainSubstring(expected))
		})
	})
})

var _ = Describe("MissingAttributeError", func() {
	var valErr, err error
	ctx := "ctx"
	name := "param"

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.MissingAttributeError(ctx, name, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrMissingAttribute))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring(name))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrMissingAttribute))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring(name))
		})
	})
})

var _ = Describe("MissingHeaderError", func() {
	var valErr, err error
	name := "param"

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.MissingHeaderError(name, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrMissingHeader))))
		Ω(tErr.Mesg).Should(ContainSubstring(name))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrMissingHeader))))
			Ω(tErr.Mesg).Should(ContainSubstring(name))
		})
	})
})

var _ = Describe("InvalidEnumValueError", func() {
	var valErr, err error
	ctx := "ctx"
	val := 42
	allowed := []interface{}{"43", "44"}

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidEnumValueError(ctx, val, allowed, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidEnumValue))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring("%d", val))
		Ω(tErr.Mesg).Should(ContainSubstring(`"43", "44"`))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidEnumValue))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring("%d", val))
			Ω(tErr.Mesg).Should(ContainSubstring(`"43", "44"`))
		})
	})
})

var _ = Describe("InvalidFormatError", func() {
	var valErr, err error
	ctx := "ctx"
	target := "target"
	format := goa.FormatDateTime
	formatError := errors.New("boo")

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidFormatError(ctx, target, format, formatError, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidFormat))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring(target))
		Ω(tErr.Mesg).Should(ContainSubstring("date-time"))
		Ω(tErr.Mesg).Should(ContainSubstring(formatError.Error()))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidFormat))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring(target))
			Ω(tErr.Mesg).Should(ContainSubstring("date-time"))
			Ω(tErr.Mesg).Should(ContainSubstring(formatError.Error()))
		})
	})
})

var _ = Describe("InvalidPatternError", func() {
	var valErr, err error
	ctx := "ctx"
	target := "target"
	pattern := "pattern"

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidPatternError(ctx, target, pattern, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidPattern))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring(target))
		Ω(tErr.Mesg).Should(ContainSubstring(pattern))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidPattern))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring(target))
			Ω(tErr.Mesg).Should(ContainSubstring(pattern))
		})
	})
})

var _ = Describe("InvalidRangeError", func() {
	var valErr, err error
	ctx := "ctx"
	target := "target"
	value := 42
	min := true

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidRangeError(ctx, target, value, min, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidRange))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring("greater or equal"))
		Ω(tErr.Mesg).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
		Ω(tErr.Mesg).Should(ContainSubstring(target))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidRange))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring("greater or equal"))
			Ω(tErr.Mesg).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(tErr.Mesg).Should(ContainSubstring(target))
		})
	})
})

var _ = Describe("InvalidLengthError", func() {
	var valErr, err error
	ctx := "ctx"
	target := "target"
	value := 42
	min := true

	BeforeEach(func() {
		err = nil
	})

	JustBeforeEach(func() {
		valErr = goa.InvalidLengthError(ctx, target, value, min, err)
	})

	It("creates a multi error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
		mErr := valErr.(goa.MultiError)
		Ω(mErr).Should(HaveLen(1))
		Ω(mErr[0]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
		tErr := mErr[0].(*goa.TypedError)
		Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidLength))))
		Ω(tErr.Mesg).Should(ContainSubstring(ctx))
		Ω(tErr.Mesg).Should(ContainSubstring("greater or equal"))
		Ω(tErr.Mesg).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
		Ω(tErr.Mesg).Should(ContainSubstring(target))
	})

	Context("with a pre-existing error", func() {
		BeforeEach(func() {
			err = errors.New("pre-existing")
		})

		It("appends to the multi-error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(goa.MultiError{}))
			mErr := valErr.(goa.MultiError)
			Ω(mErr).Should(HaveLen(2))
			Ω(mErr[0]).Should(Equal(err))
			Ω(mErr[1]).Should(BeAssignableToTypeOf(&goa.TypedError{}))
			tErr := mErr[1].(*goa.TypedError)
			Ω(tErr.Kind).Should(Equal(goa.ErrorKind((goa.ErrInvalidLength))))
			Ω(tErr.Mesg).Should(ContainSubstring(ctx))
			Ω(tErr.Mesg).Should(ContainSubstring("greater or equal"))
			Ω(tErr.Mesg).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(tErr.Mesg).Should(ContainSubstring(target))
		})
	})
})

var _ = Describe("ReportError", func() {
	var err, err2 error
	var mErr error

	BeforeEach(func() {
		err = nil
		err2 = nil
	})

	JustBeforeEach(func() {
		mErr = goa.ReportError(err, err2)
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
