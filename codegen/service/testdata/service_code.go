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
		vres = &multiplemethodsresultmultipleviewsviews.MultipleViews{p, "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &multiplemethodsresultmultipleviewsviews.MultipleViews{p, "tiny"}
	}
	return vres
}

// NewSingleView initializes result type SingleView from viewed result type
// SingleView.
func NewSingleView(vres *multiplemethodsresultmultipleviewsviews.SingleView) *SingleView {
	var res *SingleView
	switch vres.View {
	case "default", "":
		res = newSingleView(vres.Projected)
	}
	return res
}

// NewViewedSingleView initializes viewed result type SingleView from result
// type SingleView using the given view.
func NewViewedSingleView(res *SingleView, view string) *multiplemethodsresultmultipleviewsviews.SingleView {
	var vres *multiplemethodsresultmultipleviewsviews.SingleView
	switch view {
	case "default", "":
		p := newSingleViewView(res)
		vres = &multiplemethodsresultmultipleviewsviews.SingleView{p, "default"}
	}
	return vres
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

// newMultipleViewsView projects result type MultipleViews into projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *multiplemethodsresultmultipleviewsviews.MultipleViewsView {
	vres := &multiplemethodsresultmultipleviewsviews.MultipleViewsView{
		A: res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews into projected
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

// newSingleViewView projects result type SingleView into projected type
// SingleViewView using the "default" view.
func newSingleViewView(res *SingleView) *multiplemethodsresultmultipleviewsviews.SingleViewView {
	vres := &multiplemethodsresultmultipleviewsviews.SingleViewView{
		A: res.A,
		B: res.B,
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
		vres = resultcollectionmultipleviewsmethodviews.MultipleViewsCollection{p, "default"}
	case "tiny":
		p := newMultipleViewsCollectionViewTiny(res)
		vres = resultcollectionmultipleviewsmethodviews.MultipleViewsCollection{p, "tiny"}
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
// into projected type MultipleViewsCollectionView using the "default" view.
func newMultipleViewsCollectionView(res MultipleViewsCollection) resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView {
	vres := make(resultcollectionmultipleviewsmethodviews.MultipleViewsCollectionView, len(res))
	for i, n := range res {
		vres[i] = newMultipleViewsView(n)
	}
	return vres
}

// newMultipleViewsCollectionViewTiny projects result type
// MultipleViewsCollection into projected type MultipleViewsCollectionView
// using the "tiny" view.
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

// newMultipleViewsView projects result type MultipleViews into projected type
// MultipleViewsView using the "default" view.
func newMultipleViewsView(res *MultipleViews) *resultcollectionmultipleviewsmethodviews.MultipleViewsView {
	vres := &resultcollectionmultipleviewsmethodviews.MultipleViewsView{
		A: &res.A,
		B: &res.B,
	}
	return vres
}

// newMultipleViewsViewTiny projects result type MultipleViews into projected
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
		vres = &resultwithotherresultviews.MultipleViews{p, "default"}
	case "tiny":
		p := newMultipleViewsViewTiny(res)
		vres = &resultwithotherresultviews.MultipleViews{p, "tiny"}
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

// newMultipleViewsView projects result type MultipleViews into projected type
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

// newMultipleViewsViewTiny projects result type MultipleViews into projected
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

// newMultipleViews2View projects result type MultipleViews2 into projected
// type MultipleViews2View using the "default" view.
func newMultipleViews2View(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2View {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
		B: res.B,
	}
	return vres
}

// newMultipleViews2ViewTiny projects result type MultipleViews2 into projected
// type MultipleViews2View using the "tiny" view.
func newMultipleViews2ViewTiny(res *MultipleViews2) *resultwithotherresultviews.MultipleViews2View {
	vres := &resultwithotherresultviews.MultipleViews2View{
		A: &res.A,
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
