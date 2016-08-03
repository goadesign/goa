package json_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/goadesign/goa/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JsonEncoding", func() {

	Describe("handle goa/uuid/UUID", func() {
		name := "Test"
		id, _ := uuid.FromString("c0586f01-87b5-462b-a673-3b2dcf619091")

		type Payload struct {
			ID   uuid.UUID
			Name string
		}

		It("encode", func() {
			data := Payload{
				id,
				name,
			}

			var b bytes.Buffer
			encoder := json.NewEncoder(&b)
			encoder.Encode(data)
			s := b.String()

			Ω(s).Should(ContainSubstring(id.String()))
			Ω(s).Should(ContainSubstring(name))
		})

		It("decode", func() {
			encoded := fmt.Sprintf(`{"ID":"%s","Name":"%s"}`, id, name)

			var payload Payload
			var b bytes.Buffer
			b.WriteString(encoded)

			decoder := json.NewDecoder(&b)
			decoder.Decode(&payload)

			Ω(payload.ID.String()).Should(Equal(id.String()))
			Ω(payload.Name).Should(Equal(name))
		})

		It("round trip", func() {
			data := Payload{
				id,
				name,
			}

			var payload Payload
			var b bytes.Buffer

			encoder := json.NewEncoder(&b)
			encoder.Encode(data)

			decoder := json.NewDecoder(&b)
			decoder.Decode(&payload)

			Ω(payload.ID.String()).Should(Equal(data.ID.String()))
			Ω(payload.Name).Should(Equal(data.Name))
		})

	})
})
