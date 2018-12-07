package testdata

import (
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
	Service("StreamingResultNoPayloadService", func() {
		Method("StreamingResultNoPayloadMethod", func() {
			StreamingResult(AResult)
		})
	})
}

var StreamingPayloadMethodDSL = func() {
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
