package goa

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error", func() {
	const (
		id     = "foo"
		code   = "invalid"
		status = 400
		detail = "error"
	)
	var meta = []map[string]interface{}{{"what": 42}}

	var gerr *ErrorResponse

	BeforeEach(func() {
		gerr = &ErrorResponse{ID: id, Code: code, Status: status, Detail: detail, Meta: meta}
	})

	It("serializes to JSON", func() {
		b, err := json.Marshal(gerr)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(string(b)).Should(Equal(`{"id":"foo","code":"invalid","status":400,"detail":"error","meta":[{"what":42}]}`))
	})
})

var _ = Describe("InvalidParamTypeError", func() {
	var valErr error
	name := "param"
	val := 42
	expected := "43"

	JustBeforeEach(func() {
		valErr = InvalidParamTypeError(name, val, expected)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(name))
		Ω(err.Detail).Should(ContainSubstring("%d", val))
		Ω(err.Detail).Should(ContainSubstring(expected))
	})
})

var _ = Describe("MissingParaerror", func() {
	var valErr error
	name := "param"

	JustBeforeEach(func() {
		valErr = MissingParamError(name)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(name))
	})
})

var _ = Describe("InvalidAttributeTypeError", func() {
	var valErr error
	ctx := "ctx"
	val := 42
	expected := "43"

	JustBeforeEach(func() {
		valErr = InvalidAttributeTypeError(ctx, val, expected)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(ctx))
		Ω(err.Detail).Should(ContainSubstring("%d", val))
		Ω(err.Detail).Should(ContainSubstring(expected))
	})
})

var _ = Describe("MissingAttributeError", func() {
	var valErr error
	ctx := "ctx"
	name := "param"

	JustBeforeEach(func() {
		valErr = MissingAttributeError(ctx, name)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(ctx))
		Ω(err.Detail).Should(ContainSubstring(name))
	})
})

var _ = Describe("MissingHeaderError", func() {
	var valErr error
	name := "param"

	JustBeforeEach(func() {
		valErr = MissingHeaderError(name)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(name))
	})
})

var _ = Describe("InvalidEnumValueError", func() {
	var valErr error
	ctx := "ctx"
	val := 42
	allowed := []interface{}{"43", "44"}

	JustBeforeEach(func() {
		valErr = InvalidEnumValueError(ctx, val, allowed)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(ctx))
		Ω(err.Detail).Should(ContainSubstring("%d", val))
		Ω(err.Detail).Should(ContainSubstring(`"43", "44"`))
	})
})

var _ = Describe("InvalidFormaerror", func() {
	var valErr error
	ctx := "ctx"
	target := "target"
	format := FormatDateTime
	formatError := errors.New("boo")

	JustBeforeEach(func() {
		valErr = InvalidFormatError(ctx, target, format, formatError)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(ctx))
		Ω(err.Detail).Should(ContainSubstring(target))
		Ω(err.Detail).Should(ContainSubstring("date-time"))
		Ω(err.Detail).Should(ContainSubstring(formatError.Error()))
	})
})

var _ = Describe("InvalidPatternError", func() {
	var valErr error
	ctx := "ctx"
	target := "target"
	pattern := "pattern"

	JustBeforeEach(func() {
		valErr = InvalidPatternError(ctx, target, pattern)
	})

	It("creates a http error", func() {
		Ω(valErr).ShouldNot(BeNil())
		Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
		err := valErr.(*ErrorResponse)
		Ω(err.Detail).Should(ContainSubstring(ctx))
		Ω(err.Detail).Should(ContainSubstring(target))
		Ω(err.Detail).Should(ContainSubstring(pattern))
	})
})

var _ = Describe("InvalidRangeError", func() {
	var valErr error
	var value interface{}

	ctx := "ctx"
	target := "target"
	min := true

	JustBeforeEach(func() {
		valErr = InvalidRangeError(ctx, target, value, min)
	})

	Context("with an int value", func() {
		BeforeEach(func() {
			value = 42
		})

		It("creates a http error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
			err := valErr.(*ErrorResponse)
			Ω(err.Detail).Should(ContainSubstring(ctx))
			Ω(err.Detail).Should(ContainSubstring("greater or equal"))
			Ω(err.Detail).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(err.Detail).Should(ContainSubstring(target))
		})
	})

	Context("with a float64 value", func() {
		BeforeEach(func() {
			value = 42.42
		})

		It("creates a http error with no value truncation", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
			err := valErr.(*ErrorResponse)
			Ω(err.Detail).Should(ContainSubstring(ctx))
			Ω(err.Detail).Should(ContainSubstring("greater or equal"))
			Ω(err.Detail).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(err.Detail).Should(ContainSubstring(target))
		})
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
		valErr = InvalidLengthError(ctx, target, ln, value, min)
	})

	Context("on strings", func() {
		BeforeEach(func() {
			target = "target"
			ln = len("target")
		})

		It("creates a http error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
			err := valErr.(*ErrorResponse)
			Ω(err.Detail).Should(ContainSubstring(ctx))
			Ω(err.Detail).Should(ContainSubstring("greater or equal"))
			Ω(err.Detail).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(err.Detail).Should(ContainSubstring(target.(string)))
		})
	})

	Context("on slices", func() {
		BeforeEach(func() {
			target = []string{"target1", "target2"}
			ln = 2
		})

		It("creates a http error", func() {
			Ω(valErr).ShouldNot(BeNil())
			Ω(valErr).Should(BeAssignableToTypeOf(&ErrorResponse{}))
			err := valErr.(*ErrorResponse)
			Ω(err.Detail).Should(ContainSubstring(ctx))
			Ω(err.Detail).Should(ContainSubstring("greater or equal"))
			Ω(err.Detail).Should(ContainSubstring(fmt.Sprintf("%#v", value)))
			Ω(err.Detail).Should(ContainSubstring(fmt.Sprintf("%#v", target)))
		})
	})
})

