package testdata

const ResultWithMultipleViewsCode = `// ResultTypeView is a type used by ResultType type to project based on a view.
type ResultTypeView struct {
	A *string
	B *string
}

// ResultType is the viewed result type that is projected based on a view.
type ResultType struct {
	// Type to project
	Projected *ResultTypeView
	// View to render
	View string
}

// NewResultType initializes ResultType viewed result type from ResultTypeView
// projected type and a view.
func NewResultType(p *ResultTypeView, view string) *ResultType {
	return &ResultType{
		Projected: p,
		View:      view,
	}
}

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "projected"))
		}
	default:
		if projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "projected"))
		}
		if projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "projected"))
		}
	}
	return
}
`

var ResultWithUserTypeCode = `// ResultTypeView is a type used by ResultType type to project based on a view.
type ResultTypeView struct {
	A *UserType
	B *string
}

// ResultType is the viewed result type that is projected based on a view.
type ResultType struct {
	// Type to project
	Projected *ResultTypeView
	// View to render
	View string
}

// UserType is a type that runs validations on a projected type.
type UserType struct {
	A *string
}

// NewResultType initializes ResultType viewed result type from ResultTypeView
// projected type and a view.
func NewResultType(p *ResultTypeView, view string) *ResultType {
	return &ResultType{
		Projected: p,
		View:      view,
	}
}

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "projected"))
		}
	default:
		if projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "projected"))
		}
	}
	return
}
`

const ResultWithResultTypeCode = `// RTView is a type used by RT type to project based on a view.
type RTView struct {
	A *string
	B *RT2
	C *RT3
}

// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RT2View is a type used by RT2 type to project based on a view.
type RT2View struct {
	C *string
	D *UserType
	E *string
}

// RT2 is the viewed result type that is projected based on a view.
type RT2 struct {
	// Type to project
	Projected *RT2View
	// View to render
	View string
}

// UserType is a type that runs validations on a projected type.
type UserType struct {
	P *string
}

// RT3View is a type used by RT3 type to project based on a view.
type RT3View struct {
	X []string
	Y map[int]*UserType
	Z *string
}

// RT3 is the viewed result type that is projected based on a view.
type RT3 struct {
	// Type to project
	Projected *RT3View
	// View to render
	View string
}

// NewRT initializes RT viewed result type from RTView projected type and a
// view.
func NewRT(p *RTView, view string) *RT {
	return &RT{
		Projected: p,
		View:      view,
	}
}

// NewRT2 initializes RT2 viewed result type from RT2View projected type and a
// view.
func NewRT2(p *RT2View, view string) *RT2 {
	return &RT2{
		Projected: p,
		View:      view,
	}
}

// NewRT3 initializes RT3 viewed result type from RT3View projected type and a
// view.
func NewRT3(p *RT3View, view string) *RT3 {
	return &RT3{
		Projected: p,
		View:      view,
	}
}

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "projected"))
		}
		if projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "projected"))
		}
		if projected.B != nil {
			if err2 := projected.B.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if projected.C != nil {
			if err2 := projected.C.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	default:
		if projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "projected"))
		}
		if projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "projected"))
		}
		if projected.B != nil {
			if err2 := projected.B.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if projected.C != nil {
			if err2 := projected.C.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// Validate runs the validations defined on RT2.
func (result *RT2) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "extended":
		if projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "projected"))
		}
		if projected.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "projected"))
		}
	case "tiny":
		if projected.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "projected"))
		}
	default:
		if projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "projected"))
		}
		if projected.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "projected"))
		}
	}
	return
}

// Validate runs the validations defined on RT3.
func (result *RT3) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.X == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("x", "projected"))
		}
	default:
		if projected.X == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("x", "projected"))
		}
		if projected.Y == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("y", "projected"))
		}
	}
	return
}
`

var ResultWithRecursiveResultTypeCode = `// RTView is a type used by RT type to project based on a view.
type RTView struct {
	A *RTView
}

// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// NewRT initializes RT viewed result type from RTView projected type and a
// view.
func NewRT(p *RTView, view string) *RT {
	return &RT{
		Projected: p,
		View:      view,
	}
}

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "projected"))
		}
		if projected.A != nil {
			if err2 := projected.A.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	default:
		if projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "projected"))
		}
		if projected.A != nil {
			if err2 := projected.A.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}
`
