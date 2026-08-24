// This file records every package-level Go declaration written by the service
// and views generators. Each definition and reference reads its name from the
// same NameDeclaration.
package service

import (
	"cmp"
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// serviceNameRole identifies what one package-level declaration does in the
	// generated service or views package.
	serviceNameRole uint8

	// serviceNameOrder contains design names and fixed categories that order two
	// declarations requesting the same Go name. It does not depend on the order
	// in which generators find them.
	serviceNameOrder struct {
		role       serviceNameRole
		service    string
		api        string
		method     string
		subject    string
		view       string
		source     string
		target     string
		side       string
		occurrence int
		required   bool
	}

	// serviceSymbolID identifies one package declaration without using the Go
	// name that will be chosen later. Source and target distinguish conversion
	// helpers; subject and view distinguish constructors and validators.
	serviceSymbolID serviceNameOrder

	// serviceName stores the requested Go name and the NameDeclaration created
	// for it. Repeated collection must return that same declaration.
	serviceName struct {
		preferred   string
		base        *codegen.NameDeclaration
		prefix      string
		suffix      string
		declaration *codegen.NameDeclaration
	}

	// serviceNames maps each service declaration purpose to the NameDeclaration
	// stored in its generated Go package.
	serviceNames map[serviceSymbolID]serviceName
)

const (
	serviceInterfaceNameRole serviceNameRole = iota + 1
	serviceAutherNameRole
	serviceAPINameRole
	serviceAPIVersionNameRole
	serviceNameConstantRole
	serviceMethodNamesRole
	serviceServerStreamNameRole
	serviceClientStreamNameRole
	serviceErrorConstructorNameRole
	serviceViewConstructorNameRole
	servicePrivateProjectionConstructorNameRole
	serviceValidatorNameRole
	serviceViewMapNameRole
	serviceEndpointsNameRole
	serviceNewEndpointsNameRole
	serviceClientNameRole
	serviceNewClientNameRole
	serviceMethodEndpointNameRole
	serviceEndpointInputNameRole
	serviceRequestNameRole
	serviceResponseNameRole
	serviceServerInterceptorsNameRole
	serviceClientInterceptorsNameRole
	serviceInterceptorInfoNameRole
	serviceInterceptorPayloadNameRole
	serviceInterceptorResultNameRole
	serviceInterceptorStreamingPayloadNameRole
	serviceInterceptorStreamingResultNameRole
	serviceInterceptorPayloadAccessNameRole
	serviceInterceptorResultAccessNameRole
	serviceInterceptorStreamingPayloadAccessNameRole
	serviceInterceptorStreamingResultAccessNameRole
	serviceInterceptorMethodInfoNameRole
	serviceInterceptorServerUnaryInfoNameRole
	serviceInterceptorClientUnaryInfoNameRole
	serviceInterceptorStreamingSendInfoNameRole
	serviceInterceptorStreamingRecvInfoNameRole
	serviceServerEndpointWrapperNameRole
	serviceClientEndpointWrapperNameRole
	serviceServerInterceptorWrapperNameRole
	serviceClientInterceptorWrapperNameRole
	serviceServerStreamWrapperNameRole
	serviceClientStreamWrapperNameRole
	serviceTransformHelperNameRole
	serviceExampleStructNameRole
	serviceExampleConstructorNameRole
	serviceExampleServerInterceptorsStructNameRole
	serviceExampleServerInterceptorsConstructorNameRole
	serviceExampleClientInterceptorsStructNameRole
	serviceExampleClientInterceptorsConstructorNameRole
)

// ComparePackageName orders service declarations by their purpose and design
// names, so discovery order cannot change generated Go names.
func (o serviceNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(serviceNameOrder)
	if compared := cmp.Compare(o.role, right.role); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.service, right.service); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.api, right.api); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.method, right.method); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.subject, right.subject); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.view, right.view); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.source, right.source); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.target, right.target); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.side, right.side); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.occurrence, right.occurrence); compared != 0 {
		return compared
	}
	if o.required == right.required {
		return 0
	}
	if !o.required {
		return -1
	}
	return 1
}

