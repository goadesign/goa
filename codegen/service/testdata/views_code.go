package testdata

const ResultWithMultipleViewsCode = `// ResultTypeView is a type which is projected based on a view.
type ResultTypeView struct {
	A *string
	B *string
}

// ResultType is the viewed result type that projects ResultTypeView based on a
// view.
type ResultType struct {
	*ResultTypeView
	// View to render
	View string
}

// AsDefault projects viewed result type ResultType using the default view.
func (result *ResultType) AsDefault() *ResultType {
	t := &ResultTypeView{
		A: result.A,
		B: result.B,
	}
	return &ResultType{
		ResultTypeView: t,
		View:           "default",
	}
}

// AsTiny projects viewed result type ResultType using the tiny view.
func (result *ResultType) AsTiny() *ResultType {
	t := &ResultTypeView{
		A: result.A,
	}
	return &ResultType{
		ResultTypeView: t,
		View:           "tiny",
	}
}

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "default":
		if result.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
		}
		if result.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
		}
	case "tiny":
		if result.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
		}
	}
	return
}
`

var ResultWithUserTypeCode = `// ResultTypeView is a type which is projected based on a view.
type ResultTypeView struct {
	A *UserType
	B *string
}

// ResultType is the viewed result type that projects ResultTypeView based on a
// view.
type ResultType struct {
	*ResultTypeView
	// View to render
	View string
}

// UserType is a type that runs validations on a projected type.
type UserType struct {
	A *string
}

// AsDefault projects viewed result type ResultType using the default view.
func (result *ResultType) AsDefault() *ResultType {
	t := &ResultTypeView{
		B: result.B,
	}
	if result.A != nil {
		t.A = marshalUserTypeToUserType(result.A)
	}
	return &ResultType{
		ResultTypeView: t,
		View:           "default",
	}
}

// AsTiny projects viewed result type ResultType using the tiny view.
func (result *ResultType) AsTiny() *ResultType {
	t := &ResultTypeView{}
	if result.A != nil {
		t.A = marshalUserTypeToUserType(result.A)
	}
	return &ResultType{
		ResultTypeView: t,
		View:           "tiny",
	}
}

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "default":
		if result.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
		}
	case "tiny":
		if result.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
		}
	}
	return
}

// marshalUserTypeToUserType builds a value of type *UserType from a value of
// type *UserType.
func marshalUserTypeToUserType(v *UserType) *UserType {
	if v == nil {
		return nil
	}
	res := &UserType{
		A: v.A,
	}

	return res
}
`

