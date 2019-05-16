/*
Package eval implements a DSL engine for executing arbitrary Go DSLs.

DSLs executed via eval consist of package functions that build up expressions
upon execution.

A DSL that allows describing a service and its methods could look like this:

    var _ = Service("service name")       // Defines the service "service	name"

    var _ = Method("method name", func() {  // Defines the method "method name"
        Description("some method description") // Sets the method description
    })

DSL keywords are simply package functions that can be nested using anonymous
functions as last argument. Upon execution the DSL functions create expression
structs. The expression structs created by the top level functions on process
start (both Service and Method in this example) should be stored in special
expressions called root expressions. The DSL implements both the expression and
root expression structs, the only requirement is that they implement the eval
package Expression and Root interfaces respectively.

Keeping with the example above, Method creates instances of the following
MethodExpression struct:

    type MethodExpression struct {
        Name string
        DSLFunc func()
    }

where Name gets initialized with the first argument and DSLFunc with the
second. ServiceExpression is the root expression that contains the instances
of MethodExpression created by the Method function:

    type ServiceExpression struct {
        Name string
        Methods []eval.Expression
    }

The Method DSL function simply initializes a MethodExpression and stores it in
the Methods field of the root ServiceExpression:

    func Method(name string, fn func()) {
        ep := &MethodExpression{Name: name, DSLFunc: fn}
        Design.Methods = append(Design.Methods, ep)
    }

where Design is a package variable holding the ServiceExpression root
expression:

    // Design is the DSL root expression.
    var Design *ServiceExpression = &ServiceExpression{}

The Service function simply sets the Name field of Service:

    func Service(name string) {
        Design.Name = name
    }

Once the process is loaded the Design package variable contains an instance of
ServiceExpression which in turn contains all the instances of MethodExpression
that were created via the Method function. Note that at this point the
Description function used in the Method DSL hasn't run yet as it is called by
the anonymous function stored in the DSLFunc field of each MethodExpression
instance. This is where the RunDSL function of package eval comes in.

RunDSL iterates over the initial set of root expressions and calls the WalkSets
method exposed by the Root interface. This method lets the DSL engine iterate
over the sub-expressions that were initialized when the process loaded.

In this example the ServiceExpression implementation of WalkSets simply passes
the Methods field to the iterator:

    func (se *ServiceExpression) WalkSets(it eval.SetWalker) {
        it(se.Methods)
    }

Each expression in an expression set may optionally implement the Source,
Preparer, Validator, and Finalizer interfaces:

- Expressions that are initialized via a child DSL implement Source which
provides RunDSL with the corresponding anonymous function.

- Expressions that need to be prepared implement the Preparer interface.

- Expressions that need to be validated implement the Validator interface.

- Expressions that require an additional pass after validation implement the
Finalizer interface.

In our example MethodExpression implements Source and return its DSLFunc member
in the implementation of the Source interface DSL function:

    func (ep *MethodExpression) Source() func() {
        return ep.DSLFunc
    }

MethodExpression could also implement the Validator Validate method to check
that the name of the method is not empty for example.

The execution of the DSL thus happens in four phases: in the first phase
RunDSL executes all the DSLs of all the source expressions in each expression
set. In this initial phase the DSLs being executed may append to the expression
set and/or may register new expression roots. In the second phase RunDSL
prepares all the preparer expressions. In the third phase RunDSL validates all
the validator expressions and in the last phase it calls Finalize on all the
finalizer expressions.

The eval package exposes functions that the implementation of the DSL can take
advantage of to report errors, such as ReportError, InvalidArg, and
IncompatibleDSL. The engine records the errors being reported but keeps running
the current phase so that multiple errors may be reported at once. This means
that the DSL implementation must maintain a consistent state for the duration
of one iteration even though some input may be incorrect (for example it may
elect to create default value expressions instead of leaving them nil to avoid
panics later on).

The package exposes other helper functions such as Execute which allows running
a DSL function on demand.
*/
package eval
