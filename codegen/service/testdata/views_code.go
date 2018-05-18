package testdata

const ResultWithMultipleViewsCode = `// ResultTypeView is a type that runs validations on a projected type.
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

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "tiny":
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
	default:
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
		if result.Projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result.Projected"))
		}
	}
	return
}
`

const ResultCollectionMultipleViewsCode = `// ResultTypeCollection is the viewed result type that is projected based on a
// view.
type ResultTypeCollection []*ResultType

// ResultTypeView is a type that runs validations on a projected type.
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

// Validate runs the validations defined on ResultTypeCollection.
func (result ResultTypeCollection) Validate() (err error) {
	for _, projected := range result {
		if err2 := projected.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "tiny":
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
	default:
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
		if result.Projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result.Projected"))
		}
	}
	return
}
`

const ResultWithUserTypeCode = `// ResultTypeView is a type that runs validations on a projected type.
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

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "tiny":
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
	default:
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
	}
	return
}
`

const ResultWithResultTypeCode = `// RTView is a type that runs validations on a projected type.
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

// RT2View is a type that runs validations on a projected type.
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

// RT3View is a type that runs validations on a projected type.
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

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "tiny":
		if result.Projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result.Projected"))
		}
		if result.Projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result.Projected"))
		}
		if result.Projected.B != nil {
			if err2 := result.Projected.B.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if result.Projected.C != nil {
			if err2 := result.Projected.C.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	default:
		if result.Projected.B == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("b", "result.Projected"))
		}
		if result.Projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result.Projected"))
		}
		if result.Projected.B != nil {
			if err2 := result.Projected.B.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if result.Projected.C != nil {
			if err2 := result.Projected.C.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// Validate runs the validations defined on RT2.
func (result *RT2) Validate() (err error) {
	switch result.View {
	case "extended":
		if result.Projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result.Projected"))
		}
		if result.Projected.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "result.Projected"))
		}
	case "tiny":
		if result.Projected.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "result.Projected"))
		}
	default:
		if result.Projected.C == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("c", "result.Projected"))
		}
		if result.Projected.D == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("d", "result.Projected"))
		}
	}
	return
}

// Validate runs the validations defined on RT3.
func (result *RT3) Validate() (err error) {
	switch result.View {
	case "tiny":
		if result.Projected.X == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("x", "result.Projected"))
		}
	default:
		if result.Projected.X == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("x", "result.Projected"))
		}
		if result.Projected.Y == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("y", "result.Projected"))
		}
	}
	return
}
`

const ResultWithRecursiveResultTypeCode = `// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *RT
}

// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "tiny":
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
		if result.Projected.A != nil {
			if err2 := result.Projected.A.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	default:
		if result.Projected.A == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("a", "result.Projected"))
		}
		if result.Projected.A != nil {
			if err2 := result.Projected.A.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}
`
