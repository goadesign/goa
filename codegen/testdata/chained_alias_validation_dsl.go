package testdata

import . "goa.design/goa/v3/dsl"

// ChainedAliasValidationDSL defines user type alias chains two and three
// levels deep. It is used to verify that validation codegen merges alias
// chain validations into the generated code without mutating the design
// expression tree.
var ChainedAliasValidationDSL = func() {
	var (
		// String alias chain: each level carries its own validation.
		ChainBase = Type("ChainBase", String, func() {
			Pattern("^[a-z]+$")
			MinLength(2)
		})

		ChainMid = Type("ChainMid", ChainBase, func() {
			Enum("ab", "abc", "abcd")
		})

		// ChainPass adds a level with no validation of its own so the
		// effective validation comes entirely from deeper chain levels.
		ChainPass = Type("ChainPass", ChainMid)

		// Object type chain: each level adds required fields.
		ChainObjBase = Type("ChainObjBase", func() {
			Attribute("name", String, func() {
				MinLength(2)
				Pattern("^[a-z]+$")
			})
			Attribute("count", Int, func() {
				Minimum(1)
			})
			Attribute("tag", String)
			Required("name")
		})

		ChainObjMid = Type("ChainObjMid", ChainObjBase, func() {
			Required("count")
		})

		ChainObjLeaf = Type("ChainObjLeaf", ChainObjMid, func() {
			Required("tag")
		})

		ChainChild = Type("ChainChild", func() {
			Attribute("leaf", ChainMid)
			Required("leaf")
		})
	)

	_ = Type("ChainHolder", func() {
		Attribute("req_mid", ChainMid)
		Attribute("mid", ChainMid)
		Attribute("pass", ChainPass)
		Attribute("obj", ChainObjLeaf)
		Required("req_mid")
	})

	_ = Type("ChainParent", func() {
		Attribute("child", ChainChild)
	})

	_ = Type("ChainObjHolder", ChainObjLeaf)
}
