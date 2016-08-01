package json_test

import (
	"bytes"
	"encoding/json"

	"github.com/goadesign/goa/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JsonEncoding", func() {

	Describe("handle goa/uuid/UUID", func() {
		type Payload struct {
			ID   uuid.UUID
			Name string
		}
		data := Payload{
			uuid.NewV4(),
			"Test",
		}
		var encoded string

		It("encode", func() {
			var b bytes.Buffer
			encoder := json.NewEncoder(&b)
			encoder.Encode(data)
			encoded = b.String()

			Ω(encoded).Should(ContainSubstring(data.ID.String()))
			Ω(encoded).Should(ContainSubstring(data.Name))
		})

		It("decode", func() {
			var b bytes.Buffer
			b.WriteString(encoded)
			decoder := json.NewDecoder(&b)

			var payload Payload
			decoder.Decode(&payload)

			Ω(payload.ID.String()).Should(Equal(data.ID.String()))
			Ω(payload.Name).Should(Equal(data.Name))
		})

	})
})