var _ = Describe("Merge", func() {
	var err, err2 error
	var mErr *ErrorResponse

	BeforeEach(func() {
		err = nil
		err2 = nil
		mErr = nil
	})

	JustBeforeEach(func() {
		e := MergeErrors(err, err2)
		if e != nil {
			mErr = e.(*ErrorResponse)
		}
	})

	Context("with two nil errors", func() {
		It("returns a nil error", func() {
			Ω(mErr).Should(BeNil())
		})
	})

	Context("with a nil argument", func() {
		const code = "foo"

		BeforeEach(func() {
			err = &ErrorResponse{Code: code}
		})

		It("returns the target", func() {
			Ω(mErr).Should(Equal(err))
		})
	})

	Context("with a nil target", func() {
		Context("with the second argument a Error", func() {
			const detail = "foo"

			BeforeEach(func() {
				err2 = &ErrorResponse{Detail: detail}
			})

			It("returns it", func() {
				Ω(mErr).Should(Equal(err2))
			})
		})

		Context("with the second argument not a Error", func() {
			const detail = "foo"
			BeforeEach(func() {
				err2 = errors.New(detail)
			})

			It("wraps it into a Error", func() {
				Ω(mErr).ShouldNot(BeNil())
				Ω(mErr.Detail).Should(Equal(detail))
			})
		})

	})

	Context("with a non-nil target", func() {
		const detail = "foo"
		var status = 42
		var code = "common"
		var metaValues []map[string]interface{}

		BeforeEach(func() {
			err = &ErrorResponse{Detail: detail, Status: status, Code: code, Meta: metaValues}
		})

		Context("with another Error", func() {
			const detail2 = "foo2"
			var status2 = status
			var code2 = code
			var metaValues2 []map[string]interface{}
			var mErr2 *ErrorResponse

			BeforeEach(func() {
				mErr2 = &ErrorResponse{Detail: detail2, Status: status2, Code: code2, Meta: metaValues2}
				err2 = mErr2
			})

			It("concatenates both error details", func() {
				Ω(mErr.Detail).Should(Equal(detail + "; " + mErr2.Detail))
			})

			It("uses the common status", func() {
				Ω(mErr.Status).Should(Equal(status))
			})

			It("uses the common code", func() {
				Ω(mErr.Code).Should(Equal(code))
			})

			Context("with different code", func() {
				BeforeEach(func() {
					mErr2.Code = code + code
				})

				It("produces a bad_request error", func() {
					Ω(mErr.Code).Should(Equal("bad_request"))
					Ω(mErr.Status).Should(Equal(400))
					Ω(mErr.Detail).Should(Equal(detail + "; " + mErr2.Detail))
				})
			})

			Context("with different status", func() {
				BeforeEach(func() {
					mErr2.Status = status + status
				})

				It("produces a bad_request error", func() {
					Ω(mErr.Code).Should(Equal("bad_request"))
					Ω(mErr.Status).Should(Equal(400))
					Ω(mErr.Detail).Should(Equal(detail + "; " + mErr2.Detail))
				})
			})

			Context("with nil target metadata", func() {
				BeforeEach(func() {
					err.(*ErrorResponse).Meta = nil
				})

				Context("with nil/empty other metadata", func() {
					BeforeEach(func() {
						mErr2.Meta = nil
					})

					It("keeps nil target metadata if no other metadata", func() {
						Ω(mErr.Meta).Should(BeNil())
					})
				})

				Context("with other metadata", func() {
					var metaValues2 = []map[string]interface{}{{"foo": 1}, {"bar": 2}}

					BeforeEach(func() {
						err.(*ErrorResponse).Meta = nil
						mErr2.Meta = metaValues2
					})

					It("merges the metadata", func() {
						Ω(mErr.Meta).Should(HaveLen(len(metaValues2)))
						for i, val := range metaValues2 {
							for k, v := range val {
								Ω(mErr.Meta[i]).Should(HaveKeyWithValue(k, v))
							}
						}
					})
				})
			})

		})
	})

})
