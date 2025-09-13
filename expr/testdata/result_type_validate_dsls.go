package testdata

import . "goa.design/goa/v3/dsl"

// DuplicateResultTypeNamesDSL defines two result types that both end up with the
// same TypeName, which should be rejected by validation.
var DuplicateResultTypeNamesDSL = func() {
    var _ = API("MyAPI", func() {})

    var TypeA = ResultType("application/vnd.example.com.a", func() {
        TypeName("A")
        Attribute("a", Int)
    })

    var TypeB = ResultType("application/vnd.example.com.b", func() {
        TypeName("A") // intentional duplicate name
        Attribute("b", String)
    })

    var _ = Service("MyService", func() {
        Method("a", func() {
            Payload(Empty)
            Result(TypeA)
        })
        Method("b", func() {
            Payload(Empty)
            Result(TypeB)
        })
    })
}