// kind returns whether this role writes a Go type, function, constant, or
// variable. An unknown role means the generator omitted a supported case.
func (r serviceNameRole) kind() codegen.PackageNameKind {
	switch r {
	case serviceAPINameRole, serviceAPIVersionNameRole, serviceNameConstantRole:
		return codegen.NameConstant
	case serviceMethodNamesRole, serviceViewMapNameRole:
		return codegen.NameVariable
	case serviceErrorConstructorNameRole,
		serviceViewConstructorNameRole,
		servicePrivateProjectionConstructorNameRole,
		serviceValidatorNameRole,
		serviceNewEndpointsNameRole,
		serviceNewClientNameRole,
		serviceMethodEndpointNameRole,
		serviceServerEndpointWrapperNameRole,
		serviceClientEndpointWrapperNameRole,
		serviceServerInterceptorWrapperNameRole,
		serviceClientInterceptorWrapperNameRole,
		serviceTransformHelperNameRole,
		serviceExampleConstructorNameRole,
		serviceExampleServerInterceptorsConstructorNameRole,
		serviceExampleClientInterceptorsConstructorNameRole:
		return codegen.NameFunction
	case serviceInterfaceNameRole,
		serviceAutherNameRole,
		serviceServerStreamNameRole,
		serviceClientStreamNameRole,
		serviceEndpointsNameRole,
		serviceClientNameRole,
		serviceEndpointInputNameRole,
		serviceRequestNameRole,
		serviceResponseNameRole,
		serviceServerInterceptorsNameRole,
		serviceClientInterceptorsNameRole,
		serviceInterceptorInfoNameRole,
		serviceInterceptorPayloadNameRole,
		serviceInterceptorResultNameRole,
		serviceInterceptorStreamingPayloadNameRole,
		serviceInterceptorStreamingResultNameRole,
		serviceInterceptorPayloadAccessNameRole,
		serviceInterceptorResultAccessNameRole,
		serviceInterceptorStreamingPayloadAccessNameRole,
		serviceInterceptorStreamingResultAccessNameRole,
		serviceInterceptorMethodInfoNameRole,
		serviceInterceptorServerUnaryInfoNameRole,
		serviceInterceptorClientUnaryInfoNameRole,
		serviceInterceptorStreamingSendInfoNameRole,
		serviceInterceptorStreamingRecvInfoNameRole,
		serviceServerStreamWrapperNameRole,
		serviceClientStreamWrapperNameRole,
		serviceExampleStructNameRole,
		serviceExampleServerInterceptorsStructNameRole,
		serviceExampleClientInterceptorsStructNameRole:
		return codegen.NameType
	default:
		panic(fmt.Sprintf("unknown service package name role %d", r))
	}
}

// visibility reports whether callers outside the generated package can use the
// declaration.
func (r serviceNameRole) visibility() codegen.PackageNameVisibility {
	switch r {
	case servicePrivateProjectionConstructorNameRole,
		serviceInterceptorPayloadAccessNameRole,
		serviceInterceptorResultAccessNameRole,
		serviceInterceptorStreamingPayloadAccessNameRole,
		serviceInterceptorStreamingResultAccessNameRole,
		serviceInterceptorMethodInfoNameRole,
		serviceInterceptorServerUnaryInfoNameRole,
		serviceInterceptorClientUnaryInfoNameRole,
		serviceInterceptorStreamingSendInfoNameRole,
		serviceInterceptorStreamingRecvInfoNameRole,
		serviceServerInterceptorWrapperNameRole,
		serviceClientInterceptorWrapperNameRole,
		serviceServerStreamWrapperNameRole,
		serviceClientStreamWrapperNameRole,
		serviceTransformHelperNameRole,
		serviceExampleStructNameRole:
		return codegen.UnexportedName
	default:
		return codegen.ExportedName
	}
}

