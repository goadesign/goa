package codegen

import (
	"fmt"

	"goa.design/goa/v3/dsl"
)

// statusCodeToGRPCConst produces the standard name for the given gRPC status
// code. If no standard name exists then the string consisting of the code
// integer value is returned.
func statusCodeToGRPCConst(code int) string {
	if v, ok := statusCodeToConst[code]; ok {
		return fmt.Sprintf("codes.%s", v)
	}
	return fmt.Sprintf("%d", code)
}

var statusCodeToConst = map[int]string{
	dsl.CodeOK:                 "OK",
	dsl.CodeCanceled:           "Canceled",
	dsl.CodeUnknown:            "Unknown",
	dsl.CodeInvalidArgument:    "InvalidArgument",
	dsl.CodeDeadlineExceeded:   "DeadlineExceeded",
	dsl.CodeNotFound:           "NotFound",
	dsl.CodeAlreadyExists:      "AlreadyExists",
	dsl.CodePermissionDenied:   "PermissionDenied",
	dsl.CodeResourceExhausted:  "ResourceExhausted",
	dsl.CodeFailedPrecondition: "FailedPrecondition",
	dsl.CodeAborted:            "Aborted",
	dsl.CodeOutOfRange:         "OutOfRange",
	dsl.CodeUnimplemented:      "Unimplemented",
	dsl.CodeInternal:           "Internal",
	dsl.CodeUnavailable:        "Unavailable",
	dsl.CodeDataLoss:           "DataLoss",
	dsl.CodeUnauthenticated:    "Unauthenticated",
}
