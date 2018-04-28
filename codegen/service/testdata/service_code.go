package testdata

const SingleMethod = `
// Service is the SingleMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *APayload) (*AResult, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "SingleMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// APayload is the payload type of the SingleMethod service A method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the SingleMethod service A method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

const MultipleMethods = `
// Service is the MultipleMethods service interface.
type Service interface {
	// A implements A.
	A(context.Context, *APayload) (*AResult, error)
	// B implements B.
	B(context.Context, *BPayload) (*BResult, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "MultipleMethods"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "B"}

// APayload is the payload type of the MultipleMethods service A method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the MultipleMethods service A method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// BPayload is the payload type of the MultipleMethods service B method.
type BPayload struct {
	ArrayField  []bool
	MapField    map[int]string
	ObjectField *struct {
		IntField    *int
		StringField *string
	}
	UserTypeField *Parent
}

// BResult is the result type of the MultipleMethods service B method.
type BResult struct {
	ArrayField  []bool
	MapField    map[int]string
	ObjectField *struct {
		IntField    *int
		StringField *string
	}
	UserTypeField *Parent
}

type Parent struct {
	C *Child
}

type Child struct {
	P *Parent
}
`

const EmptyMethod = `
// Service is the Empty service interface.
type Service interface {
	// Empty implements Empty.
	Empty(context.Context) error
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "Empty"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"Empty"}
`

const EmptyResultMethod = `
// Service is the EmptyResult service interface.
type Service interface {
	// EmptyResult implements EmptyResult.
	EmptyResult(context.Context, *APayload) error
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "EmptyResult"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"EmptyResult"}

// APayload is the payload type of the EmptyResult service EmptyResult method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

const EmptyPayloadMethod = `
// Service is the EmptyPayload service interface.
type Service interface {
	// EmptyPayload implements EmptyPayload.
	EmptyPayload(context.Context) (*AResult, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "EmptyPayload"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"EmptyPayload"}

// AResult is the result type of the EmptyPayload service EmptyPayload method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

const ServiceError = `
// Service is the ServiceError service interface.
type Service interface {
	// A implements A.
	A(context.Context) error
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ServiceError"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// MakeError builds a goa.ServiceError from an error.
func MakeError(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:    "error",
		ID:      goa.NewErrorID(),
		Message: err.Error(),
	}
}
`

const MultipleMethodsResultMultipleViews = `
// Service is the MultipleMethodsResultMultipleViews service interface.
type Service interface {
	// A implements A.
	// It must return one of the following views
	// * default
	// * tiny
	A(context.Context, *APayload) (*MultipleViews, string, error)
	// B implements B.
	B(context.Context) (*SingleView, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "MultipleMethodsResultMultipleViews"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "B"}

// APayload is the payload type of the MultipleMethodsResultMultipleViews
// service A method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// MultipleViews is the result type of the MultipleMethodsResultMultipleViews
// service A method.
type MultipleViews struct {
	A *string
	B *string
}

// SingleView is the result type of the MultipleMethodsResultMultipleViews
// service B method.
type SingleView struct {
	A *string
	B *string
}

// NewMultipleViews converts viewed result type MultipleViews to result type
// MultipleViews.
func NewMultipleViews(vRes *multiplemethodsresultmultipleviewsviews.MultipleViews) *MultipleViews {
	res := &MultipleViews{
		A: vRes.A,
		B: vRes.B,
	}
	return res
}

// NewViewedMultipleViews converts result type MultipleViews to viewed result
// type MultipleViews.
func NewViewedMultipleViews(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViews {
	v := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return &multiplemethodsresultmultipleviewsviews.MultipleViews{MultipleViewsView: v}
}
`

const ResultWithOtherResultMethod = `
// Service is the ResultWithOtherResult service interface.
type Service interface {
	// A implements A.
	// It must return one of the following views
	// * default
	// * tiny
	A(context.Context) (*MultipleViews, string, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ResultWithOtherResult"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// MultipleViews is the result type of the ResultWithOtherResult service A
// method.
type MultipleViews struct {
	A string
	B *MultipleViews2
}

type MultipleViews2 struct {
	A string
	B *string
}

// NewMultipleViews2 converts viewed result type MultipleViews2 to result type
// MultipleViews2.
func NewMultipleViews2(vRes *resultwithotherresultviews.MultipleViews2) *MultipleViews2 {
	res := &MultipleViews2{
		B: vRes.B,
	}
	if vRes.A != nil {
		res.A = *vRes.A
	}
	return res
}

// NewViewedMultipleViews2 converts result type MultipleViews2 to viewed result
// type MultipleViews2.
func NewViewedMultipleViews2(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2 {
	v := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
		B: res.B,
	}
	return &resultwithotherresultviews.MultipleViews2{MultipleViews2View: v}
}

// NewMultipleViews converts viewed result type MultipleViews to result type
// MultipleViews.
func NewMultipleViews(vRes *resultwithotherresultviews.MultipleViews) *MultipleViews {
	res := &MultipleViews{}
	if vRes.A != nil {
		res.A = *vRes.A
	}
	if vRes.B != nil {
		res.B = NewMultipleViews2(vRes.B)
	}

	return res
}

// NewViewedMultipleViews converts result type MultipleViews to viewed result
// type MultipleViews.
func NewViewedMultipleViews(res *MultipleViews) *resultwithotherresultviews.MultipleViews {
	v := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	if res.B != nil {
		v.B = NewViewedMultipleViews2(res.B)
	}

	return &resultwithotherresultviews.MultipleViews{MultipleViewsView: v}
}
`
