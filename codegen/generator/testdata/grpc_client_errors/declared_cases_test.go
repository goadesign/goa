// These fixtures expose named errors, explicit statuses, and custom As
// representations before or after independent causes. No fixture rewrites
// the returned error's diagnostic or infers a context error's origin.
package clienterrors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	gencatalog "generated.local/gen/catalog"
	goapb "goa.design/goa/v3/grpc/pb"
	goa "goa.design/goa/v3/pkg"
)

type (
	// declaredJoin keeps nil entries so the generated selector must count
	// effective causes rather than treating a one-cause join as independent.
	declaredJoin struct {
		causes []error
	}

	// ownedDenied gives the complete result a declared name and explicitly
	// supplies its custom fields through As, while retaining independent causes.
	ownedDenied struct {
		*gencatalog.Denied
		cause error
	}

	// declaredProxy exposes one custom error through As. Its cause list lets
	// tests distinguish a single representation from visible independent causes.
	declaredProxy struct {
		value  error
		causes []error
		st     *status.Status
	}

	// statusDenied supplies both a direct complete status and a named value.
	statusDenied struct {
		*ownedDenied
		st *status.Status
	}
)

func declaredOwnershipCases(t *testing.T, reverse bool) []declaredOwnershipCase {
	t.Helper()
	denied := &gencatalog.Denied{Reason: "catalog access rejected"}
	stopped := status.Error(codes.Canceled, "independent operation stopped")
	join := func(left, right error) error {
		if reverse {
			return errors.Join(right, left)
		}
		return errors.Join(left, right)
	}
	whole, err := status.New(codes.Aborted, "complete status diagnostic").WithDetails(
		&goapb.ErrorResponse{Name: "complete", Id: "complete-id", Msg: "complete detail", Fault: true},
		&emptypb.Empty{},
	)
	require.NoError(t, err)
	other, err := status.New(codes.Aborted, "other first detail").WithDetails(&emptypb.Empty{})
	require.NoError(t, err)
	later, err := anypb.New(&goapb.ErrorResponse{Name: "later", Id: "later-id", Msg: "not first"})
	require.NoError(t, err)
	unknown := status.FromProto(&statuspb.Status{
		Code: int32(codes.Aborted), Message: "unknown first detail",
		Details: []*anypb.Any{{TypeUrl: "type.googleapis.com/example.Unknown"}, later},
	})
	malformed := status.FromProto(&statuspb.Status{
		Code: int32(codes.Aborted), Message: "malformed first detail",
		Details: []*anypb.Any{{TypeUrl: "type.googleapis.com/goa.ErrorResponse", Value: []byte{0xff}}, later},
	})
	outer := goa.NewServiceError(join(denied, stopped), "other", false, false, true)
	outer.ID, outer.Message = "outer-id", "complete outer diagnostic"
	customName := goa.NewServiceError(join(denied, stopped), "denied", false, false, false)
	customName.ID, customName.Message = "outer-custom-id", "complete outer custom diagnostic"
	singleCustom := goa.NewServiceError(denied, "denied", false, false, false)
	owned := &ownedDenied{denied, join(stopped, errors.New("cleanup failed"))}
	proxy := &declaredProxy{value: denied}
	cases := []declaredOwnershipCase{
		{"direct", denied, codes.PermissionDenied, denied, nil, nil},
		{"single effective", &declaredJoin{[]error{nil, denied, nil}}, codes.PermissionDenied, denied, nil, nil},
		{"duplicate", join(denied, denied), codes.Unknown, nil, nil, nil},
		{"mixed status", join(denied, stopped), codes.Unknown, nil, nil, nil},
		{"mixed ordinary", join(denied, errors.New("cleanup failed")), codes.Unknown, nil, nil, nil},
		{"mixed fault", join(denied, goa.Fault("cleanup failed")), codes.Unknown, nil, nil, nil},
		{"outer undeclared", outer, codes.Internal, nil, outer, nil},
		{"outer custom value below join", customName, codes.Unknown, nil, customName, nil},
		{"single custom value below owner", singleCustom, codes.PermissionDenied, denied, nil, nil},
		{"owned custom As", owned, codes.PermissionDenied, denied, nil, nil},
		{"single As", proxy, codes.PermissionDenied, denied, nil, nil},
		{"single As overrides child name", &declaredProxy{value: denied, causes: []error{goa.Fault("other name")}}, codes.PermissionDenied, denied, nil, nil},
		{"single As status remains declared", &declaredProxy{value: denied, st: whole}, codes.PermissionDenied, denied, nil, nil},
		{"multi As", &declaredProxy{value: denied, causes: []error{stopped, goa.Fault("cleanup failed")}}, codes.Unknown, nil, nil, nil},
		{"whole outside declaration", &operationStatus{denied, whole}, codes.Aborted, nil, nil, whole},
		{"whole on declaration", &statusDenied{owned, whole}, codes.Aborted, nil, nil, whole},
		{"whole outside owned Goa", &operationStatus{outer, whole}, codes.Aborted, nil, outer, whole},
		{"other first whole", &operationStatus{denied, other}, codes.Aborted, nil, nil, other},
		{"unknown first whole", &operationStatus{denied, unknown}, codes.Aborted, nil, nil, unknown},
		{"malformed first whole", &operationStatus{denied, malformed}, codes.Aborted, nil, nil, malformed},
		{"nil status declaration", &operationStatus{denied, nil}, codes.PermissionDenied, denied, nil, nil},
		{"first nil status stays declared", &operationStatus{&operationStatus{denied, whole}, nil}, codes.PermissionDenied, denied, nil, nil},
		{"nil status mixed", &operationStatus{join(denied, stopped), nil}, codes.Unknown, nil, nil, nil},
	}
	return cases
}

func (e *declaredJoin) Error() string {
	return "single catalog diagnostic"
}

func (e *declaredJoin) Unwrap() []error {
	return e.causes
}

func (e *ownedDenied) Error() string {
	return "complete catalog rejection"
}

func (e *ownedDenied) Unwrap() error {
	return e.cause
}

func (e *ownedDenied) As(target any) bool {
	if typed, ok := target.(**gencatalog.Denied); ok {
		*typed = e.Denied
		return true
	}
	return false
}

func (e *declaredProxy) Error() string {
	return "custom catalog representation"
}

func (e *declaredProxy) Unwrap() []error {
	return e.causes
}

func (e *declaredProxy) As(target any) bool {
	switch typed := target.(type) {
	case *goa.GoaErrorNamer:
		if value, ok := e.value.(goa.GoaErrorNamer); ok {
			*typed = value
			return true
		}
	case **gencatalog.Denied:
		if value, ok := e.value.(*gencatalog.Denied); ok {
			*typed = value
			return true
		}
	case **goa.ServiceError:
		if value, ok := e.value.(*goa.ServiceError); ok {
			*typed = value
			return true
		}
	}
	if e.st != nil {
		return errors.As(&operationStatus{errors.New("explicit status"), e.st}, target)
	}
	return false
}

func (e *statusDenied) GRPCStatus() *status.Status {
	return e.st
}
