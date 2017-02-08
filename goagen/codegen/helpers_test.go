package codegen_test

import (
	"testing"

	"github.com/goadesign/goa/goagen/codegen"

	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	Expect(codegen.KebabCase("testAa")).To(Equal("test-aa"))
	Expect(codegen.KebabCase("test-B")).To(Equal("test-b"))
	Expect(codegen.KebabCase("test_cA")).To(Equal("test-ca"))
	Expect(codegen.KebabCase("test_D")).To(Equal("test-d"))
	Expect(codegen.KebabCase("teste")).To(Equal("teste"))
	Expect(codegen.KebabCase("testABC")).To(Equal("testabc"))
	Expect(codegen.KebabCase("testAbc")).To(Equal("test-abc"))
}
