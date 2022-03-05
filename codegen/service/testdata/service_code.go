package testdata

const NamesWithSpaces = `
// Service is the Service With Spaces service interface.
type Service interface {
	// MethodWithSpaces implements Method With Spaces.
	MethodWithSpaces(context.Context, *PayloadWithSpace) (res *ResultWithSpace, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "Service With Spaces"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"Method With Spaces"}

// PayloadWithSpace is the payload type of the Service With Spaces service
// Method With Spaces method.
type PayloadWithSpace struct {
	String *string
}

// ResultWithSpace is the result type of the Service With Spaces service Method
// With Spaces method.
type ResultWithSpace struct {
	Int *int
}

// NewResultWithSpace initializes result type ResultWithSpace from viewed
// result type ResultWithSpace.
func NewResultWithSpace(vres *servicewithspacesviews.ResultWithSpace) *ResultWithSpace {
	return newResultWithSpace(vres.Projected)
}

// NewViewedResultWithSpace initializes viewed result type ResultWithSpace from
// result type ResultWithSpace using the given view.
func NewViewedResultWithSpace(res *ResultWithSpace, view string) *servicewithspacesviews.ResultWithSpace {
	p := newResultWithSpaceView(res)
	return &servicewithspacesviews.ResultWithSpace{Projected: p, View: "default"}
}

// newResultWithSpace converts projected type ResultWithSpace to service type
// ResultWithSpace.
func newResultWithSpace(vres *servicewithspacesviews.ResultWithSpaceView) *ResultWithSpace {
	res := &ResultWithSpace{
		Int: vres.Int,
	}
	return res
}

// newResultWithSpaceView projects result type ResultWithSpace to projected
// type ResultWithSpaceView using the "default" view.
func newResultWithSpaceView(res *ResultWithSpace) *servicewithspacesviews.ResultWithSpaceView {
	vres := &servicewithspacesviews.ResultWithSpaceView{
		Int: res.Int,
	}
	return vres
}
`

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

type Child struct {
	P *Parent
}

