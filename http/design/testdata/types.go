package testdata

import "goa.design/goa/design"

var FinalizeEndpointBodyAsExtendedType = &design.UserTypeExpr{
	AttributeExpr: &design.AttributeExpr{
		Type: &design.Object{
			{"id", &design.AttributeExpr{Type: design.String}},
			{"name", &design.AttributeExpr{Type: design.String}},
		},
	},
	TypeName: "FinalizeEndpointBodyAsExtendedType",
}

var FinalizeEndpointBodyAsPropWithExtendedType = &design.UserTypeExpr{
	AttributeExpr: &design.AttributeExpr{
		Type: &design.Object{
			{"id", &design.AttributeExpr{Type: design.String}},
			{"name", &design.AttributeExpr{Type: design.String}},
		},
	},
	TypeName: "FinalizeEndpointBodyAsPropWithExtendedTypeDSL",
}
