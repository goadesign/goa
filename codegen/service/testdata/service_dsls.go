package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var SingleMethodDSL = func() {
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
	Service("SingleMethod", func() {
		Method("A", func() {
			Payload(APayload)
			Result(AResult)
		})
	})
}

var MultipleMethodsDSL = func() {
	var _ = Type("Child", func() {
		Attribute("p", "Parent")
	})
	var ParentType = Type("Parent", func() {
		Attribute("c", "Child")
	})
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

var UnionMethodDSL = func() {
	var AUnion = Type("AUnion", func() {
		OneOf("Values", func() {
			Attribute("Int", Int)
			Attribute("String", String)
			Attribute("Boolean", Boolean)
			Attribute("Bytes", Bytes)
		})
	})
	Service("UnionService", func() {
		Method("A", func() {
			Payload(AUnion)
			Result(AUnion)
		})
	})
}

var MultiUnionMethodDSL = func() {
	var TypeA = Type("TypeA", func() {
		Attribute("a", Int)
	})
	var TypeB = Type("TypeB", func() {
		Attribute("b", String)
	})
	var Union = Type("Union", func() {
		OneOf("Values", func() {
			Attribute("a", TypeA)
			Attribute("b", TypeB)
		})
	})
	Service("MultiUnionService", func() {
		Method("MultiUnion", func() {
			Payload(Union)
			Result(Union)
		})
	})
}

var WithDefaultDSL = func() {
	Service("WithDefault", func() {
		Method("A", func() {
			Payload(func() {
				Attribute("IntField", Int, func() {
					Default(1)
				})
				Attribute("StringField", String, func() {
					Default("foo")
				})
				Attribute("OptionalField", String)
				Attribute("RequiredField", Float32)
				Required("RequiredField")
			})
			Result(func() {
				Attribute("IntField", Int, func() {
					Default(1)
				})
				Attribute("StringField", String, func() {
					Default("foo")
				})
				Attribute("OptionalField", String)
				Attribute("RequiredField", Float32)
				Required("RequiredField")
			})
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
	var AResult = Type("AResult", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})
	Service("EmptyPayload", func() {
		Method("EmptyPayload", func() {
			Result(AResult)
		})
	})
}

var EmptyResultMethodDSL = func() {
	var APayload = Type("APayload", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})

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

var CustomErrorsDSL = func() {
	var APayload = Type("APayload", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})

	var Result = ResultType("application/vnd.goa.error", func() {
		TypeName("Result")
		Attribute("a", String)
		ErrorName("b", String)
		Required("b")
	})
	Service("CustomErrors", func() {
		Method("A", func() {
			Error("primitive", String, "primitive error description")
			Error("user_type", APayload, "user type error description")
			Error("struct_error_name", Result, "struct error name description")
		})
	})
}

var CustomErrorsCustomFieldDSL = func() {
	var Result = ResultType("application/vnd.goa.error", func() {
		ErrorName("error", String, func() {
			Meta("struct:field:name", "ErrorCode")
		})
		Required("error")
	})
	Service("CustomErrorsCustomFields", func() {
		Method("A", func() {
			Error("struct_error_name", Result, "struct error name description")
		})
	})
}

var MultipleMethodsResultMultipleViewsDSL = func() {
	var APayload = Type("APayload", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})
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

var WithExplicitAndDefaultViewsDSL = func() {
	var RTWithViews = ResultType("application/vnd.result.multiple.views", func() {
		TypeName("MultipleViews")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
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
	Service("WithExplicitAndDefaultViews", func() {
		Method("A", func() {
			Result(RTWithViews)
		})
		Method("A", func() {
			Result(RTWithViews, func() {
				View("tiny")
			})
		})
	})
}

var ResultCollectionMultipleViewsMethodDSL = func() {
	var RTWithViews = ResultType("application/vnd.result.multiple.views", func() {
		TypeName("MultipleViews")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
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
	Service("ResultCollectionMultipleViewsMethod", func() {
		Method("A", func() {
			Result(CollectionOf(RTWithViews))
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

var ResultWithResultCollectionMethodDSL = func() {
	var RT2 = ResultType("application/vnd.result.2", func() {
		TypeName("RT2")
		Attributes(func() {
			Field(1, "c", String)
			Field(2, "d", Int)
			Field(3, "e", String)
			Required("c", "d")
		})
		View("default", func() {
			Attribute("c")
			Attribute("d")
		})
		View("extended", func() {
			Attribute("c")
			Attribute("d")
			Attribute("e")
		})
		View("tiny", func() {
			Attribute("d")
		})
	})
	var RT = ResultType("application/vnd.result", func() {
		TypeName("RT")
		Attributes(func() {
			Field(1, "a", CollectionOf(RT2))
		})
		View("default", func() {
			Attribute("a")
		})
		View("extended", func() {
			Attribute("a", func() {
				View("extended")
			})
		})
		View("tiny", func() {
			Attribute("a", func() {
				View("tiny")
			})
		})
	})
	Service("ResultWithResultTypeCollection", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultWithDashedMimeTypeMethodDSL = func() {
	var RT = ResultType("application/vnd.application.dashed-type", func() {
		Attributes(func() {
			Attribute("name")
		})
	})
	var _ = Service("ResultWithDashedMimeType", func() {
		Method("A", func() {
			Result(RT)
		})
		Method("list", func() {
			Result(func() {
				Attribute("items", CollectionOf(RT))
			})
		})
	})
}

var ResultWithOneOfTypeMethodDSL = func() {
	var Item = Type("Item", func() { Attribute("a") })
	var T = Type("T", func() { Attribute("message") })
	var U = Type("U", func() { Attribute("item", Item) })
	var RT = ResultType("application/vnd.result.oneof", func() {
		Attributes(func() {
			OneOf("result", func() {
				Attribute("t", T)
				Attribute("u", U)
			})
		})
	})
	Service("ResultWithOneOfType", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultWithInlineValidationDSL = func() {
	var RT = ResultType("application/vnd.result.inline.validation", func() {
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", Int)
		})
	})
	Service("ResultWithInlineValidation", func() {
		Method("A", func() {
			Result(RT)
		})
		Method("B", func() {
			Result(RT, func() {
				Required("a")
			})
		})
	})
}

var ForceGenerateTypeDSL = func() {
	var _ = Type("ForcedType", func() {
		Attribute("a", String)
		Meta("type:generate:force")
	})
	Service("ForceGenerateType", func() {
		Method("A", func() {})
	})
}

var ForceGenerateTypeExplicitDSL = func() {
	var _ = Type("ForcedType", func() {
		Attribute("a", String)
		Meta("type:generate:force", "ForceGenerateTypeExplicit")
	})
	Service("ForceGenerateTypeExplicit", func() {
		Method("A", func() {})
	})
}

var StreamingResultMethodDSL = func() {
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
	Service("StreamingResultService", func() {
		Method("StreamingResultMethod", func() {
			Payload(APayload)
			StreamingResult(AResult)
		})
	})
}

var StreamingResultWithViewsMethodDSL = func() {
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
	Service("StreamingResultWithViewsService", func() {
		Method("StreamingResultWithViewsMethod", func() {
			Payload(String)
			StreamingResult(RTWithViews)
		})
	})
}

var StreamingResultWithExplicitViewMethodDSL = func() {
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
	Service("StreamingResultWithExplicitViewService", func() {
		Method("StreamingResultWithExplicitViewMethod", func() {
			Payload(ArrayOf(Int32))
			StreamingResult(RTWithViews, func() {
				View("tiny")
			})
		})
	})
}

var StreamingResultNoPayloadMethodDSL = func() {
	var AResult = Type("AResult", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})

	Service("StreamingResultNoPayloadService", func() {
		Method("StreamingResultNoPayloadMethod", func() {
			StreamingResult(AResult)
		})
	})
}

var StreamingPayloadMethodDSL = func() {
	var _ = Type("Child", func() {
		Attribute("p", "Parent")
	})
	var ParentType = Type("Parent", func() {
		Attribute("c", "Child")
	})
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

	Service("StreamingPayloadService", func() {
		Method("StreamingPayloadMethod", func() {
			Payload(BPayload)
			StreamingPayload(APayload)
			Result(AResult)
		})
	})
}

var StreamingPayloadNoPayloadMethodDSL = func() {
	Service("StreamingPayloadNoPayloadService", func() {
		Method("StreamingPayloadNoPayloadMethod", func() {
			StreamingPayload(Any)
			Result(String)
		})
	})
}

var StreamingPayloadNoResultMethodDSL = func() {
	Service("StreamingPayloadNoResultService", func() {
		Method("StreamingPayloadNoResultMethod", func() {
			StreamingPayload(Int)
		})
	})
}

var StreamingPayloadResultWithViewsMethodDSL = func() {
	var APayload = Type("APayload", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})
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

	Service("StreamingPayloadResultWithViewsService", func() {
		Method("StreamingPayloadResultWithViewsMethod", func() {
			StreamingPayload(APayload)
			Result(RTWithViews)
		})
	})
}

var StreamingPayloadResultWithExplicitViewMethodDSL = func() {
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
	Service("StreamingPayloadResultWithExplicitViewService", func() {
		Method("StreamingPayloadResultWithExplicitViewMethod", func() {
			StreamingPayload(ArrayOf(String))
			Result(RTWithViews, func() {
				View("tiny")
			})
		})
	})
}

var BidirectionalStreamingMethodDSL = func() {
	var _ = Type("Child", func() {
		Attribute("p", "Parent")
	})
	var ParentType = Type("Parent", func() {
		Attribute("c", "Child")
	})
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
	Service("BidirectionalStreamingService", func() {
		Method("BidirectionalStreamingMethod", func() {
			Payload(BPayload)
			StreamingPayload(APayload)
			StreamingResult(AResult)
		})
	})
}

var BidirectionalStreamingNoPayloadMethodDSL = func() {
	Service("BidirectionalStreamingNoPayloadService", func() {
		Method("BidirectionalStreamingNoPayloadMethod", func() {
			StreamingPayload(String)
			StreamingResult(Int)
		})
	})
}

var BidirectionalStreamingResultWithViewsMethodDSL = func() {
	var APayload = Type("APayload", func() {
		Attribute("IntField", Int)
		Attribute("StringField", String)
		Attribute("BooleanField", Boolean)
		Attribute("BytesField", Bytes)
		Attribute("OptionalField", String)
		Required("IntField", "StringField", "BooleanField", "BytesField")
	})

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
	Service("BidirectionalStreamingResultWithViewsService", func() {
		Method("BidirectionalStreamingResultWithViewsMethod", func() {
			StreamingPayload(APayload)
			StreamingResult(RTWithViews)
		})
	})
}

var BidirectionalStreamingResultWithExplicitViewMethodDSL = func() {
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
	Service("BidirectionalStreamingResultWithExplicitViewService", func() {
		Method("BidirectionalStreamingResultWithExplicitViewMethod", func() {
			StreamingPayload(ArrayOf(Bytes))
			StreamingResult(RTWithViews, func() {
				View("default")
			})
		})
	})
}

var NamesWithSpacesDSL = func() {
	API("API With Spaces", func() {
		Server("Server With Spaces", func() {
			Services("Service With Spaces")
		})
	})
	var APayload = Type("Payload With Space", func() {
		Field(1, "String", String)
	})
	var AResult = ResultType("application/vnd.goa.result", func() {
		TypeName("Result With Space")
		Attributes(func() {
			Field(1, "Int", Int)
		})
	})
	Service("Service With Spaces", func() {
		Method("Method With Spaces", func() {
			Payload(APayload)
			Result(AResult)
			HTTP(func() {
				GET("/")
			})
			GRPC(func() {})
		})
	})
}

var PkgPathDSL = func() {
	var Foo = Type("Foo", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "foo")
	})
	Service("PkgPathMethod", func() {
		Method("A", func() {
			Payload(Foo)
			Result(Foo)
		})
	})
}

var PkgPathArrayDSL = func() {
	var Foo = Type("Foo", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "foo")
	})
	Service("PkgPathArrayMethod", func() {
		Method("A", func() {
			Payload(ArrayOf(Foo))
			Result(ArrayOf(Foo))
		})
	})
}

var PkgPathRecursiveDSL = func() {
	var Foo = Type("Foo", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "foo")
	})
	var RecursiveFoo = Type("RecursiveFoo", func() {
		Attribute("Foo", Foo)
		Meta("struct:pkg:path", "foo")
	})

	Service("PkgPathRecursiveMethod", func() {
		Method("A", func() {
			Payload(RecursiveFoo)
			Result(RecursiveFoo)
		})
	})
}

var PkgPathMultipleDSL = func() {
	var Bar = Type("Bar", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "bar")
	})
	var Baz = Type("Baz", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "baz")
	})

	Service("MultiplePkgPathMethod", func() {
		Method("A", func() {
			Payload(Bar)
			Result(Bar)
		})

		Method("B", func() {
			Payload(Baz)
			Result(Baz)
		})

		Method("EnvelopedB", func() {
			Payload(func() {
				Attribute("Baz", Baz)
			})
			Result(func() {
				Attribute("Baz", Baz)
			})
		})
	})
}

var PkgPathNoDirDSL = func() {
	var NoDir = Type("NoDir", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "")
	})

	Service("NoDirMethod", func() {
		Method("A", func() {
			Payload(NoDir)
			Result(NoDir)
		})
	})
}

var PkgPathDupeDSL = func() {
	var Foo = Type("Foo", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "foo")
	})

	Service("PkgPathDupeMethod", func() {
		Method("A", func() {
			Payload(Foo)
			Result(Foo)
		})
		Method("B", func() {
			Payload(Foo)
			Result(Foo)
		})
	})
	Service("PkgPathDupeMethod2", func() {
		Method("A", func() {
			Payload(Foo)
			Result(Foo)
		})
		Method("B", func() {
			Payload(Foo)
			Result(Foo)
		})
	})
}

var PkgPathPayloadAttributeDSL = func() {
	var Foo = Type("Foo", func() {
		Attribute("IntField", Int)
		Meta("struct:pkg:path", "foo")
	})
	var Bar = Type("Bar", func() {
		Attribute("Foo", Foo)
	})

	Service("PkgPathPayloadAttributeDSL", func() {
		Method("Foo", func() {
			Payload(Bar)
			Result(Bar)
		})
	})
}
