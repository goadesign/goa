package client_test

import (
	"context"

	"github.com/goadesign/goa/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("client", func() {
	Context("with background context", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.Background()
		})

		Context("ContextRequestID", func() {
			It("should have empty request ID", func() {
				Expect(client.ContextRequestID(ctx)).To(BeEmpty())
			})
		})

		Context("ContextWithRequestID", func() {
			It("should generate a new request ID", func() {
				newCtx, reqID := client.ContextWithRequestID(ctx)
				Expect(reqID).ToNot(BeEmpty())
				Expect(client.ContextRequestID(newCtx)).To(Equal(reqID))
			})
		})

		Context("SetContextRequestID", func() {
			It("should set a custom request ID", func() {
				const customID = "foo"
				newCtx := client.SetContextRequestID(ctx, customID)
				Expect(newCtx).ToNot(Equal(ctx))
				Expect(client.ContextRequestID(newCtx)).To(Equal(customID))

				// request ID should not need to be generated again. the same context should be returned instead.
				newCtx2, reqID := client.ContextWithRequestID(newCtx)
				Expect(newCtx).To(Equal(newCtx2))
				Expect(reqID).To(Equal(customID))
			})
		})
	})
})