// declare submits one requested Go name to pkg. Repeated calls for the same id
// return the same NameDeclaration and reject a different requested name.
func (n serviceNames) declare(pkg *codegen.GeneratedPackage, id serviceSymbolID, preferred string) (*codegen.NameDeclaration, error) {
	return n.declareForAPI(pkg, id, preferred, "")
}

// declareForAPI submits one generated name using the API to distinguish two
// roots that intentionally contribute to the same service package.
func (n serviceNames) declareForAPI(pkg *codegen.GeneratedPackage, id serviceSymbolID, preferred, api string) (*codegen.NameDeclaration, error) {
	if existing, ok := n[id]; ok {
		if existing.base != nil || existing.preferred != preferred {
			return nil, fmt.Errorf(
				"service symbol role %d cannot declare both %q and %q",
				id.role,
				existing.preferred,
				preferred,
			)
		}
		if err := pkg.DeclareName(existing.declaration); err != nil {
			return nil, err
		}
		return existing.declaration, nil
	}

	order := serviceNameOrder(id)
	order.api = api
	declaration := codegen.NewPreferredName(
		id.role.kind(),
		preferred,
		id.role.visibility(),
		order,
	)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	n[id] = serviceName{preferred: preferred, declaration: declaration}
	return declaration, nil
}

// declareDependent submits a declaration whose Go name is built by adding
// prefix and suffix to base's final name. Repeated calls for the same id must
// use the same base, prefix, and suffix.
func (n serviceNames) declareDependent(pkg *codegen.GeneratedPackage, id serviceSymbolID, base *codegen.NameDeclaration, prefix, suffix string) (*codegen.NameDeclaration, error) {
	return n.declareDependentForAPI(pkg, id, base, prefix, suffix, "")
}

// declareDependentForAPI submits one dependent generated name using the API to
// distinguish two roots that intentionally contribute to the same package.
func (n serviceNames) declareDependentForAPI(pkg *codegen.GeneratedPackage, id serviceSymbolID, base *codegen.NameDeclaration, prefix, suffix, api string) (*codegen.NameDeclaration, error) {
	if existing, ok := n[id]; ok {
		if existing.base != base || existing.prefix != prefix || existing.suffix != suffix {
			return nil, fmt.Errorf("service symbol role %d cannot change its dependent declaration family", id.role)
		}
		if err := pkg.DeclareName(existing.declaration); err != nil {
			return nil, err
		}
		return existing.declaration, nil
	}

	order := serviceNameOrder(id)
	order.api = api
	declaration, err := pkg.DeclareDependentName(
		id.role.kind(),
		base,
		prefix,
		suffix,
		order,
	)
	if err != nil {
		return nil, err
	}
	n[id] = serviceName{
		base:        base,
		prefix:      prefix,
		suffix:      suffix,
		declaration: declaration,
	}
	return declaration, nil
}

// declaration returns the NameDeclaration previously stored for id. It panics
// when name collection did not submit that id.
func (n serviceNames) declaration(id serviceSymbolID) *codegen.NameDeclaration {
	name, ok := n[id]
	if !ok {
		panic(fmt.Sprintf("service symbol role %d was not declared", id.role))
	}
	return name.declaration
}

// transformDataTypeName returns the design name and ID used to order one side
// of a generated conversion helper.
func transformDataTypeName(dataType expr.DataType) (string, string) {
	if userType, ok := dataType.(expr.UserType); ok {
		return userType.Name(), userType.ID()
	}
	return dataType.Name(), ""
}

// canonicalValidatorView returns an empty string for the default result view
// so it matches validation calls that omit a view name.
func canonicalValidatorView(view string) string {
	if view == expr.DefaultView {
		return ""
	}
	return view
}
