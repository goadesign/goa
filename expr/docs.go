/*
Package expr defines expressions and data types used by the DSL and the code
generators. The expressions implement the Preparer, Validator, and Finalizer
interfaces. The code generators use the finalized expressions to generate the
final output.

The data types defined in the expr package are primitive types corresponding
to scalar values (bool, string, integers, and numbers), array types which
repressent a collection of items, map types which represent maps of key/value
pairs, and object types describing data structures with fields. The package
also defines user types which can also be a result types. A result type is a
user type used to describe response messages rendered using a view.
*/
package expr
