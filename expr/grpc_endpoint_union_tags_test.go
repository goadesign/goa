package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRPCTagsRejectsConstructorUnionBranchTagCollidingWithSibling(t *testing.T) {
	fields := &Object{
		{
			Name: "id",
			Attribute: &AttributeExpr{
				Type: String,
				Meta: MetaExpr{"rpc:tag": []string{"1"}},
			},
		},
		{
			Name: "choice",
			Attribute: &AttributeExpr{
				Type: &Union{
					TypeName: "Choice",
					Values: []*NamedAttributeExpr{
						{
							Name: "Text",
							Attribute: &AttributeExpr{
								Type: String,
								Meta: MetaExpr{"rpc:tag": []string{"1"}},
							},
						},
						{
							Name: "JSON",
							Attribute: &AttributeExpr{
								Type: String,
								Meta: MetaExpr{"rpc:tag": []string{"2"}},
							},
						},
					},
				},
			},
		},
	}

	verr := validateRPCTags(fields, grpcEndpointForTagValidationTest())
	require.EqualError(t, verr, `service "Service" gRPC endpoint "Method": field number 1 in attribute "choice.Text" already exists for attribute "id"`)
}

func TestValidateRPCTagsRejectsDuplicateConstructorUnionBranchTags(t *testing.T) {
	fields := &Object{
		{
			Name: "choice",
			Attribute: &AttributeExpr{
				Type: &Union{
					TypeName: "Choice",
					Values: []*NamedAttributeExpr{
						{
							Name: "Text",
							Attribute: &AttributeExpr{
								Type: String,
								Meta: MetaExpr{"rpc:tag": []string{"1"}},
							},
						},
						{
							Name: "JSON",
							Attribute: &AttributeExpr{
								Type: String,
								Meta: MetaExpr{"rpc:tag": []string{"1"}},
							},
						},
					},
				},
			},
		},
	}

	verr := validateRPCTags(fields, grpcEndpointForTagValidationTest())
	require.EqualError(t, verr, `service "Service" gRPC endpoint "Method": field number 1 in attribute "choice.JSON" already exists for attribute "choice.Text"`)
}

func grpcEndpointForTagValidationTest() *GRPCEndpointExpr {
	service := &ServiceExpr{Name: "Service"}
	return &GRPCEndpointExpr{
		MethodExpr: &MethodExpr{Name: "Method"},
		Service:    &GRPCServiceExpr{ServiceExpr: service},
	}
}
