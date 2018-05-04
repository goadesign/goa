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
`

var ResultWithRecursiveResultTypeCode = `// RTView is a type which is projected based on a view.
type RTView struct {
	A *RT
}

// RT is the viewed result type that projects RTView based on a view.
type RT struct {
	*RTView
	// View to render
	View string
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
