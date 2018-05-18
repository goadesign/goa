package testdata

const SingleMethod = `
// Service is the SingleMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *APayload) (res *AResult, err error)
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
	A(context.Context, *APayload) (res *AResult, err error)
	// B implements B.
	B(context.Context, *BPayload) (res *BResult, err error)
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
	Empty(context.Context) (err error)
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
	EmptyResult(context.Context, *APayload) (err error)
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
	EmptyPayload(context.Context) (res *AResult, err error)
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
	A(context.Context) (err error)
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
	// The "view" return value must have one of the following views
	// * "tiny"
	// * "default"
	A(context.Context, *APayload) (res *MultipleViews, view string, err error)
	// B implements B.
	B(context.Context) (res *SingleView, err error)
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
func NewMultipleViews(vres *multiplemethodsresultmultipleviewsviews.MultipleViews) *MultipleViews {
	res := &MultipleViews{
		A: vres.Projected.A,
		B: vres.Projected.B,
	}
	return res
}

// NewMultipleViewsTiny projects result type MultipleViews into viewed result
// type MultipleViews using the tiny view.
func NewMultipleViewsTiny(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViews {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
	}
	return &multiplemethodsresultmultipleviewsviews.MultipleViews{vres, "tiny"}
}

// NewMultipleViewsDefault projects result type MultipleViews into viewed
// result type MultipleViews using the default view.
func NewMultipleViewsDefault(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViews {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return &multiplemethodsresultmultipleviewsviews.MultipleViews{vres, "default"}
}
`

const ResultCollectionMultipleViewsMethod = `
// Service is the ResultCollectionMultipleViewsMethod service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	// * "tiny"
	// * "default"
	A(context.Context) (res MultipleViewsCollection, view string, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ResultCollectionMultipleViewsMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// MultipleViewsCollection is the result type of the
// ResultCollectionMultipleViewsMethod service A method.
type MultipleViewsCollection []*MultipleViews

type MultipleViews struct {
	A string
	B int
}

// NewMultipleViewsCollection converts viewed result type
// MultipleViewsCollection to result type MultipleViewsCollection.
func NewMultipleViewsCollection(vres resultcollectionmultipleviewsmethodviews.MultipleViewsCollection) MultipleViewsCollection {
	res := make(MultipleViewsCollection, len(vres))
	for i, n := range vres {
		res[i] = NewMultipleViews(n)
	}
	return res
}

// NewMultipleViewsCollectionTiny projects result type MultipleViewsCollection
// into viewed result type MultipleViewsCollection using the tiny view.
func NewMultipleViewsCollectionTiny(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollection {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollection, len(res))
	for i, n := range res {
		vres[i] = NewMultipleViewsTiny(n)
	}
	return vres
}

// NewMultipleViewsCollectionDefault projects result type
// MultipleViewsCollection into viewed result type MultipleViewsCollection
// using the default view.
func NewMultipleViewsCollectionDefault(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollection {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollection, len(res))
	for i, n := range res {
		vres[i] = NewMultipleViewsDefault(n)
	}
	return vres
}

// NewMultipleViews converts viewed result type MultipleViews to result type
// MultipleViews.
func NewMultipleViews(vres *resultcollectionmultipleviewsmethodviews.MultipleViews) *MultipleViews {
	res := &MultipleViews{}
	if vres.Projected.A != nil {
		res.A = *vres.Projected.A
	}
	if vres.Projected.B != nil {
		res.B = *vres.Projected.B
	}
	return res
}

// NewMultipleViewsTiny projects result type MultipleViews into viewed result
// type MultipleViews using the tiny view.
func NewMultipleViewsTiny(res *MultipleViews) *resultcollectionmultipleviewsmethodviews.MultipleViews {
	vres := &resultcollectionmultipleviewsmethodviews.MultipleViewsView{
		A: &res.A,
	}
	return &resultcollectionmultipleviewsmethodviews.MultipleViews{vres, "tiny"}
}

// NewMultipleViewsDefault projects result type MultipleViews into viewed
// result type MultipleViews using the default view.
func NewMultipleViewsDefault(res *MultipleViews) *resultcollectionmultipleviewsmethodviews.MultipleViews {
	vres := &resultcollectionmultipleviewsmethodviews.MultipleViewsView{
		A: &res.A,
		B: &res.B,
	}
	return &resultcollectionmultipleviewsmethodviews.MultipleViews{vres, "default"}
}
`

const ResultWithOtherResultMethod = `
// Service is the ResultWithOtherResult service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	// * "tiny"
	// * "default"
	A(context.Context) (res *MultipleViews, view string, err error)
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

// NewMultipleViews converts viewed result type MultipleViews to result type
// MultipleViews.
func NewMultipleViews(vres *resultwithotherresultviews.MultipleViews) *MultipleViews {
	res := &MultipleViews{}
	if vres.Projected.A != nil {
		res.A = *vres.Projected.A
	}
	if vres.Projected.B != nil {
		res.B = NewMultipleViews2(vres.Projected.B)
	}
	return res
}

// NewMultipleViewsTiny projects result type MultipleViews into viewed result
// type MultipleViews using the tiny view.
func NewMultipleViewsTiny(res *MultipleViews) *resultwithotherresultviews.MultipleViews {
	vres := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	return &resultwithotherresultviews.MultipleViews{vres, "tiny"}
}

// NewMultipleViewsDefault projects result type MultipleViews into viewed
// result type MultipleViews using the default view.
func NewMultipleViewsDefault(res *MultipleViews) *resultwithotherresultviews.MultipleViews {
	vres := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	if res.B != nil {
		vres.B = NewMultipleViews2Default(res.B)
	}
	return &resultwithotherresultviews.MultipleViews{vres, "default"}
}

// NewMultipleViews2 converts viewed result type MultipleViews2 to result type
// MultipleViews2.
func NewMultipleViews2(vres *resultwithotherresultviews.MultipleViews2) *MultipleViews2 {
	res := &MultipleViews2{
		B: vres.Projected.B,
	}
	if vres.Projected.A != nil {
		res.A = *vres.Projected.A
	}
	return res
}

// NewMultipleViews2Tiny projects result type MultipleViews2 into viewed result
// type MultipleViews2 using the tiny view.
func NewMultipleViews2Tiny(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2 {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
	}
	return &resultwithotherresultviews.MultipleViews2{vres, "tiny"}
}

// NewMultipleViews2Default projects result type MultipleViews2 into viewed
// result type MultipleViews2 using the default view.
func NewMultipleViews2Default(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2 {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
		B: res.B,
	}
	return &resultwithotherresultviews.MultipleViews2{vres, "default"}
}
`
