/*
Package eval implements a DSL engine for executing arbitrary Go DSLs.

DSLs executed via eval consist of package functions that build up expressions upon execution.
A DSL that allows describing a service and its endpoints could look like this:

    var _ = Service("service name")              // Defines the service "service name"

    var _ = Endpoint("endpoint name", func() {   // Defines the endpoint "endpoint name"
        Description("some endpoint description") // Sets the endpoint description
    })

DSL keywords are simply package functions that can be nested using anonymous functions as last
argument. Upon execution the DSL functions create expression structs. The expression structs created
by the top level functions on process start (both Service and Endpoint in this example) should be
stored in special expressions called root expressions. The DSL implements both the expression and
root expression structs, the only requirement is that they implement the eval package Expression
and Root interfaces respectively.

Keeping with the example above, Endpoint creates instances of the following EndpointExpression
struct:

    type EndpointExpression struct {
        Name string
        DSLFunc func()
    }

where Name gets initialized with the first argument and DSLFunc with the second. ServiceExpression
is the root expression that contains the instances of EndpointExpression created by the Endpoint
function:

    type ServiceExpression struct {
        Name string
        Endpoints []eval.Expression
    }

The Endpoint DSL function simply initializes a EndpointExpression and stores it in the Endpoints
field of the root ServiceExpression:

    func Endpoint(name string, dsl func()) {
        ep := &EndpointExpression{Name: name, DSLFunc: dsl}
        Design.Endpoints = append(Design.Endpoints, ep)
    }

where Design is a package variable holding the ServiceExpression root expression:

    // Design is the DSL root expression.
    var Design *ServiceExpression = &ServiceExpression{}

The Service function simply sets the Name field of Service:

    func Service(name string) {
        Design.Name = name
    }

Once the process is loaded the Design package variable contains an instance of ServiceExpression
which in turn contains all the instances of EndpointExpression that were created via the Endpoint
function. Note that at this point the Description function used in the Endpoint DSL hasn't run yet
as it is called by the anonymous function stored in the DSLFunc field of each EndpointExpression
instance. This is where the RunDSL function of package eval comes in.

RunDSL iterates over the initial set of root expressions and calls the IterateSets method exposed
by the Root interface. This method lets the DSL engine iterate over the sub-expressions that were
initialized when the process loaded.

In this example the ServiceExpression implementation of IterateSets simply passes the Endpoints
field to the iterator:

    func (se *ServiceExpression) IterateSets(it eval.SetIterator) {
        it(se.Endpoints)
    }

Each expression in an expression set may optionally implement the Source, Validator and Finalizer
interfaces:

- Expressions that are initialized via a child DSL implement Source which provides RunDSL with the
  corresponding anonymous function.

- Expressions that need to be validated implement the Validator interface.

- Expressions that require an additional pass after validation implement the Finalizer interface.

In our example EndpointExpression implements Source and return its DSLFunc member in the
implementation of the Source interface DSL function:

    func (ep *EndpointExpression) Source() func() {
        return ep.DSLFunc
    }

EndpointExpression could also implement the Validator Validate method to check that the name of the
endpoint is not empty for example.

The execution of the DSL thus happens in three phases: in the first phase RunDSL executes all the
DSLs of all the source expressions in each expression set. In this initial phase the DSLs being
executed may append to the expression set and/or may register new expression roots. In the second
phase RunDSL validates all the validator expressions and in the last phase it calls Finalize on all
the finalizer expressions.

The eval package exposes functions that the implementation of the DSL can take advantage of to
report errors, such as ReportError, InvalidArg and IncompatibleDSL. The engine records the errors
being reported but keeps running the current phase so that multiple errors may be reported at once.
This means that the DSL implementation must maintain a consistent state for the duration of one
iteration even though some input may be incorrect (for example it may elect to create default value
expressions instead of leaving them nil to avoid panics later on).

The package exposes other helper functions such as Execute which allows running a DSL function
manually or IsTop which reports whether the expression being currently built is a top level
expression (such as Service and Endpoint in our example).
*/
package eval
