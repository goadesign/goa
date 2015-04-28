package design

// attributeDSL corresponds to a single attribute DSL definition
type attributeDSL struct {
	name string   // Name of attribute
	typ  DataType // Attribute type
	dsl  func()   // DSL
}

// Attribute creates an attribute DSL
func Attribute(name string, typ DataType, dsl func()) {
	return attributeDSL{
		name: name,
		typ:  typ,
		dsl:  dsl,
	}
}

// Run DSL to produce API definition
func (d *attributeDSL) execute() (*Attribute, error) {
	att := &Attribute{
		Name:     name,
		DataType: typ,
	}
	ctxStack = append(ctxStack, att)
	ctx = att
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	ctx = ctxStack[len(ctxStack)-1]
	if dslError != nil {
		return nil, dslError
	}
	return att, nil
}
