package codegen_test

import (
	"testing"

	"github.com/goadesign/goa/goagen/codegen"

	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	var (
		stringTestA = "testAa"
		stringTestB = "test-B"
		stringTestC = "test_cA"
		stringTestD = "test_D"
		stringTestE = "teste"
		stringTestF = "testABC"
		stringTestG = "testAbc"

		expectedTestA = "test-aa"
		expectedTestB = "test-b"
		expectedTestC = "test-ca"
		expectedTestD = "test-d"
		expectedTestE = "teste"
		expectedTestF = "testabc"
		expectedTestG = "test-abc"
	)

	Expect(codegen.KebabCase(stringTestA)).To(Equal(expectedTestA))
	Expect(codegen.KebabCase(stringTestB)).To(Equal(expectedTestB))
	Expect(codegen.KebabCase(stringTestC)).To(Equal(expectedTestC))
	Expect(codegen.KebabCase(stringTestD)).To(Equal(expectedTestD))
	Expect(codegen.KebabCase(stringTestE)).To(Equal(expectedTestE))
	Expect(codegen.KebabCase(stringTestF)).To(Equal(expectedTestF))
	Expect(codegen.KebabCase(stringTestG)).To(Equal(expectedTestG))
}
