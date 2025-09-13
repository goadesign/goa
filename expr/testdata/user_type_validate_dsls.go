package testdata

import . "goa.design/goa/v3/dsl"

// DuplicateUserTypeNamesDSL defines two user types that both set the same
// TypeName, which should be rejected by validation.
var DuplicateUserTypeNamesDSL = func() {
    var _ = API("MyAPI", func() {})

    var P1 = Type("PayloadOne", func() {
        TypeName("P")
        Attribute("a", Int)
    })

    var P2 = Type("PayloadTwo", func() {
        TypeName("P") // duplicate type name
        Attribute("b", String)
    })

    Service("Svc", func() {
        Method("m1", func() {
            Payload(P1)
            Result(Empty)
        })
        Method("m2", func() {
            Payload(P2)
            Result(Empty)
        })
    })
}

