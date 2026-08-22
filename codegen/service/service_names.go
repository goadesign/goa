// This file defines stable identities for every package-level declaration
// emitted by the core service and views generators. Retained service plans
// declare these records before generation freeze and render their final names
// from the same records afterward.
package service

import (
	"cmp"
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// serviceNameRole identifies one closed family of package-level declarations
	// emitted by the core service and views generators.
	serviceNameRole uint8

	// serviceNameOrder contains only stable semantic values, giving colliding
	// service declarations a deterministic total order across traversals.
	serviceNameOrder struct {
		role       serviceNameRole
		service    string
		method     string
		subject    string
		view       string
		source     string
		target     string
		side       string
		occurrence int
		required   bool
	}

	// serviceSymbolID identifies one package declaration without using its
	// provisional Go spelling. Source and target distinguish transform helpers;
	// subject and view distinguish constructors and validators.
	serviceSymbolID serviceNameOrder

	// serviceName retains the preferred spelling with its canonical declaration
	// so repeated collection cannot silently rename one semantic symbol.
	serviceName struct {
		preferred   string
		base        *codegen.NameDeclaration
		prefix      string
		suffix      string
		declaration *codegen.NameDeclaration
	}

	// serviceNames owns the core declarations collected for one retained service
	// plan. The declaration itself remains owned by its generated Go package.
	serviceNames map[serviceSymbolID]serviceName
)

const (
	serviceInterfaceNameRole serviceNameRole = iota + 1
	serviceAutherNameRole
	serviceAPINameRole
	serviceAPIVersionNameRole
	serviceNameConstantRole
	serviceMethodNamesRole
	serviceMethodEventNameRole
	serviceServerStreamNameRole
	serviceClientStreamNameRole
	serviceStreamNameRole
	serviceEventNameRole
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

// ComparePackageName orders declarations from the core service generator by
// their complete stable semantic identity.
func (o serviceNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(serviceNameOrder)
	if compared := cmp.Compare(o.role, right.role); compared != 0 {
		return compared
	}
	if compared := cmp.Compare(o.service, right.service); compared != 0 {
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

// kind returns the package declaration category fixed by this service symbol
// family. An unknown role is an internal planner bug.
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
		serviceMethodEventNameRole,
		serviceServerStreamNameRole,
		serviceClientStreamNameRole,
		serviceStreamNameRole,
		serviceEventNameRole,
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

// visibility reports whether the emitted declaration is part of the generated
// package API or an implementation detail used only by neighboring sections.
func (r serviceNameRole) visibility() codegen.PackageNameVisibility {
	switch r {
	case servicePrivateProjectionConstructorNameRole,
		serviceInterceptorPayloadAccessNameRole,
		serviceInterceptorResultAccessNameRole,
		serviceInterceptorStreamingPayloadAccessNameRole,
		serviceInterceptorStreamingResultAccessNameRole,
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

// declare records id in pkg and returns the same canonical declaration when a
// planning traversal encounters that exact semantic symbol again.
func (n serviceNames) declare(pkg *codegen.GeneratedPackage, id serviceSymbolID, preferred string) (*codegen.NameDeclaration, error) {
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

	declaration := codegen.NewPreferredName(
		id.role.kind(),
		preferred,
		id.role.visibility(),
		serviceNameOrder(id),
	)
	if err := pkg.DeclareName(declaration); err != nil {
		return nil, err
	}
	n[id] = serviceName{preferred: preferred, declaration: declaration}
	return declaration, nil
}

// declareDependent records a companion whose preferred spelling follows the
// exact final name of base. Repeated collection must use the same base record
// and affixes, so one semantic symbol cannot silently change families.
func (n serviceNames) declareDependent(pkg *codegen.GeneratedPackage, id serviceSymbolID, base *codegen.NameDeclaration, prefix, suffix string) (*codegen.NameDeclaration, error) {
	if existing, ok := n[id]; ok {
		if existing.base != base || existing.prefix != prefix || existing.suffix != suffix {
			return nil, fmt.Errorf("service symbol role %d cannot change its dependent declaration family", id.role)
		}
		if err := pkg.DeclareName(existing.declaration); err != nil {
			return nil, err
		}
		return existing.declaration, nil
	}

	declaration, err := pkg.DeclareDependentName(
		id.role.kind(),
		base,
		prefix,
		suffix,
		serviceNameOrder(id),
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

// declaration returns the canonical record for id. Calling it for a symbol
// that collection did not declare is an internal retained-plan bug.
func (n serviceNames) declaration(id serviceSymbolID) *codegen.NameDeclaration {
	name, ok := n[id]
	if !ok {
		panic(fmt.Sprintf("service symbol role %d was not declared", id.role))
	}
	return name.declaration
}

// transformDataTypeIdentity returns the authored declaration identity used by
// TransformPlan when the operation crosses copied named attributes.
func transformDataTypeIdentity(dataType expr.DataType) expr.DataType {
	if userType, ok := dataType.(expr.UserType); ok {
		return userType.Origin()
	}
	return dataType
}

// transformDataTypeName returns stable semantic labels for one helper side.
func transformDataTypeName(dataType expr.DataType) (string, string) {
	if userType, ok := dataType.(expr.UserType); ok {
		return userType.Name(), userType.ID()
	}
	return dataType.Name(), ""
}

// canonicalValidatorView gives a default result view the same identity used by
// validation calls that omit an explicit view.
func canonicalValidatorView(view string) string {
	if view == expr.DefaultView {
		return ""
	}
	return view
}
