package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var ResultWithMultipleViewsDSL = func() {
	var RT = ResultType("application/vnd.result", func() {
		TypeName("ResultType")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
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
	Service("ResultWithMultipleViews", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultCollectionMultipleViewsDSL = func() {
	var RT = ResultType("application/vnd.result", func() {
		TypeName("ResultType")
		Attributes(func() {
			Attribute("a", String)
			Attribute("b", String)
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
	Service("ResultCollectionMultipleViews", func() {
		Method("A", func() {
			Result(CollectionOf(RT))
		})
	})
}

var ResultWithUserTypeDSL = func() {
	var UT = Type("UserType", func() {
		Attribute("a")
	})
	var RT = ResultType("application/vnd.result", func() {
		TypeName("ResultType")
		Attributes(func() {
			Attribute("a", UT)
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
	Service("ResultWithUserType", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultWithResultTypeDSL = func() {
	var UT = Type("UserType", func() {
		Attribute("p")
	})
	var RT3 = ResultType("application/vnd.result.3", func() {
		TypeName("RT3")
		Attributes(func() {
			Attribute("x", ArrayOf(String))
			Attribute("y", MapOf(Int, UT))
			Attribute("z", String)
			Required("x", "y", "z")
		})
		View("default", func() {
			Attribute("x")
			Attribute("y")
		})
		View("tiny", func() {
			Attribute("x")
		})
	})
	var RT2 = ResultType("application/vnd.result.2", func() {
		TypeName("RT2")
		Attributes(func() {
			Attribute("c", String)
			Attribute("d", UT)
			Attribute("e", String)
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
			Attribute("a", String)
			Attribute("b", RT2)
			Attribute("c", RT3)
			Required("b", "c")
		})
		View("default", func() {
			Attribute("a")
			Attribute("b", func() {
				View("extended")
			})
			Attribute("c")
		})
		View("tiny", func() {
			Attribute("b", func() {
				View("tiny")
			})
			Attribute("c")
		})
	})
	Service("ResultWithResultType", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultWithRecursiveResultTypeDSL = func() {
	var RT = ResultType("application/vnd.result", func() {
		TypeName("RT")
		Attributes(func() {
			Attribute("a", "RT")
			Required("a")
		})
		View("default", func() {
			Attribute("a", func() {
				View("tiny")
			})
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	Service("ResultWithRecursiveResultType", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultWithRecursiveCollectionOfResultTypeDSL = func() {
	var SomeRT = ResultType("application/vnd.some_result", func() {
		TypeName("SomeRT")
		Attributes(func() {
			Attribute("a", CollectionOf("SomeRT"))
			Required("a")
		})
		View("default", func() {
			Attribute("a", func() {
				View("tiny")
			})
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	var AnotherRT = ResultType("application/vnd.another_result", func() {
		Attributes(func() {
			Attribute("a", CollectionOf("application/vnd.another_result"))
			Required("a")
		})
	})
	Service("ResultWithRecursiveCollectionOfResultType", func() {
		Method("A", func() {
			Result(SomeRT)
		})
		Method("B", func() {
			Result(AnotherRT)
		})
	})
}

var ResultWithMultipleMethodsDSL = func() {
	var RT = ResultType("application/vnd.some_result", func() {
		TypeName("RT")
		Attributes(func() {
			Attribute("a")
		})
	})
	Service("ResultWithMultipleMethods", func() {
		Method("A", func() {
			Result(RT)
		})
		Method("B", func() {
			Result(RT)
		})
	})
}

var ResultWithEnumTypeDSL = func() {
	var T = Type("UserType", String, func()  {
		Enum("a", "b")
	})
	var RT = ResultType("application/vnd.result", func() {
		Attributes(func() {
			Attribute("t", ArrayOf(T))
		})
	})
	Service("ResultWithEnumType", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}

var ResultWithCustomFieldsDSL = func() {
	var RT = ResultType("application/vnd.result", func() {
		TypeName("RT")
		Attributes(func() {
			Attribute("a", String, func() {
				Meta("struct:field:name", "CustomA")
			})
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
	Service("ResultWithCustomFields", func() {
		Method("A", func() {
			Result(RT)
		})
	})
}
