package design

// An action parameter (path element, query string or payload)
type ActionParam struct {
	Name   string               // Name of parameter
	Member *AttributeDefinition // Type and validations (if any)
}

// A map of action parameters indexed by name
type ActionParams map[string]*ActionParam

// Null sets the action parameter type to Null
func (p *ActionParam) Null() *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: Null}
	} else {
		p.Member.Type = Null
	}
	return p
}

// Boolean sets the action parameter type to Boolean
func (p *ActionParam) Boolean() *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: Boolean}
	} else {
		p.Member.Type = Boolean
	}
	return p
}

// Integer sets the action parameter type to Integer
func (p *ActionParam) Integer() *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: Integer}
	} else {
		p.Member.Type = Integer
	}
	return p
}

// Number sets the action parameter type to Number
func (p *ActionParam) Number() *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: Number}
	} else {
		p.Member.Type = Number
	}
	return p
}

// String sets the action parameter type to String
func (p *ActionParam) String() *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: String}
	} else {
		p.Member.Type = String
	}
	return p
}

// Array sets the action parameter type to Array
func (p *ActionParam) Array(elemType DataType) *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: &Array{ElemType: elemType}}
	} else {
		p.Member.Type = &Array{ElemType: elemType}
	}
	return p
}

// Object sets the action parameter type to Object
func (p *ActionParam) Object(obj Object) *ActionParam {
	if p.Member == nil {
		p.Member = &AttributeDefinition{Type: obj}
	} else {
		p.Member.Type = obj
	}
	return p
}
