package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var TestUnionDSL = func() {
	var (
		SomeType = Type("SomeType", func() {
			Attribute("someField", String)
		})

		UnionString = Type("UnionString", func() {
			OneOf("UnionString", func() {
				Attribute("String", String)
			})
		})
		UnionString2 = Type("UnionString2", func() {
			OneOf("UnionString2", func() {
				Attribute("String", String)
			})
		})
		UnionStringInt = Type("UnionStringInt", func() {
			OneOf("UnionStringInt", func() {
				Attribute("String", String)
				Attribute("Int", Int)
			})
		})
		UnionStringInt2 = Type("UnionStringInt2", func() {
			OneOf("UnionStringInt2", func() {
				Attribute("String", String)
				Attribute("Int", Int)
			})
		})
		UnionSomeType = Type("UnionSomeType", func() {
			OneOf("UnionSomeType", func() {
				Attribute("SomeType", SomeType)
			})
		})
		UnionSomeType2 = Type("UnionSomeType2", func() {
			OneOf("UnionSomeType2", func() {
				Attribute("SomeType", SomeType)
			})
		})

		_ = Type("Container", func() {
			Attribute("UnionString", UnionString)
			Attribute("UnionString2", UnionString2)
			Attribute("UnionStringInt", UnionStringInt)
			Attribute("UnionStringInt2", UnionStringInt2)
			Attribute("UnionSomeType", UnionSomeType)
			Attribute("UnionSomeType2", UnionSomeType2)
		})

		_ = Type("UnionUserType", func() {
			Attribute("Type", String)
			Attribute("Value", String)
			Required("Type", "Value")
		})

		// Union with custom type and value keys
		UnionCustomKeys = Type("UnionCustomKeys", func() {
			OneOf("UnionCustomKeys", func() {
				Meta("oneof:type:field", "kind")
				Meta("oneof:value:field", "data")
				Attribute("String", String)
				Attribute("Int", Int)
			})
		})

		// Union with different custom keys for testing
		UnionPaymentMethod = Type("UnionPaymentMethod", func() {
			OneOf("UnionPaymentMethod", func() {
				Meta("oneof:type:field", "paymentType")
				Meta("oneof:value:field", "details")
				Attribute("CreditCard", SomeType)
				Attribute("PayPal", String)
			})
		})

		_ = Type("CustomKeysContainer", func() {
			Attribute("UnionCustomKeys", UnionCustomKeys)
			Attribute("UnionPaymentMethod", UnionPaymentMethod)
		})
	)
}

var ConstructorUnionCollectionsDSL = func() {
	var TextPayload = Type("ConstructorUnionCollectionsTextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("ConstructorUnionCollectionsJSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	var Choice = OneOf(TextPayload, JSONPayload)

	var _ = Type("ConstructorUnionCollections", func() {
		Attribute("ArrayChoices", ArrayOf(Choice))
		Attribute("MapChoices", MapOf(String, Choice))
	})
}

var DeclarationAndConstructorUnionSymmetryDSL = func() {
	var TextPayload = Type("DeclarationConstructorTextPayload", func() {
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("DeclarationConstructorJSONPayload", func() {
		Attribute("message", String)
		Required("message")
	})
	var DeclaredChoice = Type("DeclarationConstructorDeclaredChoice", func() {
		OneOf("Value", func() {
			Attribute("DeclarationConstructorTextPayload", TextPayload)
			Attribute("DeclarationConstructorJSONPayload", JSONPayload)
		})
	})

	var _ = Type("DeclarationUnionContainer", func() {
		Attribute("Choice", DeclaredChoice)
		Required("Choice")
	})
	var _ = Type("ConstructorUnionContainer", func() {
		Attribute("Choice", OneOf(TextPayload, JSONPayload))
		Required("Choice")
	})
}

var DeclarationAndConstructorUnionTaggedSymmetryDSL = func() {
	var TextPayload = Type("TaggedDeclarationConstructorTextPayload", func() {
		Meta("oneof:type:tag", "text")
		Attribute("text", String)
		Required("text")
	})
	var JSONPayload = Type("TaggedDeclarationConstructorJSONPayload", func() {
		Meta("oneof:type:tag", "json")
		Attribute("message", String)
		Required("message")
	})
	var DeclaredChoice = Type("TaggedDeclarationConstructorDeclaredChoice", func() {
		OneOf("Value", func() {
			Attribute("TaggedDeclarationConstructorTextPayload", TextPayload)
			Attribute("TaggedDeclarationConstructorJSONPayload", JSONPayload)
		})
	})

	var _ = Type("TaggedDeclarationUnionContainer", func() {
		Attribute("Choice", DeclaredChoice)
		Required("Choice")
	})
	var _ = Type("TaggedConstructorUnionContainer", func() {
		Attribute("Choice", OneOf(TextPayload, JSONPayload))
		Required("Choice")
	})
}
