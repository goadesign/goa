package testdata

const PayloadWithValidationsBuildCode = `// PayloadWithValidation gRPC client CLI support package
//
// Command:
// goa

package client

import (
	"fmt"
	payloadwithvalidation "payload_with_validation"
	"strconv"
	"unicode/utf8"

	goa "goa.design/goa/v3/pkg"
)

// BuildMethodAPayload builds the payload for the PayloadWithValidation
// method_a endpoint from CLI flags.
func BuildMethodAPayload(payloadWithValidationMethodAMetadataInt string, payloadWithValidationMethodAMetadataString string) (*payloadwithvalidation.MethodAPayload, error) {
	var err error
	var metadataInt *int
	{
		if payloadWithValidationMethodAMetadataInt != "" {
			var v int64
			v, err = strconv.ParseInt(payloadWithValidationMethodAMetadataInt, 10, 64)
			val := int(v)
			metadataInt = &val
			if err != nil {
				return nil, fmt.Errorf("invalid value for metadataInt, must be INT")
			}
			if metadataInt != nil {
				if *metadataInt < 0 {
					err = goa.MergeErrors(err, goa.InvalidRangeError("metadataInt", *metadataInt, 0, true))
				}
			}
			if metadataInt != nil {
				if *metadataInt > 100 {
					err = goa.MergeErrors(err, goa.InvalidRangeError("metadataInt", *metadataInt, 100, false))
				}
			}
			if err != nil {
				return nil, err
			}
		}
	}
	var metadataString *string
	{
		if payloadWithValidationMethodAMetadataString != "" {
			metadataString = &payloadWithValidationMethodAMetadataString
			if metadataString != nil {
				if utf8.RuneCountInString(*metadataString) < 5 {
					err = goa.MergeErrors(err, goa.InvalidLengthError("metadataString", *metadataString, utf8.RuneCountInString(*metadataString), 5, true))
				}
			}
			if metadataString != nil {
				if utf8.RuneCountInString(*metadataString) > 10 {
					err = goa.MergeErrors(err, goa.InvalidLengthError("metadataString", *metadataString, utf8.RuneCountInString(*metadataString), 10, false))
				}
			}
			if err != nil {
				return nil, err
			}
		}
	}
	v := &payloadwithvalidation.MethodAPayload{}
	v.MetadataInt = metadataInt
	v.MetadataString = metadataString

	return v, nil
}
`