const ResultWithResultTypeCode = `// RTView is a type which is projected based on a view.
type RTView struct {
	A *string
	B *RT2
	C *RT3
}

// RT is the viewed result type that projects RTView based on a view.
type RT struct {
	*RTView
	// View to render
	View string
}

// RT2View is a type which is projected based on a view.
type RT2View struct {
	C *string
	D *UserType
	E *string
}

// RT2 is the viewed result type that projects RT2View based on a view.
type RT2 struct {
	*RT2View
	// View to render
	View string
}

// UserType is a type that runs validations on a projected type.
type UserType struct {
	P *string
}

// RT3View is a type which is projected based on a view.
type RT3View struct {
	X []string
	Y map[int]*UserType
	Z *string
}

// RT3 is the viewed result type that projects RT3View based on a view.
type RT3 struct {
	*RT3View
	// View to render
	View string
}

// AsDefault projects viewed result type RT using the default view.
func (result *RT) AsDefault() *RT {
	t := &RTView{
		A: result.A,
	}
	if result.B != nil {
		t.B = result.B.AsExtended()
	}

	if result.C != nil {
		t.C = result.C.AsDefault()
	}

	return &RT{
		RTView: t,
		View:   "default",
	}
}

// AsTiny projects viewed result type RT using the tiny view.
func (result *RT) AsTiny() *RT {
	t := &RTView{}
	if result.B != nil {
		t.B = result.B.AsTiny()
	}

	if result.C != nil {
		t.C = result.C.AsDefault()
	}

	return &RT{
		RTView: t,
		View:   "tiny",
	}
}

// AsDefault projects viewed result type RT2 using the default view.
func (result *RT2) AsDefault() *RT2 {
	t := &RT2View{
		C: result.C,
	}
	if result.D != nil {
		t.D = marshalUserTypeToUserType(result.D)
	}
	return &RT2{
		RT2View: t,
		View:    "default",
	}
}

// AsExtended projects viewed result type RT2 using the extended view.
func (result *RT2) AsExtended() *RT2 {
	t := &RT2View{
		C: result.C,
		E: result.E,
	}
	if result.D != nil {
		t.D = marshalUserTypeToUserType(result.D)
	}
	return &RT2{
		RT2View: t,
		View:    "extended",
	}
}

// AsTiny projects viewed result type RT2 using the tiny view.
func (result *RT2) AsTiny() *RT2 {
	t := &RT2View{}
	if result.D != nil {
		t.D = marshalUserTypeToUserType(result.D)
	}
	return &RT2{
		RT2View: t,
		View:    "tiny",
	}
}

// AsDefault projects viewed result type RT3 using the default view.
func (result *RT3) AsDefault() *RT3 {
	t := &RT3View{}
	if result.X != nil {
		t.X = make([]string, len(result.X))
		for j, val := range result.X {
			t.X[j] = val
		}
	}
	if result.Y != nil {
		t.Y = make(map[int]*UserType, len(result.Y))
		for key, val := range result.Y {
			tk := key
			tv := &UserType{
				P: val.P,
			}
			t.Y[tk] = tv
		}
	}
	return &RT3{
		RT3View: t,
		View:    "default",
	}
}

// AsTiny projects viewed result type RT3 using the tiny view.
func (result *RT3) AsTiny() *RT3 {
	t := &RT3View{}
	if result.X != nil {
		t.X = make([]string, len(result.X))
		for j, val := range result.X {
			t.X[j] = val
		}
	}
	return &RT3{
		RT3View: t,
		View:    "tiny",
	}
}

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "default":
		if result.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
		}
		if result.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
		}
		if result.B != nil {
			if err2 := result.B.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if result.C != nil {
			if err2 := result.C.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	case "tiny":
		if result.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
		}
		if result.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
		}
		if result.B != nil {
			if err2 := result.B.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if result.C != nil {
			if err2 := result.C.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// Validate runs the validations defined on RT2.
func (result *RT2) Validate() (err error) {
	switch result.View {
	case "default":
		if result.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
		}
		if result.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
		}
	case "extended":
		if result.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
		}
		if result.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
		}
	case "tiny":
		if result.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
		}
	}
	return
}

// Validate runs the validations defined on RT3.
func (result *RT3) Validate() (err error) {
	switch result.View {
	case "default":
		if result.X == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
		}
		if result.Y == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("y", "result"))
		}
	case "tiny":
		if result.X == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
		}
	}
	return
}

// marshalUserTypeToUserType builds a value of type *UserType from a value of
// type *UserType.
func marshalUserTypeToUserType(v *UserType) *UserType {
	if v == nil {
		return nil
	}
	res := &UserType{
		P: v.P,
	}

	return res
}
`

var ResultWithRecursiveResultTypeCode = `// RTView is a type which is projected based on a view.
type RTView struct {
	A *RTView
}

// RT is the viewed result type that projects RTView based on a view.
type RT struct {
	*RTView
	// View to render
	View string
}

// AsDefault projects viewed result type RT using the default view.
func (result *RT) AsDefault() *RT {
	t := &RTView{}
	if result.A != nil {
		t.A = result.A.AsTiny()
	}

	return &RT{
		RTView: t,
		View:   "default",
	}
}

// AsTiny projects viewed result type RT using the tiny view.
func (result *RT) AsTiny() *RT {
	t := &RTView{}
	if result.A != nil {
		t.A = result.A.AsDefault()
	}

	return &RT{
		RTView: t,
		View:   "tiny",
	}
}

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "default":
		if result.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
		}
		if result.A != nil {
			if err2 := result.A.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	case "tiny":
		if result.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
		}
		if result.A != nil {
			if err2 := result.A.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}
`
