package design

import "fmt"

// Attribute defines an attribute type, description and an optional validation DSL.
// When Attribute() is used in an action parameter definition all the arguments are optional and
// the corresponding attribute definition fields are inherited from the resource media type.
// Valid usage:
// * Attribute(dataType DataType, description string, dsl func())
// * Attribute(dataType DataType, dsl func())
// * Attribute(dsl func())
// * Attribute()
func Attribute(args ...interface{}) *AttributeDefinition {
	var dataType DataType
	var description string
	var dsl func()
	var ok bool
	if len(args) == 1 {
		if dsl, ok = args[0].(func()); !ok {
			invalidArgError("func()", args[0])
		}
	} else if len(args) == 2 {
		if dataType, ok = args[0].(DataType); !ok {
			invalidArgError("DataType", args[0])
		}
		if dsl, ok = args[1].(func()); !ok {
			invalidArgError("func()", args[1])
		}
	} else if len(args) == 3 {
		if dataType, ok = args[0].(DataType); !ok {
			invalidArgError("DataType", args[0])
		}
		if description, ok = args[1].(string); !ok {
			invalidArgError("string", args[1])
		}
		if dsl, ok = args[2].(func()); !ok {
			invalidArgError("func()", args[2])
		}
	} else if len(args) != 0 {
		appendError(fmt.Errorf("too many arguments in call to Attribute"))
	}
	att := AttributeDefinition{
		Type:        dataType,
		Description: description,
	}
	if dsl != nil {
		executeDSL(dsl, &att)
	}
	return &att
}
