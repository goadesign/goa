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
	// * "default"
	// * "tiny"
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

// NewMultipleViewsDefault projects result type MultipleViews into viewed
// result type MultipleViews using the default view.
func NewMultipleViewsDefault(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViews {
	p := newMultipleViewsViewDefault(res)
	return &multiplemethodsresultmultipleviewsviews.MultipleViews{p, "default"}
}

// NewMultipleViewsTiny projects result type MultipleViews into viewed result
// type MultipleViews using the tiny view.
func NewMultipleViewsTiny(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViews {
	p := newMultipleViewsViewTiny(res)
	return &multiplemethodsresultmultipleviewsviews.MultipleViews{p, "tiny"}
}

// newMultipleViewsViewDefault projects result type MultipleViews into
// projected type MultipleViewsView using the default view.
func newMultipleViewsViewDefault(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViewsView {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews into projected
// type MultipleViewsView using the tiny view.
func newMultipleViewsViewTiny(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViewsView {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const ResultCollectionMultipleViewsMethod = `
// Service is the ResultCollectionMultipleViewsMethod service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	// * "default"
	// * "tiny"
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
	res := make([]*MultipleViews, len(vres.Projected))
	for i, val := range vres.Projected {
		res[i] = &MultipleViews{
			A: *val.A,
			B: *val.B,
		}
	}
	return res
}

// NewMultipleViewsCollectionDefault projects result type
// MultipleViewsCollection into viewed result type MultipleViewsCollection
// using the default view.
func NewMultipleViewsCollectionDefault(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollection {
	p := newMultipleViewsCollectionViewDefault(res)
	return resultcollectionmultipleviewsmethodviews.MultipleViewsCollection{p, "default"}
}

// NewMultipleViewsCollectionTiny projects result type MultipleViewsCollection
// into viewed result type MultipleViewsCollection using the tiny view.
func NewMultipleViewsCollectionTiny(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollection {
	p := newMultipleViewsCollectionViewTiny(res)
	return resultcollectionmultipleviewsmethodviews.MultipleViewsCollection{p, "tiny"}
}

// newMultipleViewsCollectionViewDefault projects result type
// MultipleViewsCollection into projected type MultipleViewsCollectionView
// using the default view.
func newMultipleViewsCollectionViewDefault(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView, len(res))
	for i, n := range res {
		vres[i] = newMultipleViewsViewDefault(n)
	}
	return vres
}

// newMultipleViewsCollectionViewTiny projects result type
// MultipleViewsCollection into projected type MultipleViewsCollectionView
// using the tiny view.
func newMultipleViewsCollectionViewTiny(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView, len(res))
	for i, n := range res {
		vres[i] = newMultipleViewsViewTiny(n)
	}
	return vres
}

// newMultipleViewsViewDefault projects result type MultipleViews into
// projected type MultipleViewsView using the default view.
func newMultipleViewsViewDefault(res *MultipleViews) *resultcollectionmultipleviewsmethodviews.MultipleViewsView {
	vres := &resultcollectionmultipleviewsmethodviews.MultipleViewsView{
		A: &res.A,
		B: &res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews into projected
// type MultipleViewsView using the tiny view.
func newMultipleViewsViewTiny(res *MultipleViews) *resultcollectionmultipleviewsmethodviews.MultipleViewsView {
	vres := &resultcollectionmultipleviewsmethodviews.MultipleViewsView{
		A: &res.A,
	}
	return vres
}
`

const ResultWithOtherResultMethod = `
// Service is the ResultWithOtherResult service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	// * "default"
	// * "tiny"
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
	res := &MultipleViews{
		A: *vres.Projected.A,
	}
	res.B = unmarshalMultipleViews2ViewToMultipleViews2(vres.Projected.B)
	return res
}

// NewMultipleViewsDefault projects result type MultipleViews into viewed
// result type MultipleViews using the default view.
func NewMultipleViewsDefault(res *MultipleViews) *resultwithotherresultviews.MultipleViews {
	p := newMultipleViewsViewDefault(res)
	return &resultwithotherresultviews.MultipleViews{p, "default"}
}

// NewMultipleViewsTiny projects result type MultipleViews into viewed result
// type MultipleViews using the tiny view.
func NewMultipleViewsTiny(res *MultipleViews) *resultwithotherresultviews.MultipleViews {
	p := newMultipleViewsViewTiny(res)
	return &resultwithotherresultviews.MultipleViews{p, "tiny"}
}

// newMultipleViewsViewDefault projects result type MultipleViews into
// projected type MultipleViewsView using the default view.
func newMultipleViewsViewDefault(res *MultipleViews) *resultwithotherresultviews.MultipleViewsView {
	vres := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	if res.B != nil {
		vres.B = newMultipleViews2ViewDefault(res.B)
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews into projected
// type MultipleViewsView using the tiny view.
func newMultipleViewsViewTiny(res *MultipleViews) *resultwithotherresultviews.MultipleViewsView {
	vres := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	return vres
}

// newMultipleViews2ViewDefault projects result type MultipleViews2 into
// projected type MultipleViews2View using the default view.
func newMultipleViews2ViewDefault(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2View {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViews2ViewTiny projects result type MultipleViews2 into projected
// type MultipleViews2View using the tiny view.
func newMultipleViews2ViewTiny(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2View {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
	}
	return vres
}

// unmarshalMultipleViews2ViewToMultipleViews2 builds a value of type
// *MultipleViews2 from a value of type
// *resultwithotherresultviews.MultipleViews2View.
func unmarshalMultipleViews2ViewToMultipleViews2(v *resultwithotherresultviews.MultipleViews2View) *MultipleViews2 {
	res := &MultipleViews2{
		A: *v.A,
		B: v.B,
	}

	return res
}
`
