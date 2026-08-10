package testdata

const ResultWithMultipleViewsCode = `// ResultType is the viewed result type that is projected based on a view.
type ResultType struct {
	// Type to project
	Projected *ResultTypeView
	// View to render
	View string
}

// ResultTypeView is a type that runs validations on a projected type.
type ResultTypeView struct {
	A *string
	B *string
}

var (
	// ResultTypeMap is a map indexing the attribute names of ResultType by view
	// name.
	ResultTypeMap = map[string][]string{
		"default": {
			"a",
			"b",
		},
		"tiny": {
			"a",
		},
	}
)

// ValidateResultType runs the validations defined on the viewed result type
// ResultType.
func ValidateResultType(result *ResultType) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultTypeView(result.Projected)
	case "tiny":
		err = ValidateResultTypeViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateResultTypeView runs the validations defined on ResultTypeView using
// the "default" view.
func ValidateResultTypeView(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateResultTypeViewTiny runs the validations defined on ResultTypeView
// using the "tiny" view.
func ValidateResultTypeViewTiny(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}
`

const ResultCollectionMultipleViewsCode = `// ResultTypeCollection is the viewed result type that is projected based on a
// view.
type ResultTypeCollection struct {
	// Type to project
	Projected ResultTypeCollectionView
	// View to render
	View string
}

// ResultTypeCollectionView is a type that runs validations on a projected type.
type ResultTypeCollectionView []*ResultTypeView

// ResultTypeView is a type that runs validations on a projected type.
type ResultTypeView struct {
	A *string
	B *string
}

var (
	// ResultTypeCollectionMap is a map indexing the attribute names of
	// ResultTypeCollection by view name.
	ResultTypeCollectionMap = map[string][]string{
		"default": {
			"a",
			"b",
		},
		"tiny": {
			"a",
		},
	}
	// ResultTypeMap is a map indexing the attribute names of ResultType by view
	// name.
	ResultTypeMap = map[string][]string{
		"default": {
			"a",
			"b",
		},
		"tiny": {
			"a",
		},
	}
)

// ValidateResultTypeCollection runs the validations defined on the viewed
// result type ResultTypeCollection.
func ValidateResultTypeCollection(result ResultTypeCollection) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultTypeCollectionView(result.Projected)
	case "tiny":
		err = ValidateResultTypeCollectionViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateResultTypeCollectionView runs the validations defined on
// ResultTypeCollectionView using the "default" view.
func ValidateResultTypeCollectionView(result ResultTypeCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateResultTypeView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateResultTypeCollectionViewTiny runs the validations defined on
// ResultTypeCollectionView using the "tiny" view.
func ValidateResultTypeCollectionViewTiny(result ResultTypeCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateResultTypeViewTiny(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateResultTypeView runs the validations defined on ResultTypeView using
// the "default" view.
func ValidateResultTypeView(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateResultTypeViewTiny runs the validations defined on ResultTypeView
// using the "tiny" view.
func ValidateResultTypeViewTiny(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}
`

const ResultWithUserTypeCode = `// ResultType is the viewed result type that is projected based on a view.
type ResultType struct {
	// Type to project
	Projected *ResultTypeView
	// View to render
	View string
}

// ResultTypeView is a type that runs validations on a projected type.
type ResultTypeView struct {
	A *UserTypeView
	B *string
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView struct {
	A *string
}

var (
	// ResultTypeMap is a map indexing the attribute names of ResultType by view
	// name.
	ResultTypeMap = map[string][]string{
		"default": {
			"a",
			"b",
		},
		"tiny": {
			"a",
		},
	}
)

// ValidateResultType runs the validations defined on the viewed result type
// ResultType.
func ValidateResultType(result *ResultType) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultTypeView(result.Projected)
	case "tiny":
		err = ValidateResultTypeViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateResultTypeView runs the validations defined on ResultTypeView using
// the "default" view.
func ValidateResultTypeView(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// ValidateResultTypeViewTiny runs the validations defined on ResultTypeView
// using the "tiny" view.
func ValidateResultTypeViewTiny(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// ValidateUserTypeView runs the validations defined on UserTypeView.
func ValidateUserTypeView(result *UserTypeView) (err error) {

	return
}
`

const ResultWithResultTypeCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *string
	B *RT2View
	C *RT3View
}

// RT2View is a type that runs validations on a projected type.
type RT2View struct {
	C *string
	D *UserTypeView
	E *string
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView struct {
	P *string
}

// RT3View is a type that runs validations on a projected type.
type RT3View struct {
	X []string
	Y map[int]*UserTypeView
	Z *string
}

var (
	// RTMap is a map indexing the attribute names of RT by view name.
	RTMap = map[string][]string{
		"default": {
			"a",
			"b",
			"c",
		},
		"tiny": {
			"b",
			"c",
		},
	}
	// RT2Map is a map indexing the attribute names of RT2 by view name.
	RT2Map = map[string][]string{
		"default": {
			"c",
			"d",
		},
		"extended": {
			"c",
			"d",
			"e",
		},
		"tiny": {
			"d",
		},
	}
	// RT3Map is a map indexing the attribute names of RT3 by view name.
	RT3Map = map[string][]string{
		"default": {
			"x",
			"y",
		},
		"tiny": {
			"x",
		},
	}
)

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	case "tiny":
		err = ValidateRTViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {

	if result.B != nil {
		if err2 := ValidateRT2ViewExtended(result.B); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := ValidateRT3View(result.C); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateRTViewTiny runs the validations defined on RTView using the "tiny"
// view.
func ValidateRTViewTiny(result *RTView) (err error) {

	if result.B != nil {
		if err2 := ValidateRT2ViewTiny(result.B); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := ValidateRT3View(result.C); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateRT2View runs the validations defined on RT2View using the "default"
// view.
func ValidateRT2View(result *RT2View) (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateRT2ViewExtended runs the validations defined on RT2View using the
// "extended" view.
func ValidateRT2ViewExtended(result *RT2View) (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateRT2ViewTiny runs the validations defined on RT2View using the "tiny"
// view.
func ValidateRT2ViewTiny(result *RT2View) (err error) {
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateUserTypeView runs the validations defined on UserTypeView.
func ValidateUserTypeView(result *UserTypeView) (err error) {

	return
}

// ValidateRT3View runs the validations defined on RT3View using the "default"
// view.
func ValidateRT3View(result *RT3View) (err error) {
	if result.X == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
	}
	if result.Y == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("y", "result"))
	}
	return
}

// ValidateRT3ViewTiny runs the validations defined on RT3View using the "tiny"
// view.
func ValidateRT3ViewTiny(result *RT3View) (err error) {
	if result.X == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
	}
	return
}
`

const ResultWithRecursiveResultTypeCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *RTView
}

var (
	// RTMap is a map indexing the attribute names of RT by view name.
	RTMap = map[string][]string{
		"default": {
			"a",
		},
		"tiny": {
			"a",
		},
	}
)

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	case "tiny":
		err = ValidateRTViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := ValidateRTViewTiny(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateRTViewTiny runs the validations defined on RTView using the "tiny"
// view.
func ValidateRTViewTiny(result *RTView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := ValidateRTView(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}
`

const ResultWithCustomFieldsCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	CustomA *string
	B       *int
}

var (
	// RTMap is a map indexing the attribute names of RT by view name.
	RTMap = map[string][]string{
		"default": {
			"a",
			"b",
		},
		"tiny": {
			"a",
		},
	}
)

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	case "tiny":
		err = ValidateRTViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {
	if result.CustomA == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateRTViewTiny runs the validations defined on RTView using the "tiny"
// view.
func ValidateRTViewTiny(result *RTView) (err error) {
	if result.CustomA == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}
`

const ResultWithRecursiveCollectionOfResultTypeCode = `// SomeRT is the viewed result type that is projected based on a view.
type SomeRT struct {
	// Type to project
	Projected *SomeRTView
	// View to render
	View string
}

// AnotherResult is the viewed result type that is projected based on a view.
type AnotherResult struct {
	// Type to project
	Projected *AnotherResultView
	// View to render
	View string
}

// SomeRTView is a type that runs validations on a projected type.
type SomeRTView struct {
	A SomeRTCollectionView
}

// SomeRTCollectionView is a type that runs validations on a projected type.
type SomeRTCollectionView []*SomeRTView

// AnotherResultView is a type that runs validations on a projected type.
type AnotherResultView struct {
	A AnotherResultCollectionView
}

// AnotherResultCollectionView is a type that runs validations on a projected
// type.
type AnotherResultCollectionView []*AnotherResultView

var (
	// SomeRTMap is a map indexing the attribute names of SomeRT by view name.
	SomeRTMap = map[string][]string{
		"default": {
			"a",
		},
		"tiny": {
			"a",
		},
	}
	// AnotherResultMap is a map indexing the attribute names of AnotherResult by
	// view name.
	AnotherResultMap = map[string][]string{
		"default": {
			"a",
		},
	}
	// SomeRTCollectionMap is a map indexing the attribute names of
	// SomeRTCollection by view name.
	SomeRTCollectionMap = map[string][]string{
		"default": {
			"a",
		},
		"tiny": {
			"a",
		},
	}
	// AnotherResultCollectionMap is a map indexing the attribute names of
	// AnotherResultCollection by view name.
	AnotherResultCollectionMap = map[string][]string{
		"default": {
			"a",
		},
	}
)

// ValidateSomeRT runs the validations defined on the viewed result type SomeRT.
func ValidateSomeRT(result *SomeRT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateSomeRTView(result.Projected)
	case "tiny":
		err = ValidateSomeRTViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default", "tiny"})
	}
	return
}

// ValidateAnotherResult runs the validations defined on the viewed result type
// AnotherResult.
func ValidateAnotherResult(result *AnotherResult) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateAnotherResultView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateSomeRTView runs the validations defined on SomeRTView using the
// "default" view.
func ValidateSomeRTView(result *SomeRTView) (err error) {

	if result.A != nil {
		if err2 := ValidateSomeRTCollectionViewTiny(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateSomeRTViewTiny runs the validations defined on SomeRTView using the
// "tiny" view.
func ValidateSomeRTViewTiny(result *SomeRTView) (err error) {

	if result.A != nil {
		if err2 := ValidateSomeRTCollectionView(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateSomeRTCollectionView runs the validations defined on
// SomeRTCollectionView using the "default" view.
func ValidateSomeRTCollectionView(result SomeRTCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateSomeRTView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateSomeRTCollectionViewTiny runs the validations defined on
// SomeRTCollectionView using the "tiny" view.
func ValidateSomeRTCollectionViewTiny(result SomeRTCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateSomeRTViewTiny(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAnotherResultView runs the validations defined on AnotherResultView
// using the "default" view.
func ValidateAnotherResultView(result *AnotherResultView) (err error) {

	if result.A != nil {
		if err2 := ValidateAnotherResultCollectionView(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAnotherResultCollectionView runs the validations defined on
// AnotherResultCollectionView using the "default" view.
func ValidateAnotherResultCollectionView(result AnotherResultCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateAnotherResultView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}
`

const ResultWithMultipleMethodsCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *string
}

var (
	// RTMap is a map indexing the attribute names of RT by view name.
	RTMap = map[string][]string{
		"default": {
			"a",
		},
	}
)

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {

	return
}
`

const ResultWithEnumType = `// Result is the viewed result type that is projected based on a view.
type Result struct {
	// Type to project
	Projected *ResultView
	// View to render
	View string
}

// ResultView is a type that runs validations on a projected type.
type ResultView struct {
	T []UserTypeView
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView string

var (
	// ResultMap is a map indexing the attribute names of Result by view name.
	ResultMap = map[string][]string{
		"default": {
			"t",
		},
	}
)

// ValidateResult runs the validations defined on the viewed result type Result.
func ValidateResult(result *Result) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateResultView runs the validations defined on ResultView using the
// "default" view.
func ValidateResultView(result *ResultView) (err error) {
	for _, e := range result.T {
		if !(string(e) == "a" || string(e) == "b") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("result.t[*]", string(e), []any{"a", "b"}))
		}
	}
	return
}

// ValidateUserTypeView runs the validations defined on UserTypeView.
func ValidateUserTypeView(result UserTypeView) (err error) {
	if !(string(result) == "a" || string(result) == "b") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("result", string(result), []any{"a", "b"}))
	}
	return
}
`

const ResultWithPkgPathCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *UserTypeView
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView struct {
	A *string
}

var (
	// RTMap is a map indexing the attribute names of RT by view name.
	RTMap = map[string][]string{
		"default": {
			"a",
		},
	}
)

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {

	return
}

// ValidateUserTypeView runs the validations defined on UserTypeView.
func ValidateUserTypeView(result *UserTypeView) (err error) {

	return
}
`

const ResultWithOneOfInResultTypeCode = "// OneOfResource is the viewed result type that is projected based on a view.\ntype OneOfResource struct {\n\t// Type to project\n\tProjected *OneOfResourceView\n\t// View to render\n\tView string\n}\n\n// OneOfResourceView is a type that runs validations on a projected type.\ntype OneOfResourceView struct {\n\t// Data (type depends on flag)\n\tData *OneOfValueView\n}\n\n// OneOfValueView is a type that runs validations on a projected type.\ntype OneOfValueView struct {\n\tFlag *Flag\n}\n\n// FlagAstringView is a type that runs validations on a projected type.\ntype FlagAstringView string\n\n// FlagAintView is a type that runs validations on a projected type.\ntype FlagAintView int64\n\n// Flag is a sum-type union.\ntype Flag struct {\n\tkind    FlagKind\n\tAstring FlagAstringView\n\tAint    FlagAintView\n}\n\n// FlagKind enumerates the union variants for Flag.\ntype FlagKind string\n\nconst (\n\t// FlagKindAstring identifies the astring branch of the union.\n\tFlagKindAstring FlagKind = \"astring\"\n\t// FlagKindAint identifies the aint branch of the union.\n\tFlagKindAint FlagKind = \"aint\"\n)\n\n// Kind returns the discriminator value of the union.\nfunc (u Flag) Kind() FlagKind {\n\treturn u.kind\n}\n\n// NewFlagAstring constructs Flag with the astring branch set.\nfunc NewFlagAstring(v FlagAstringView) Flag {\n\treturn Flag{\n\t\tkind:    FlagKindAstring,\n\t\tAstring: v,\n\t}\n}\n\n// AsAstring returns the value of the astring branch if set.\nfunc (u Flag) AsAstring() (_ FlagAstringView, ok bool) {\n\tif u.kind != FlagKindAstring {\n\t\treturn\n\t}\n\treturn u.Astring, true\n}\n\n// SetAstring sets the astring branch of the union.\nfunc (u *Flag) SetAstring(v FlagAstringView) {\n\tu.kind = FlagKindAstring\n\tu.Astring = v\n}\n\n// NewFlagAint constructs Flag with the aint branch set.\nfunc NewFlagAint(v FlagAintView) Flag {\n\treturn Flag{\n\t\tkind: FlagKindAint,\n\t\tAint: v,\n\t}\n}\n\n// AsAint returns the value of the aint branch if set.\nfunc (u Flag) AsAint() (_ FlagAintView, ok bool) {\n\tif u.kind != FlagKindAint {\n\t\treturn\n\t}\n\treturn u.Aint, true\n}\n\n// SetAint sets the aint branch of the union.\nfunc (u *Flag) SetAint(v FlagAintView) {\n\tu.kind = FlagKindAint\n\tu.Aint = v\n}\n\n// Validate ensures the union discriminant is valid.\nfunc (u Flag) Validate() error {\n\tswitch u.kind {\n\tcase \"\":\n\t\treturn goa.InvalidEnumValueError(\"type\", \"\", []any{\n\t\t\tstring(FlagKindAstring),\n\t\t\tstring(FlagKindAint),\n\t\t})\n\tcase FlagKindAstring:\n\t\treturn nil\n\tcase FlagKindAint:\n\t\treturn nil\n\tdefault:\n\t\treturn goa.InvalidEnumValueError(\"type\", u.kind, []any{\n\t\t\tstring(FlagKindAstring),\n\t\t\tstring(FlagKindAint),\n\t\t})\n\t}\n}\n\n// MarshalJSON marshals the union into the canonical {type,value} JSON shape.\nfunc (u Flag) MarshalJSON() ([]byte, error) {\n\tif err := u.Validate(); err != nil {\n\t\treturn nil, err\n\t}\n\tvar (\n\t\tvalue any\n\t)\n\tswitch u.kind {\n\tcase FlagKindAstring:\n\t\tvalue = u.Astring\n\tcase FlagKindAint:\n\t\tvalue = u.Aint\n\tdefault:\n\t\treturn nil, fmt.Errorf(\"unexpected Flag discriminant %q\", u.kind)\n\t}\n\treturn json.Marshal(struct {\n\t\tType  string `json:\"type\"`\n\t\tValue any    `json:\"value\"`\n\t}{\n\t\tType:  string(u.kind),\n\t\tValue: value,\n\t})\n}\n\n// UnmarshalJSON unmarshals the union from the canonical {type,value} JSON shape.\nfunc (u *Flag) UnmarshalJSON(data []byte) error {\n\tvar raw struct {\n\t\tType  string          `json:\"type\"`\n\t\tValue json.RawMessage `json:\"value\"`\n\t}\n\tif err := json.Unmarshal(data, &raw); err != nil {\n\t\treturn err\n\t}\n\tif len(raw.Value) == 0 {\n\t\treturn goa.MissingFieldError(\"value\", \"Flag\")\n\t}\n\tif bytes.Equal(bytes.TrimSpace(raw.Value), []byte(\"null\")) {\n\t\treturn goa.InvalidFieldTypeError(\"value\", nil, \"non-null JSON value\")\n\t}\n\tswitch raw.Type {\n\tcase string(FlagKindAstring):\n\t\tvar v FlagAstringView\n\t\tif err := json.Unmarshal(raw.Value, &v); err != nil {\n\t\t\treturn err\n\t\t}\n\t\tu.kind = FlagKindAstring\n\t\tu.Astring = v\n\tcase string(FlagKindAint):\n\t\tvar v FlagAintView\n\t\tif err := json.Unmarshal(raw.Value, &v); err != nil {\n\t\t\treturn err\n\t\t}\n\t\tu.kind = FlagKindAint\n\t\tu.Aint = v\n\tdefault:\n\t\tif raw.Type == \"\" {\n\t\t\treturn goa.MissingFieldError(\"type\", \"Flag\")\n\t\t}\n\t\treturn goa.InvalidEnumValueError(\"type\", raw.Type, []any{\n\t\t\tstring(FlagKindAstring),\n\t\t\tstring(FlagKindAint),\n\t\t})\n\t}\n\treturn nil\n}\n\nvar (\n\t// OneOfResourceMap is a map indexing the attribute names of OneOfResource by\n\t// view name.\n\tOneOfResourceMap = map[string][]string{\n\t\t\"default\": {\n\t\t\t\"data\",\n\t\t},\n\t}\n)\n\n// ValidateOneOfResource runs the validations defined on the viewed result type\n// OneOfResource.\nfunc ValidateOneOfResource(result *OneOfResource) (err error) {\n\tswitch result.View {\n\tcase \"default\", \"\":\n\t\terr = ValidateOneOfResourceView(result.Projected)\n\tdefault:\n\t\terr = goa.InvalidEnumValueError(\"view\", result.View, []any{\"default\"})\n\t}\n\treturn\n}\n\n// ValidateOneOfResourceView runs the validations defined on OneOfResourceView\n// using the \"default\" view.\nfunc ValidateOneOfResourceView(result *OneOfResourceView) (err error) {\n\tif result.Data == nil {\n\t\terr = goa.MergeErrors(err, goa.MissingFieldError(\"data\", \"result\"))\n\t}\n\treturn\n}\n\n// ValidateOneOfValueView runs the validations defined on OneOfValueView.\nfunc ValidateOneOfValueView(result *OneOfValueView) (err error) {\n\n\treturn\n}\n\n// ValidateFlagAstringView runs the validations defined on FlagAstringView.\nfunc ValidateFlagAstringView(result FlagAstringView) (err error) {\n\n\treturn\n}\n\n// ValidateFlagAintView runs the validations defined on FlagAintView.\nfunc ValidateFlagAintView(result FlagAintView) (err error) {\n\n\treturn\n}\n"