type Parent struct {
	C *Child
}
`

const WithDefault = `
// Service is the WithDefault service interface.
type Service interface {
	// A implements A.
	A(context.Context, *APayload) (res *AResult, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "WithDefault"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// APayload is the payload type of the WithDefault service A method.
type APayload struct {
	IntField      int
	StringField   string
	OptionalField *string
	RequiredField float32
}

// AResult is the result type of the WithDefault service A method.
type AResult struct {
	IntField      int
	StringField   string
	OptionalField *string
	RequiredField float32
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

const CustomErrors = `
// Service is the CustomErrors service interface.
type Service interface {
	// A implements A.
	A(context.Context) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "CustomErrors"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

type Result struct {
	A *string
	B string
}

// primitive error description
type Primitive string

// Error returns an error description.
func (e *APayload) Error() string {
	return ""
}

// ErrorName returns "APayload".
func (e *APayload) ErrorName() string {
	return "user_type"
}

// Error returns an error description.
func (e *Result) Error() string {
	return ""
}

// ErrorName returns "Result".
func (e *Result) ErrorName() string {
	return e.B
}

// Error returns an error description.
func (e Primitive) Error() string {
	return "primitive error description"
}

// ErrorName returns "primitive".
func (e Primitive) ErrorName() string {
	return "primitive"
}
`

const CustomErrorsCustomField = `
// Service is the CustomErrorsCustomFields service interface.
type Service interface {
	// A implements A.
	A(context.Context) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "CustomErrorsCustomFields"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

type GoaError struct {
	ErrorCode string
}

// Error returns an error description.
func (e *GoaError) Error() string {
	return ""
}

// ErrorName returns "GoaError".
func (e *GoaError) ErrorName() string {
	return e.ErrorCode
}
`

const MultipleMethodsResultMultipleViews = `
// Service is the MultipleMethodsResultMultipleViews service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "tiny"
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

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *multiplemethodsresultmultipleviewsviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *multiplemethodsresultmultipleviewsviews.MultipleViews {
	var vres *multiplemethodsresultmultipleviewsviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &multiplemethodsresultmultipleviewsviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &multiplemethodsresultmultipleviewsviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// NewSingleView initializes result type SingleView from viewed result type
// SingleView.
func NewSingleView(vres *multiplemethodsresultmultipleviewsviews.SingleView) *SingleView {
	return newSingleView(vres.Projected)
}

// NewViewedSingleView initializes viewed result type SingleView from result
// type SingleView using the given view.
func NewViewedSingleView(res *SingleView, view string) *multiplemethodsresultmultipleviewsviews.SingleView {
	p := newSingleViewView(res)
	return &multiplemethodsresultmultipleviewsviews.SingleView{Projected: p, View: "default"}
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *multiplemethodsresultmultipleviewsviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *multiplemethodsresultmultipleviewsviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViewsView {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViewsView {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}

// newSingleView converts projected type SingleView to service type SingleView.
func newSingleView(vres *multiplemethodsresultmultipleviewsviews.SingleViewView) *SingleView {
	res := &SingleView{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newSingleViewView projects result type SingleView to projected type
// SingleViewView using the "default" view.
func newSingleViewView(res *SingleView) *multiplemethodsresultmultipleviewsviews.SingleViewView {
	vres := &multiplemethodsresultmultipleviewsviews.SingleViewView{
		A: res.A,
		B: res.B,
	}
	return vres
}
`

const WithExplicitAndDefaultViews = `
// Service is the WithExplicitAndDefaultViews service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "tiny"
	A(context.Context) (res *MultipleViews, view string, err error)
	// A implements A.
	AEndpoint(context.Context) (res *MultipleViews, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "WithExplicitAndDefaultViews"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "A"}

// MultipleViews is the result type of the WithExplicitAndDefaultViews service
// A method.
type MultipleViews struct {
	A string
	B int
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *withexplicitanddefaultviewsviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *withexplicitanddefaultviewsviews.MultipleViews {
	var vres *withexplicitanddefaultviewsviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &withexplicitanddefaultviewsviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &withexplicitanddefaultviewsviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *withexplicitanddefaultviewsviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{}
	if vres.A != nil {
		res.A = *vres.A
	}
	if vres.B != nil {
		res.B = *vres.B
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *withexplicitanddefaultviewsviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{}
	if vres.A != nil {
		res.A = *vres.A
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *withexplicitanddefaultviewsviews.MultipleViewsView {
	vres := &withexplicitanddefaultviewsviews.MultipleViewsView{
		A: &res.A,
		B: &res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *withexplicitanddefaultviewsviews.MultipleViewsView {
	vres := &withexplicitanddefaultviewsviews.MultipleViewsView{
		A: &res.A,
	}
	return vres
}
`

const ResultCollectionMultipleViewsMethod = `
// Service is the ResultCollectionMultipleViewsMethod service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "tiny"
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

type MultipleViews struct {
	A string
	B int
}

// MultipleViewsCollection is the result type of the
// ResultCollectionMultipleViewsMethod service A method.
type MultipleViewsCollection []*MultipleViews

// NewMultipleViewsCollection initializes result type MultipleViewsCollection
// from viewed result type MultipleViewsCollection.
func NewMultipleViewsCollection(vres resultcollectionmultipleviewsmethodviews.MultipleViewsCollection) MultipleViewsCollection {
	var res MultipleViewsCollection
	switch vres.View {
	case "default", "":
		res = newMultipleViewsCollection(vres.Projected)
	case "tiny":
		res = newMultipleViewsCollectionTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViewsCollection initializes viewed result type
// MultipleViewsCollection from result type MultipleViewsCollection using the
// given view.
func NewViewedMultipleViewsCollection(res MultipleViewsCollection, view string) resultcollectionmultipleviewsmethodviews.MultipleViewsCollection {
	var vres resultcollectionmultipleviewsmethodviews.MultipleViewsCollection
	switch view {
	case "default", "":
		p := newMultipleViewsCollectionView(res)
		vres = resultcollectionmultipleviewsmethodviews.MultipleViewsCollection{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsCollectionViewTiny(res)
		vres = resultcollectionmultipleviewsmethodviews.MultipleViewsCollection{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViewsCollection converts projected type MultipleViewsCollection
// to service type MultipleViewsCollection.
func newMultipleViewsCollection(vres resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView) MultipleViewsCollection {
	res := make(MultipleViewsCollection, len(vres))
	for i, n := range vres {
		res[i] = newMultipleViews(n)
	}
	return res
}

// newMultipleViewsCollectionTiny converts projected type
// MultipleViewsCollection to service type MultipleViewsCollection.
func newMultipleViewsCollectionTiny(vres resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView) MultipleViewsCollection {
	res := make(MultipleViewsCollection, len(vres))
	for i, n := range vres {
		res[i] = newMultipleViewsTiny(n)
	}
	return res
}

// newMultipleViewsCollectionView projects result type MultipleViewsCollection
// to projected type MultipleViewsCollectionView using the "default" view.
func newMultipleViewsCollectionView(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView, len(res))
	for i, n := range res {
		vres[i] = newMultipleViewsView(n)
	}
	return vres
}

// newMultipleViewsCollectionViewTiny projects result type
// MultipleViewsCollection to projected type MultipleViewsCollectionView using
// the "tiny" view.
func newMultipleViewsCollectionViewTiny(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView, len(res))
	for i, n := range res {
		vres[i] = newMultipleViewsViewTiny(n)
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *resultcollectionmultipleviewsmethodviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{}
	if vres.A != nil {
		res.A = *vres.A
	}
	if vres.B != nil {
		res.B = *vres.B
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *resultcollectionmultipleviewsmethodviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{}
	if vres.A != nil {
		res.A = *vres.A
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *resultcollectionmultipleviewsmethodviews.MultipleViewsView {
	vres := &resultcollectionmultipleviewsmethodviews.MultipleViewsView{
		A: &res.A,
		B: &res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
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
	//	- "default"
	//	- "tiny"
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

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *resultwithotherresultviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *resultwithotherresultviews.MultipleViews {
	var vres *resultwithotherresultviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &resultwithotherresultviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &resultwithotherresultviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *resultwithotherresultviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{}
	if vres.A != nil {
		res.A = *vres.A
	}
	if vres.B != nil {
		res.B = newMultipleViews2(vres.B)
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *resultwithotherresultviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{}
	if vres.A != nil {
		res.A = *vres.A
	}
	if vres.B != nil {
		res.B = newMultipleViews2(vres.B)
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *resultwithotherresultviews.MultipleViewsView {
	vres := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	if res.B != nil {
		vres.B = newMultipleViews2View(res.B)
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *resultwithotherresultviews.MultipleViewsView {
	vres := &resultwithotherresultviews.MultipleViewsView{
		A: &res.A,
	}
	return vres
}

// newMultipleViews2 converts projected type MultipleViews2 to service type
// MultipleViews2.
func newMultipleViews2(vres *resultwithotherresultviews.MultipleViews2View) *MultipleViews2 {
	res := &MultipleViews2{
		B: vres.B,
	}
	if vres.A != nil {
		res.A = *vres.A
	}
	return res
}

// newMultipleViews2Tiny converts projected type MultipleViews2 to service type
// MultipleViews2.
func newMultipleViews2Tiny(vres *resultwithotherresultviews.MultipleViews2View) *MultipleViews2 {
	res := &MultipleViews2{}
	if vres.A != nil {
		res.A = *vres.A
	}
	return res
}

// newMultipleViews2View projects result type MultipleViews2 to projected type
// MultipleViews2View using the "default" view.
func newMultipleViews2View(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2View {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViews2ViewTiny projects result type MultipleViews2 to projected
// type MultipleViews2View using the "tiny" view.
func newMultipleViews2ViewTiny(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2View {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
	}
	return vres
}
`

const ResultWithResultCollectionMethod = `
// Service is the ResultWithResultTypeCollection service interface.
type Service interface {
	// A implements A.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "extended"
	//	- "tiny"
	A(context.Context) (res *RT, view string, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ResultWithResultTypeCollection"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// RT is the result type of the ResultWithResultTypeCollection service A method.
type RT struct {
	A RT2Collection
}

type RT2 struct {
	C string
	D int
	E *string
}

type RT2Collection []*RT2

// NewRT initializes result type RT from viewed result type RT.
func NewRT(vres *resultwithresulttypecollectionviews.RT) *RT {
	var res *RT
	switch vres.View {
	case "default", "":
		res = newRT(vres.Projected)
	case "extended":
		res = newRTExtended(vres.Projected)
	case "tiny":
		res = newRTTiny(vres.Projected)
	}
	return res
}

// NewViewedRT initializes viewed result type RT from result type RT using the
// given view.
func NewViewedRT(res *RT, view string) *resultwithresulttypecollectionviews.RT {
	var vres *resultwithresulttypecollectionviews.RT
	switch view {
	case "default", "":
		p := newRTView(res)
		vres = &resultwithresulttypecollectionviews.RT{Projected: p, View: "default"}
	case "extended":
		p := newRTViewExtended(res)
		vres = &resultwithresulttypecollectionviews.RT{Projected: p, View: "extended"}
	case "tiny":
		p := newRTViewTiny(res)
		vres = &resultwithresulttypecollectionviews.RT{Projected: p, View: "tiny"}
	}
	return vres
}

// newRT converts projected type RT to service type RT.
func newRT(vres *resultwithresulttypecollectionviews.RTView) *RT {
	res := &RT{}
	if vres.A != nil {
		res.A = newRT2Collection(vres.A)
	}
	return res
}

// newRTExtended converts projected type RT to service type RT.
func newRTExtended(vres *resultwithresulttypecollectionviews.RTView) *RT {
	res := &RT{}
	if vres.A != nil {
		res.A = newRT2CollectionExtended(vres.A)
	}
	return res
}

// newRTTiny converts projected type RT to service type RT.
func newRTTiny(vres *resultwithresulttypecollectionviews.RTView) *RT {
	res := &RT{}
	if vres.A != nil {
		res.A = newRT2CollectionTiny(vres.A)
	}
	return res
}

// newRTView projects result type RT to projected type RTView using the
// "default" view.
func newRTView(res *RT) *resultwithresulttypecollectionviews.RTView {
	vres := &resultwithresulttypecollectionviews.RTView{}
	if res.A != nil {
		vres.A = newRT2CollectionView(res.A)
	}
	return vres
}

// newRTViewExtended projects result type RT to projected type RTView using the
// "extended" view.
func newRTViewExtended(res *RT) *resultwithresulttypecollectionviews.RTView {
	vres := &resultwithresulttypecollectionviews.RTView{}
	if res.A != nil {
		vres.A = newRT2CollectionViewExtended(res.A)
	}
	return vres
}

// newRTViewTiny projects result type RT to projected type RTView using the
// "tiny" view.
func newRTViewTiny(res *RT) *resultwithresulttypecollectionviews.RTView {
	vres := &resultwithresulttypecollectionviews.RTView{}
	if res.A != nil {
		vres.A = newRT2CollectionViewTiny(res.A)
	}
	return vres
}

// newRT2Collection converts projected type RT2Collection to service type
// RT2Collection.
func newRT2Collection(vres resultwithresulttypecollectionviews.RT2CollectionView) RT2Collection {
	res := make(RT2Collection, len(vres))
	for i, n := range vres {
		res[i] = newRT2(n)
	}
	return res
}

// newRT2CollectionExtended converts projected type RT2Collection to service
// type RT2Collection.
func newRT2CollectionExtended(vres resultwithresulttypecollectionviews.RT2CollectionView) RT2Collection {
	res := make(RT2Collection, len(vres))
	for i, n := range vres {
		res[i] = newRT2Extended(n)
	}
	return res
}

// newRT2CollectionTiny converts projected type RT2Collection to service type
// RT2Collection.
func newRT2CollectionTiny(vres resultwithresulttypecollectionviews.RT2CollectionView) RT2Collection {
	res := make(RT2Collection, len(vres))
	for i, n := range vres {
		res[i] = newRT2Tiny(n)
	}
	return res
}

// newRT2CollectionView projects result type RT2Collection to projected type
// RT2CollectionView using the "default" view.
func newRT2CollectionView(res RT2Collection) resultwithresulttypecollectionviews.RT2CollectionView {
	vres := make(resultwithresulttypecollectionviews.RT2CollectionView, len(res))
	for i, n := range res {
		vres[i] = newRT2View(n)
	}
	return vres
}

// newRT2CollectionViewExtended projects result type RT2Collection to projected
// type RT2CollectionView using the "extended" view.
func newRT2CollectionViewExtended(res RT2Collection) resultwithresulttypecollectionviews.RT2CollectionView {
	vres := make(resultwithresulttypecollectionviews.RT2CollectionView, len(res))
	for i, n := range res {
		vres[i] = newRT2ViewExtended(n)
	}
	return vres
}

// newRT2CollectionViewTiny projects result type RT2Collection to projected
// type RT2CollectionView using the "tiny" view.
func newRT2CollectionViewTiny(res RT2Collection) resultwithresulttypecollectionviews.RT2CollectionView {
	vres := make(resultwithresulttypecollectionviews.RT2CollectionView, len(res))
	for i, n := range res {
		vres[i] = newRT2ViewTiny(n)
	}
	return vres
}

// newRT2 converts projected type RT2 to service type RT2.
func newRT2(vres *resultwithresulttypecollectionviews.RT2View) *RT2 {
	res := &RT2{}
	if vres.C != nil {
		res.C = *vres.C
	}
	if vres.D != nil {
		res.D = *vres.D
	}
	return res
}

// newRT2Extended converts projected type RT2 to service type RT2.
func newRT2Extended(vres *resultwithresulttypecollectionviews.RT2View) *RT2 {
	res := &RT2{
		E: vres.E,
	}
	if vres.C != nil {
		res.C = *vres.C
	}
	if vres.D != nil {
		res.D = *vres.D
	}
	return res
}

// newRT2Tiny converts projected type RT2 to service type RT2.
func newRT2Tiny(vres *resultwithresulttypecollectionviews.RT2View) *RT2 {
	res := &RT2{}
	if vres.D != nil {
		res.D = *vres.D
	}
	return res
}

// newRT2View projects result type RT2 to projected type RT2View using the
// "default" view.
func newRT2View(res *RT2) *resultwithresulttypecollectionviews.RT2View {
	vres := &resultwithresulttypecollectionviews.RT2View{
		C: &res.C,
		D: &res.D,
	}
	return vres
}

// newRT2ViewExtended projects result type RT2 to projected type RT2View using
// the "extended" view.
func newRT2ViewExtended(res *RT2) *resultwithresulttypecollectionviews.RT2View {
	vres := &resultwithresulttypecollectionviews.RT2View{
		C: &res.C,
		D: &res.D,
		E: res.E,
	}
	return vres
}

// newRT2ViewTiny projects result type RT2 to projected type RT2View using the
// "tiny" view.
func newRT2ViewTiny(res *RT2) *resultwithresulttypecollectionviews.RT2View {
	vres := &resultwithresulttypecollectionviews.RT2View{
		D: &res.D,
	}
	return vres
}
`

const ResultWithDashedMimeTypeMethod = `
// Service is the ResultWithDashedMimeType service interface.
type Service interface {
	// A implements A.
	A(context.Context) (res *ApplicationDashedType, err error)
	// List implements list.
	List(context.Context) (res *ListResult, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ResultWithDashedMimeType"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "list"}

// ApplicationDashedType is the result type of the ResultWithDashedMimeType
// service A method.
type ApplicationDashedType struct {
	Name *string
}

type ApplicationDashedTypeCollection []*ApplicationDashedType

// ListResult is the result type of the ResultWithDashedMimeType service list
// method.
type ListResult struct {
	Items ApplicationDashedTypeCollection
}

// NewApplicationDashedType initializes result type ApplicationDashedType from
// viewed result type ApplicationDashedType.
func NewApplicationDashedType(vres *resultwithdashedmimetypeviews.ApplicationDashedType) *ApplicationDashedType {
	return newApplicationDashedType(vres.Projected)
}

// NewViewedApplicationDashedType initializes viewed result type
// ApplicationDashedType from result type ApplicationDashedType using the given
// view.
func NewViewedApplicationDashedType(res *ApplicationDashedType, view string) *resultwithdashedmimetypeviews.ApplicationDashedType {
	p := newApplicationDashedTypeView(res)
	return &resultwithdashedmimetypeviews.ApplicationDashedType{Projected: p, View: "default"}
}

// newApplicationDashedType converts projected type ApplicationDashedType to
// service type ApplicationDashedType.
func newApplicationDashedType(vres *resultwithdashedmimetypeviews.ApplicationDashedTypeView) *ApplicationDashedType {
	res := &ApplicationDashedType{
		Name: vres.Name,
	}
	return res
}

// newApplicationDashedTypeView projects result type ApplicationDashedType to
// projected type ApplicationDashedTypeView using the "default" view.
func newApplicationDashedTypeView(res *ApplicationDashedType) *resultwithdashedmimetypeviews.ApplicationDashedTypeView {
	vres := &resultwithdashedmimetypeviews.ApplicationDashedTypeView{
		Name: res.Name,
	}
	return vres
}
`

const ForceGenerateType = `
// Service is the ForceGenerateType service interface.
type Service interface {
	// A implements A.
	A(context.Context) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ForceGenerateType"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

type ForcedType struct {
	A *string
}
`

const ForceGenerateTypeExplicit = `
// Service is the ForceGenerateTypeExplicit service interface.
type Service interface {
	// A implements A.
	A(context.Context) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "ForceGenerateTypeExplicit"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

type ForcedType struct {
	A *string
}
`

const StreamingResultMethod = `
// Service is the StreamingResultService service interface.
type Service interface {
	// StreamingResultMethod implements StreamingResultMethod.
	StreamingResultMethod(context.Context, *APayload, StreamingResultMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingResultService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingResultMethod"}

// StreamingResultMethodServerStream is the interface a "StreamingResultMethod"
// endpoint server stream must satisfy.
type StreamingResultMethodServerStream interface {
	// Send streams instances of "AResult".
	Send(*AResult) error
	// Close closes the stream.
	Close() error
}

// StreamingResultMethodClientStream is the interface a "StreamingResultMethod"
// endpoint client stream must satisfy.
type StreamingResultMethodClientStream interface {
	// Recv reads instances of "AResult" from the stream.
	Recv() (*AResult, error)
}

// APayload is the payload type of the StreamingResultService service
// StreamingResultMethod method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the StreamingResultService service
// StreamingResultMethod method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

const StreamingResultWithViewsMethod = `
// Service is the StreamingResultWithViewsService service interface.
type Service interface {
	// StreamingResultWithViewsMethod implements StreamingResultWithViewsMethod.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "tiny"
	StreamingResultWithViewsMethod(context.Context, string, StreamingResultWithViewsMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingResultWithViewsService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingResultWithViewsMethod"}

// StreamingResultWithViewsMethodServerStream is the interface a
// "StreamingResultWithViewsMethod" endpoint server stream must satisfy.
type StreamingResultWithViewsMethodServerStream interface {
	// Send streams instances of "MultipleViews".
	Send(*MultipleViews) error
	// Close closes the stream.
	Close() error
	// SetView sets the view used to render the result before streaming.
	SetView(view string)
}

// StreamingResultWithViewsMethodClientStream is the interface a
// "StreamingResultWithViewsMethod" endpoint client stream must satisfy.
type StreamingResultWithViewsMethodClientStream interface {
	// Recv reads instances of "MultipleViews" from the stream.
	Recv() (*MultipleViews, error)
}

// MultipleViews is the result type of the StreamingResultWithViewsService
// service StreamingResultWithViewsMethod method.
type MultipleViews struct {
	A *string
	B *string
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *streamingresultwithviewsserviceviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *streamingresultwithviewsserviceviews.MultipleViews {
	var vres *streamingresultwithviewsserviceviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &streamingresultwithviewsserviceviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &streamingresultwithviewsserviceviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *streamingresultwithviewsserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *streamingresultwithviewsserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *streamingresultwithviewsserviceviews.MultipleViewsView {
	vres := &streamingresultwithviewsserviceviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *streamingresultwithviewsserviceviews.MultipleViewsView {
	vres := &streamingresultwithviewsserviceviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const StreamingResultWithExplicitViewMethod = `
// Service is the StreamingResultWithExplicitViewService service interface.
type Service interface {
	// StreamingResultWithExplicitViewMethod implements
	// StreamingResultWithExplicitViewMethod.
	StreamingResultWithExplicitViewMethod(context.Context, []int32, StreamingResultWithExplicitViewMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingResultWithExplicitViewService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingResultWithExplicitViewMethod"}

// StreamingResultWithExplicitViewMethodServerStream is the interface a
// "StreamingResultWithExplicitViewMethod" endpoint server stream must satisfy.
type StreamingResultWithExplicitViewMethodServerStream interface {
	// Send streams instances of "MultipleViews".
	Send(*MultipleViews) error
	// Close closes the stream.
	Close() error
}

// StreamingResultWithExplicitViewMethodClientStream is the interface a
// "StreamingResultWithExplicitViewMethod" endpoint client stream must satisfy.
type StreamingResultWithExplicitViewMethodClientStream interface {
	// Recv reads instances of "MultipleViews" from the stream.
	Recv() (*MultipleViews, error)
}

// MultipleViews is the result type of the
// StreamingResultWithExplicitViewService service
// StreamingResultWithExplicitViewMethod method.
type MultipleViews struct {
	A *string
	B *string
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *streamingresultwithexplicitviewserviceviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *streamingresultwithexplicitviewserviceviews.MultipleViews {
	var vres *streamingresultwithexplicitviewserviceviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &streamingresultwithexplicitviewserviceviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &streamingresultwithexplicitviewserviceviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *streamingresultwithexplicitviewserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *streamingresultwithexplicitviewserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *streamingresultwithexplicitviewserviceviews.MultipleViewsView {
	vres := &streamingresultwithexplicitviewserviceviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *streamingresultwithexplicitviewserviceviews.MultipleViewsView {
	vres := &streamingresultwithexplicitviewserviceviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const StreamingResultNoPayloadMethod = `
// Service is the StreamingResultNoPayloadService service interface.
type Service interface {
	// StreamingResultNoPayloadMethod implements StreamingResultNoPayloadMethod.
	StreamingResultNoPayloadMethod(context.Context, StreamingResultNoPayloadMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingResultNoPayloadService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingResultNoPayloadMethod"}

// StreamingResultNoPayloadMethodServerStream is the interface a
// "StreamingResultNoPayloadMethod" endpoint server stream must satisfy.
type StreamingResultNoPayloadMethodServerStream interface {
	// Send streams instances of "AResult".
	Send(*AResult) error
	// Close closes the stream.
	Close() error
}

// StreamingResultNoPayloadMethodClientStream is the interface a
// "StreamingResultNoPayloadMethod" endpoint client stream must satisfy.
type StreamingResultNoPayloadMethodClientStream interface {
	// Recv reads instances of "AResult" from the stream.
	Recv() (*AResult, error)
}

// AResult is the result type of the StreamingResultNoPayloadService service
// StreamingResultNoPayloadMethod method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}
`

const StreamingPayloadMethod = `
// Service is the StreamingPayloadService service interface.
type Service interface {
	// StreamingPayloadMethod implements StreamingPayloadMethod.
	StreamingPayloadMethod(context.Context, *BPayload, StreamingPayloadMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingPayloadService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingPayloadMethod"}

// StreamingPayloadMethodServerStream is the interface a
// "StreamingPayloadMethod" endpoint server stream must satisfy.
type StreamingPayloadMethodServerStream interface {
	// SendAndClose streams instances of "AResult" and closes the stream.
	SendAndClose(*AResult) error
	// Recv reads instances of "APayload" from the stream.
	Recv() (*APayload, error)
}

// StreamingPayloadMethodClientStream is the interface a
// "StreamingPayloadMethod" endpoint client stream must satisfy.
type StreamingPayloadMethodClientStream interface {
	// Send streams instances of "APayload".
	Send(*APayload) error
	// CloseAndRecv stops sending messages to the stream and reads instances of
	// "AResult" from the stream.
	CloseAndRecv() (*AResult, error)
}

// APayload is the streaming payload type of the StreamingPayloadService
// service StreamingPayloadMethod method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the StreamingPayloadService service
// StreamingPayloadMethod method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// BPayload is the payload type of the StreamingPayloadService service
// StreamingPayloadMethod method.
type BPayload struct {
	ArrayField  []bool
	MapField    map[int]string
	ObjectField *struct {
		IntField    *int
		StringField *string
	}
	UserTypeField *Parent
}

type Child struct {
	P *Parent
}

type Parent struct {
	C *Child
}
`

const StreamingPayloadNoPayloadMethod = `
// Service is the StreamingPayloadNoPayloadService service interface.
type Service interface {
	// StreamingPayloadNoPayloadMethod implements StreamingPayloadNoPayloadMethod.
	StreamingPayloadNoPayloadMethod(context.Context, StreamingPayloadNoPayloadMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingPayloadNoPayloadService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingPayloadNoPayloadMethod"}

// StreamingPayloadNoPayloadMethodServerStream is the interface a
// "StreamingPayloadNoPayloadMethod" endpoint server stream must satisfy.
type StreamingPayloadNoPayloadMethodServerStream interface {
	// SendAndClose streams instances of "string" and closes the stream.
	SendAndClose(string) error
	// Recv reads instances of "interface{}" from the stream.
	Recv() (interface{}, error)
}

// StreamingPayloadNoPayloadMethodClientStream is the interface a
// "StreamingPayloadNoPayloadMethod" endpoint client stream must satisfy.
type StreamingPayloadNoPayloadMethodClientStream interface {
	// Send streams instances of "interface{}".
	Send(interface{}) error
	// CloseAndRecv stops sending messages to the stream and reads instances of
	// "string" from the stream.
	CloseAndRecv() (string, error)
}
`

const StreamingPayloadNoResultMethod = `
// Service is the StreamingPayloadNoResultService service interface.
type Service interface {
	// StreamingPayloadNoResultMethod implements StreamingPayloadNoResultMethod.
	StreamingPayloadNoResultMethod(context.Context, StreamingPayloadNoResultMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingPayloadNoResultService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingPayloadNoResultMethod"}

// StreamingPayloadNoResultMethodServerStream is the interface a
// "StreamingPayloadNoResultMethod" endpoint server stream must satisfy.
type StreamingPayloadNoResultMethodServerStream interface {
	// Recv reads instances of "int" from the stream.
	Recv() (int, error)
	// Close closes the stream.
	Close() error
}

// StreamingPayloadNoResultMethodClientStream is the interface a
// "StreamingPayloadNoResultMethod" endpoint client stream must satisfy.
type StreamingPayloadNoResultMethodClientStream interface {
	// Send streams instances of "int".
	Send(int) error
	// Close closes the stream.
	Close() error
}
`

const StreamingPayloadResultWithViewsMethod = `
// Service is the StreamingPayloadResultWithViewsService service interface.
type Service interface {
	// StreamingPayloadResultWithViewsMethod implements
	// StreamingPayloadResultWithViewsMethod.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "tiny"
	StreamingPayloadResultWithViewsMethod(context.Context, StreamingPayloadResultWithViewsMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingPayloadResultWithViewsService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingPayloadResultWithViewsMethod"}

// StreamingPayloadResultWithViewsMethodServerStream is the interface a
// "StreamingPayloadResultWithViewsMethod" endpoint server stream must satisfy.
type StreamingPayloadResultWithViewsMethodServerStream interface {
	// SendAndClose streams instances of "MultipleViews" and closes the stream.
	SendAndClose(*MultipleViews) error
	// Recv reads instances of "APayload" from the stream.
	Recv() (*APayload, error)
	// SetView sets the view used to render the result before streaming.
	SetView(view string)
}

// StreamingPayloadResultWithViewsMethodClientStream is the interface a
// "StreamingPayloadResultWithViewsMethod" endpoint client stream must satisfy.
type StreamingPayloadResultWithViewsMethodClientStream interface {
	// Send streams instances of "APayload".
	Send(*APayload) error
	// CloseAndRecv stops sending messages to the stream and reads instances of
	// "MultipleViews" from the stream.
	CloseAndRecv() (*MultipleViews, error)
}

// APayload is the streaming payload type of the
// StreamingPayloadResultWithViewsService service
// StreamingPayloadResultWithViewsMethod method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// MultipleViews is the result type of the
// StreamingPayloadResultWithViewsService service
// StreamingPayloadResultWithViewsMethod method.
type MultipleViews struct {
	A *string
	B *string
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *streamingpayloadresultwithviewsserviceviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *streamingpayloadresultwithviewsserviceviews.MultipleViews {
	var vres *streamingpayloadresultwithviewsserviceviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &streamingpayloadresultwithviewsserviceviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &streamingpayloadresultwithviewsserviceviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *streamingpayloadresultwithviewsserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *streamingpayloadresultwithviewsserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *streamingpayloadresultwithviewsserviceviews.MultipleViewsView {
	vres := &streamingpayloadresultwithviewsserviceviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *streamingpayloadresultwithviewsserviceviews.MultipleViewsView {
	vres := &streamingpayloadresultwithviewsserviceviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const StreamingPayloadResultWithExplicitViewMethod = `
// Service is the StreamingPayloadResultWithExplicitViewService service
// interface.
type Service interface {
	// StreamingPayloadResultWithExplicitViewMethod implements
	// StreamingPayloadResultWithExplicitViewMethod.
	StreamingPayloadResultWithExplicitViewMethod(context.Context, StreamingPayloadResultWithExplicitViewMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "StreamingPayloadResultWithExplicitViewService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"StreamingPayloadResultWithExplicitViewMethod"}

// StreamingPayloadResultWithExplicitViewMethodServerStream is the interface a
// "StreamingPayloadResultWithExplicitViewMethod" endpoint server stream must
// satisfy.
type StreamingPayloadResultWithExplicitViewMethodServerStream interface {
	// SendAndClose streams instances of "MultipleViews" and closes the stream.
	SendAndClose(*MultipleViews) error
	// Recv reads instances of "[]string" from the stream.
	Recv() ([]string, error)
}

// StreamingPayloadResultWithExplicitViewMethodClientStream is the interface a
// "StreamingPayloadResultWithExplicitViewMethod" endpoint client stream must
// satisfy.
type StreamingPayloadResultWithExplicitViewMethodClientStream interface {
	// Send streams instances of "[]string".
	Send([]string) error
	// CloseAndRecv stops sending messages to the stream and reads instances of
	// "MultipleViews" from the stream.
	CloseAndRecv() (*MultipleViews, error)
}

// MultipleViews is the result type of the
// StreamingPayloadResultWithExplicitViewService service
// StreamingPayloadResultWithExplicitViewMethod method.
type MultipleViews struct {
	A *string
	B *string
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *streamingpayloadresultwithexplicitviewserviceviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *streamingpayloadresultwithexplicitviewserviceviews.MultipleViews {
	var vres *streamingpayloadresultwithexplicitviewserviceviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &streamingpayloadresultwithexplicitviewserviceviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &streamingpayloadresultwithexplicitviewserviceviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *streamingpayloadresultwithexplicitviewserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *streamingpayloadresultwithexplicitviewserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *streamingpayloadresultwithexplicitviewserviceviews.MultipleViewsView {
	vres := &streamingpayloadresultwithexplicitviewserviceviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *streamingpayloadresultwithexplicitviewserviceviews.MultipleViewsView {
	vres := &streamingpayloadresultwithexplicitviewserviceviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const BidirectionalStreamingMethod = `
// Service is the BidirectionalStreamingService service interface.
type Service interface {
	// BidirectionalStreamingMethod implements BidirectionalStreamingMethod.
	BidirectionalStreamingMethod(context.Context, *BPayload, BidirectionalStreamingMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "BidirectionalStreamingService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"BidirectionalStreamingMethod"}

// BidirectionalStreamingMethodServerStream is the interface a
// "BidirectionalStreamingMethod" endpoint server stream must satisfy.
type BidirectionalStreamingMethodServerStream interface {
	// Send streams instances of "AResult".
	Send(*AResult) error
	// Recv reads instances of "APayload" from the stream.
	Recv() (*APayload, error)
	// Close closes the stream.
	Close() error
}

// BidirectionalStreamingMethodClientStream is the interface a
// "BidirectionalStreamingMethod" endpoint client stream must satisfy.
type BidirectionalStreamingMethodClientStream interface {
	// Send streams instances of "APayload".
	Send(*APayload) error
	// Recv reads instances of "AResult" from the stream.
	Recv() (*AResult, error)
	// Close closes the stream.
	Close() error
}

// APayload is the streaming payload type of the BidirectionalStreamingService
// service BidirectionalStreamingMethod method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// AResult is the result type of the BidirectionalStreamingService service
// BidirectionalStreamingMethod method.
type AResult struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// BPayload is the payload type of the BidirectionalStreamingService service
// BidirectionalStreamingMethod method.
type BPayload struct {
	ArrayField  []bool
	MapField    map[int]string
	ObjectField *struct {
		IntField    *int
		StringField *string
	}
	UserTypeField *Parent
}

type Child struct {
	P *Parent
}

type Parent struct {
	C *Child
}
`

const BidirectionalStreamingNoPayloadMethod = `
// Service is the BidirectionalStreamingNoPayloadService service interface.
type Service interface {
	// BidirectionalStreamingNoPayloadMethod implements
	// BidirectionalStreamingNoPayloadMethod.
	BidirectionalStreamingNoPayloadMethod(context.Context, BidirectionalStreamingNoPayloadMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "BidirectionalStreamingNoPayloadService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"BidirectionalStreamingNoPayloadMethod"}

// BidirectionalStreamingNoPayloadMethodServerStream is the interface a
// "BidirectionalStreamingNoPayloadMethod" endpoint server stream must satisfy.
type BidirectionalStreamingNoPayloadMethodServerStream interface {
	// Send streams instances of "int".
	Send(int) error
	// Recv reads instances of "string" from the stream.
	Recv() (string, error)
	// Close closes the stream.
	Close() error
}

// BidirectionalStreamingNoPayloadMethodClientStream is the interface a
// "BidirectionalStreamingNoPayloadMethod" endpoint client stream must satisfy.
type BidirectionalStreamingNoPayloadMethodClientStream interface {
	// Send streams instances of "string".
	Send(string) error
	// Recv reads instances of "int" from the stream.
	Recv() (int, error)
	// Close closes the stream.
	Close() error
}
`

const BidirectionalStreamingResultWithViewsMethod = `
// Service is the BidirectionalStreamingResultWithViewsService service
// interface.
type Service interface {
	// BidirectionalStreamingResultWithViewsMethod implements
	// BidirectionalStreamingResultWithViewsMethod.
	// The "view" return value must have one of the following views
	//	- "default"
	//	- "tiny"
	BidirectionalStreamingResultWithViewsMethod(context.Context, BidirectionalStreamingResultWithViewsMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "BidirectionalStreamingResultWithViewsService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"BidirectionalStreamingResultWithViewsMethod"}

// BidirectionalStreamingResultWithViewsMethodServerStream is the interface a
// "BidirectionalStreamingResultWithViewsMethod" endpoint server stream must
// satisfy.
type BidirectionalStreamingResultWithViewsMethodServerStream interface {
	// Send streams instances of "MultipleViews".
	Send(*MultipleViews) error
	// Recv reads instances of "APayload" from the stream.
	Recv() (*APayload, error)
	// Close closes the stream.
	Close() error
	// SetView sets the view used to render the result before streaming.
	SetView(view string)
}

// BidirectionalStreamingResultWithViewsMethodClientStream is the interface a
// "BidirectionalStreamingResultWithViewsMethod" endpoint client stream must
// satisfy.
type BidirectionalStreamingResultWithViewsMethodClientStream interface {
	// Send streams instances of "APayload".
	Send(*APayload) error
	// Recv reads instances of "MultipleViews" from the stream.
	Recv() (*MultipleViews, error)
	// Close closes the stream.
	Close() error
}

// APayload is the streaming payload type of the
// BidirectionalStreamingResultWithViewsService service
// BidirectionalStreamingResultWithViewsMethod method.
type APayload struct {
	IntField      int
	StringField   string
	BooleanField  bool
	BytesField    []byte
	OptionalField *string
}

// MultipleViews is the result type of the
// BidirectionalStreamingResultWithViewsService service
// BidirectionalStreamingResultWithViewsMethod method.
type MultipleViews struct {
	A *string
	B *string
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *bidirectionalstreamingresultwithviewsserviceviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *bidirectionalstreamingresultwithviewsserviceviews.MultipleViews {
	var vres *bidirectionalstreamingresultwithviewsserviceviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &bidirectionalstreamingresultwithviewsserviceviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &bidirectionalstreamingresultwithviewsserviceviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *bidirectionalstreamingresultwithviewsserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *bidirectionalstreamingresultwithviewsserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *bidirectionalstreamingresultwithviewsserviceviews.MultipleViewsView {
	vres := &bidirectionalstreamingresultwithviewsserviceviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *bidirectionalstreamingresultwithviewsserviceviews.MultipleViewsView {
	vres := &bidirectionalstreamingresultwithviewsserviceviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const BidirectionalStreamingResultWithExplicitViewMethod = `
// Service is the BidirectionalStreamingResultWithExplicitViewService service
// interface.
type Service interface {
	// BidirectionalStreamingResultWithExplicitViewMethod implements
	// BidirectionalStreamingResultWithExplicitViewMethod.
	BidirectionalStreamingResultWithExplicitViewMethod(context.Context, BidirectionalStreamingResultWithExplicitViewMethodServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "BidirectionalStreamingResultWithExplicitViewService"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"BidirectionalStreamingResultWithExplicitViewMethod"}

// BidirectionalStreamingResultWithExplicitViewMethodServerStream is the
// interface a "BidirectionalStreamingResultWithExplicitViewMethod" endpoint
// server stream must satisfy.
type BidirectionalStreamingResultWithExplicitViewMethodServerStream interface {
	// Send streams instances of "MultipleViews".
	Send(*MultipleViews) error
	// Recv reads instances of "[][]byte" from the stream.
	Recv() ([][]byte, error)
	// Close closes the stream.
	Close() error
}

// BidirectionalStreamingResultWithExplicitViewMethodClientStream is the
// interface a "BidirectionalStreamingResultWithExplicitViewMethod" endpoint
// client stream must satisfy.
type BidirectionalStreamingResultWithExplicitViewMethodClientStream interface {
	// Send streams instances of "[][]byte".
	Send([][]byte) error
	// Recv reads instances of "MultipleViews" from the stream.
	Recv() (*MultipleViews, error)
	// Close closes the stream.
	Close() error
}

// MultipleViews is the result type of the
// BidirectionalStreamingResultWithExplicitViewService service
// BidirectionalStreamingResultWithExplicitViewMethod method.
type MultipleViews struct {
	A *string
	B *string
}

// NewMultipleViews initializes result type MultipleViews from viewed result
// type MultipleViews.
func NewMultipleViews(vres *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViews) *MultipleViews {
	var res *MultipleViews
	switch vres.View {
	case "default", "":
		res = newMultipleViews(vres.Projected)
	case "tiny":
		res = newMultipleViewsTiny(vres.Projected)
	}
	return res
}

// NewViewedMultipleViews initializes viewed result type MultipleViews from
// result type MultipleViews using the given view.
func NewViewedMultipleViews(res *MultipleViews, view string) *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViews {
	var vres *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViews
	switch view {
	case "default", "":
		p := newMultipleViewsView(res)
		vres = &bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViews{Projected: p, View: "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViews{Projected: p, View: "tiny"}
	}
	return vres
}

// newMultipleViews converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViews(vres *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
		B: vres.B,
	}
	return res
}

// newMultipleViewsTiny converts projected type MultipleViews to service type
// MultipleViews.
func newMultipleViewsTiny(vres *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViewsView) *MultipleViews {
	res := &MultipleViews{
		A: vres.A,
	}
	return res
}

// newMultipleViewsView projects result type MultipleViews to projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViewsView {
	vres := &bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews to projected
// type MultipleViewsView using the "tiny" view.
func newMultipleViewsViewTiny(res *MultipleViews) *bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViewsView {
	vres := &bidirectionalstreamingresultwithexplicitviewserviceviews.MultipleViewsView{
		A: res.A,
	}
	return vres
}
`

const PkgPath = `
// Service is the PkgPathMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *foo.Foo) (res *foo.Foo, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "PkgPathMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}
`

const PkgPathRecursive = `
// Service is the PkgPathRecursiveMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *foo.RecursiveFoo) (res *foo.RecursiveFoo, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "PkgPathRecursiveMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}
`

const PkgPathMultiple = `
// Service is the MultiplePkgPathMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *bar.Bar) (res *bar.Bar, err error)
	// B implements B.
	B(context.Context, *baz.Baz) (res *baz.Baz, err error)
	// EnvelopedB implements EnvelopedB.
	EnvelopedB(context.Context, *EnvelopedBPayload) (res *EnvelopedBResult, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "MultiplePkgPathMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [3]string{"A", "B", "EnvelopedB"}

// EnvelopedBPayload is the payload type of the MultiplePkgPathMethod service
// EnvelopedB method.
type EnvelopedBPayload struct {
	Baz *Baz
}

// EnvelopedBResult is the result type of the MultiplePkgPathMethod service
// EnvelopedB method.
type EnvelopedBResult struct {
	Baz *Baz
}
`

const PkgPathFoo = `// Foo is the payload type of the PkgPathMethod service A method.
type Foo struct {
	IntField *int
}
`

const PkgPathRecursiveFooFoo = `
type Foo struct {
	IntField *int
}
`

const PkgPathRecursiveFoo = `// RecursiveFoo is the payload type of the PkgPathRecursiveMethod service A
// method.
type RecursiveFoo struct {
	Foo *Foo
}
`

const PkgPathBar = `// Bar is the payload type of the MultiplePkgPathMethod service A method.
type Bar struct {
	IntField *int
}
`

const PkgPathBaz = `// Baz is the payload type of the MultiplePkgPathMethod service B method.
type Baz struct {
	IntField *int
}
`

const PkgPathNoDir = `
// Service is the NoDirMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *NoDir) (res *NoDir, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "NoDirMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"A"}

// NoDir is the payload type of the NoDirMethod service A method.
type NoDir struct {
	IntField *int
}
`

const PkgPathDupe1 = `
// Service is the PkgPathDupeMethod service interface.
type Service interface {
	// A implements A.
	A(context.Context, *foo.Foo) (res *foo.Foo, err error)
	// B implements B.
	B(context.Context, *foo.Foo) (res *foo.Foo, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "PkgPathDupeMethod"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "B"}
`

const PkgPathFooDupe = `// Foo is the payload type of the PkgPathDupeMethod service A method.
type Foo struct {
	IntField *int
}
`

const PkgPathDupe2 = `
// Service is the PkgPathDupeMethod2 service interface.
type Service interface {
	// A implements A.
	A(context.Context, *foo.Foo) (res *foo.Foo, err error)
	// B implements B.
	B(context.Context, *foo.Foo) (res *foo.Foo, err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "PkgPathDupeMethod2"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"A", "B"}
`
