package codegen

import (
	"fmt"
	"net/http"
)

// statusCodeToHTTPConst produces the standard name for the given HTTP status
// code. If no standard name exists then the string consisting of the code
// integer value is returned.
func statusCodeToHTTPConst(statusCode int) string {
	if v, ok := statusCodeToConst[statusCode]; ok {
		return fmt.Sprintf("http.%s", v)
	}
	return fmt.Sprintf("%d", statusCode)
}

var statusCodeToConst = map[int]string{
	http.StatusContinue:                      "StatusContinue",
	http.StatusSwitchingProtocols:            "StatusSwitchingProtocols",
	http.StatusProcessing:                    "StatusProcessing",
	http.StatusOK:                            "StatusOK",
	http.StatusCreated:                       "StatusCreated",
	http.StatusAccepted:                      "StatusAccepted",
	http.StatusNonAuthoritativeInfo:          "StatusNonAuthoritativeInfo",
	http.StatusNoContent:                     "StatusNoContent",
	http.StatusResetContent:                  "StatusResetContent",
	http.StatusPartialContent:                "StatusPartialContent",
	http.StatusMultiStatus:                   "StatusMultiStatus",
	http.StatusAlreadyReported:               "StatusAlreadyReported",
	http.StatusIMUsed:                        "StatusIMUsed",
	http.StatusMultipleChoices:               "StatusMultipleChoices",
	http.StatusMovedPermanently:              "StatusMovedPermanently",
	http.StatusFound:                         "StatusFound",
	http.StatusSeeOther:                      "StatusSeeOther",
	http.StatusNotModified:                   "StatusNotModified",
	http.StatusUseProxy:                      "StatusUseProxy",
	http.StatusTemporaryRedirect:             "StatusTemporaryRedirect",
	http.StatusPermanentRedirect:             "StatusPermanentRedirect",
	http.StatusBadRequest:                    "StatusBadRequest",
	http.StatusUnauthorized:                  "StatusUnauthorized",
	http.StatusPaymentRequired:               "StatusPaymentRequired",
	http.StatusForbidden:                     "StatusForbidden",
	http.StatusNotFound:                      "StatusNotFound",
	http.StatusMethodNotAllowed:              "StatusMethodNotAllowed",
	http.StatusNotAcceptable:                 "StatusNotAcceptable",
	http.StatusProxyAuthRequired:             "StatusProxyAuthRequired",
	http.StatusRequestTimeout:                "StatusRequestTimeout",
	http.StatusConflict:                      "StatusConflict",
	http.StatusGone:                          "StatusGone",
	http.StatusLengthRequired:                "StatusLengthRequired",
	http.StatusPreconditionFailed:            "StatusPreconditionFailed",
	http.StatusRequestEntityTooLarge:         "StatusRequestEntityTooLarge",
	http.StatusRequestURITooLong:             "StatusRequestURITooLong",
	http.StatusUnsupportedMediaType:          "StatusUnsupportedMediaType",
	http.StatusRequestedRangeNotSatisfiable:  "StatusRequestedRangeNotSatisfiable",
	http.StatusExpectationFailed:             "StatusExpectationFailed",
	http.StatusTeapot:                        "StatusTeapot",
	http.StatusUnprocessableEntity:           "StatusUnprocessableEntity",
	http.StatusLocked:                        "StatusLocked",
	http.StatusFailedDependency:              "StatusFailedDependency",
	http.StatusUpgradeRequired:               "StatusUpgradeRequired",
	http.StatusPreconditionRequired:          "StatusPreconditionRequired",
	http.StatusTooManyRequests:               "StatusTooManyRequests",
	http.StatusRequestHeaderFieldsTooLarge:   "StatusRequestHeaderFieldsTooLarge",
	http.StatusUnavailableForLegalReasons:    "StatusUnavailableForLegalReasons",
	http.StatusInternalServerError:           "StatusInternalServerError",
	http.StatusNotImplemented:                "StatusNotImplemented",
	http.StatusBadGateway:                    "StatusBadGateway",
	http.StatusServiceUnavailable:            "StatusServiceUnavailable",
	http.StatusGatewayTimeout:                "StatusGatewayTimeout",
	http.StatusHTTPVersionNotSupported:       "StatusHTTPVersionNotSupported",
	http.StatusVariantAlsoNegotiates:         "StatusVariantAlsoNegotiates",
	http.StatusInsufficientStorage:           "StatusInsufficientStorage",
	http.StatusLoopDetected:                  "StatusLoopDetected",
	http.StatusNotExtended:                   "StatusNotExtended",
	http.StatusNetworkAuthenticationRequired: "StatusNetworkAuthenticationRequired",
}
