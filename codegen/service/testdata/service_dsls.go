package testdata

import (
	. "goa.design/goa/design"
	. "goa.design/goa/dsl"
)

var APayload = Type("APayload", func() {
	Attribute("IntField", Int)
	Attribute("StringField", String)
	Attribute("BooleanField", Boolean)
	Attribute("BytesField", Bytes)
	Attribute("OptionalField", String)
	Required("IntField", "StringField", "BooleanField", "BytesField")
})

var AResult = Type("AResult", func() {
	Attribute("IntField", Int)
	Attribute("StringField", String)
	Attribute("BooleanField", Boolean)
	Attribute("BytesField", Bytes)
	Attribute("OptionalField", String)
	Required("IntField", "StringField", "BooleanField", "BytesField")
})

var BPayload = Type("BPayload", func() {
	Attribute("ArrayField", ArrayOf(Boolean))
	Attribute("MapField", MapOf(Int, String))
	Attribute("ObjectField", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
	})
	Attribute("UserTypeField", ParentType)
})

var BResult = Type("BResult", func() {
	Attribute("ArrayField", ArrayOf(Boolean))
	Attribute("MapField", MapOf(Int, String))
	Attribute("ObjectField", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
	})
	Attribute("UserTypeField", ParentType)
})

var ParentType = Type("Parent", func() {
	Attribute("c", "Child")
})

var ChildType = Type("Child", func() {
	Attribute("p", "Parent")
})

var SingleMethodDSL = func() {
	Service("SingleMethod", func() {
		Method("A", func() {
			Payload(APayload)
			Result(AResult)
		})
	})
}

var MultipleMethodsDSL = func() {
	Service("MultipleMethods", func() {
		Method("A", func() {
			Payload(APayload)
			Result(AResult)
		})
		Method("B", func() {
			Payload(BPayload)
			Result(BResult)
		})
	})
}

var EmptyMethodDSL = func() {
	Service("Empty", func() {
		Method("Empty", func() {
		})
	})
}

var EmptyPayloadMethodDSL = func() {
	Service("EmptyPayload", func() {
		Method("EmptyPayload", func() {
			Result(AResult)
		})
	})
}

var EmptyResultMethodDSL = func() {
	Service("EmptyResult", func() {
		Method("EmptyResult", func() {
			Payload(APayload)
		})
	})
}

var ServiceErrorDSL = func() {
	Service("ServiceError", func() {
		Error("error")
		Method("A", func() {})
	})
}

var MultipleMethodsResultMultipleViewsDSL = func() {
	var RTWithViews = ResultType("application/vnd.result.multiple.views", func() {
		TypeName("MultipleViews")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("a")
			Attribute("b")
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	var RTWithSingleView = ResultType("application/vnd.result.single.view", func() {
		TypeName("SingleView")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("a")
			Attribute("b")
		})
	})
	Service("MultipleMethodsResultMultipleViews", func() {
		Method("A", func() {
			Payload(APayload)
			Result(RTWithViews)
		})
		Method("B", func() {
			Result(RTWithSingleView)
		})
	})
}

var ResultWithOtherResultMethodDSL = func() {
	var RTWithViews2 = ResultType("application/vnd.result.multiple.view.2", func() {
		TypeName("MultipleViews2")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
			Required("a")
		})
		View("default", func() {
			Attribute("a")
			Attribute("b")
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	var RTWithViews = ResultType("application/vnd.result.multiple.views", func() {
		TypeName("MultipleViews")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", RTWithViews2)
			Required("a", "b")
		})
		View("default", func() {
			Attribute("a")
			Attribute("b")
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	Service("ResultWithOtherResult", func() {
		Method("A", func() {
			Result(RTWithViews)
		})
	})
}
