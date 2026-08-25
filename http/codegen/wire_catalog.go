// This file assigns Go names to request and response types in one generated
// HTTP or JSON-RPC package. Each copied type is recorded before names are
// assigned. Its definition, references, and validation function then use the
// same record.
package codegen

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// wireTypeCatalog stores every request or response type written into one Go
	// package and the Go name chosen for each type.
	wireTypeCatalog struct {
		pkg                  *codegen.GeneratedPackage
		scope                *codegen.NameScope
		records              []*wireTypeRecord
		transforms           []*wireTransformRecord
		unionOccurrences     []wireUnionOccurrence
		unions               []*wireUnionRecord
		validationRoots      []wireValidationRoot
		transformHelpers     []*wireTransformHelperRecord
		transformBindings    map[codegen.TransformHelperID]*wireTransformHelperRecord
		transformDefinitions map[*codegen.NameDeclaration]*codegen.TransformFunctionData
		releasedSuffixes     map[*expr.AttributeExpr]string
		declared             bool
		linked               bool
		bindings             map[*expr.AttributeExpr]*wireTypeRecord
		unionBindings        map[*expr.Union]*wireUnionRecord
	}

	// wireTransformHandle identifies one exact request or response conversion
	// recorded before generated package names are assigned.
	wireTransformHandle struct {
		catalog *wireTypeCatalog
		record  *wireTransformRecord
	}

	// wireTransformRecord stores one value conversion and any extra functions it
	// needs.
	wireTransformRecord struct {
		source *expr.AttributeExpr
		target *expr.AttributeExpr
		prefix string
		owner  string
		layout wireTransformLayout
		plan   *codegen.TransformPlan
		used   bool
	}

	// wireTransformLayout records which value belongs to the transport package,
	// which pointer rules it uses, and whether the other value belongs to the
	// generated service package or its views package.
	wireTransformLayout struct {
		wireSide       wireTransformSide
		wirePolicy     wireTypePolicy
		wireUse        wireUnionUse
		servicePointer bool
		servicePackage codegen.ImportSpec
	}

	// wireTransformHelperRecord stores one function declaration shared by
	// matching conversions in the same generated package.
	wireTransformHelperRecord struct {
		identity    wireTransformHelperIdentity
		declaration *codegen.NameDeclaration
		prefix      string
		preferred   string
		order       wireNameOrder
	}

	// wireTransformHelperIdentity contains the generated source and target Go
	// types plus the nil behavior of one conversion function.
	wireTransformHelperIdentity struct {
		source   wireTransformTypeIdentity
		target   wireTransformTypeIdentity
		required bool
	}

	// wireTransformTypeIdentity selects either one HTTP package declaration or
	// one service type and records the field layout used by its generated code.
	wireTransformTypeIdentity struct {
		wire           *wireTypeRecord
		origin         expr.UserType
		attribute      *expr.AttributeExpr
		layout         codegen.GoLayoutPolicy
		servicePackage codegen.ImportSpec
	}

	// wireUnionRecord stores one generated union and the Go names used for its
	// type, branches, constants, and functions.
	wireUnionRecord struct {
		identity     wireUnionIdentity
		attribute    *expr.AttributeExpr
		declaration  *codegen.NameDeclaration
		kind         *codegen.NameDeclaration
		kindDecls    []*codegen.NameDeclaration
		ctorDecls    []*codegen.NameDeclaration
		name         string
		kindName     string
		kindConsts   []string
		constructors []string
		storageNames []string
		data         *service.UnionTypeData
	}

	// wireUnionOccurrence stores one copied union until Goa assigns names to its branches.
	wireUnionOccurrence struct {
		attribute *expr.AttributeExpr
		use       wireUnionUse
		role      wireTypeRole
		policy    wireTypePolicy
	}

	// wireUnionUse identifies the HTTP body role that owns a union declaration.
	wireUnionUse struct {
		role wireTypeRole
		view string
	}

	// wireUnionOwner identifies one authored OneOf in one HTTP body role and
	// response view.
	wireUnionOwner struct {
		authored *expr.AttributeExpr
		use      wireUnionUse
	}

	// wireValidationRoot stores one HTTP value whose generated code runs
	// validation directly instead of calling a named validator.
	wireValidationRoot struct {
		attribute *expr.AttributeExpr
		policy    wireTypePolicy
	}

	// wireUnionIdentity records which OneOf was written, where HTTP uses it, and
	// the Go type used by each branch.
	wireUnionIdentity struct {
		declaration  codegen.UnionDeclarationID
		owner        wireUnionOwner
		declarations []*wireTypeRecord
	}

	// wireTypeRecord stores one generated type and its optional functions.
	wireTypeRecord struct {
		identity         wireTypeIdentity
		declaration      *codegen.NameDeclaration
		validator        *codegen.NameDeclaration
		nestedValidator  *codegen.NameDeclaration
		constructor      *codegen.NameDeclaration
		needsValidator   bool
		needsNestedCall  bool
		needsConstructor bool
		name             string
		ref              string
		data             *TypeData
		errorUses        []wireErrorUse
		releasedNames    []string
	}

	// wireErrorUse records one service error whose HTTP body uses a generated
	// type declaration.
	wireErrorUse struct {
		service string
		method  string
		name    string
	}

	// wireTypeIdentity contains a designed type and the rules that change its Go definition.
	wireTypeIdentity struct {
		api       string
		sourceID  string
		resultID  string
		role      wireTypeRole
		preferred string
		attribute *expr.AttributeExpr
		policy    wireTypePolicy
	}

	// wireTypePolicy records how one copied type represents fields, pointers,
	// default values, validation, and result views.
	wireTypePolicy struct {
		request             bool
		pointer             bool
		useDefault          bool
		validate            bool
		arrayElementPointer bool
		view                string
	}

	// wireTypeRole says whether an unnamed designed type is used for a request,
	// response, field, or stream value.
	wireTypeRole uint8

	// wireTransformSide identifies which value in a conversion is declared in
	// the generated HTTP package.
	wireTransformSide uint8

	// wireNameKind identifies the declaration being ordered for a generated
	// package name.
	wireNameKind uint8

	// wireAttributePair remembers two attributes already compared so values that
	// refer back to themselves do not cause an endless loop.
	wireAttributePair struct {
		left  *expr.AttributeExpr
		right *expr.AttributeExpr
	}

	// wireNameOrder contains designed values used to choose stable suffixes when
	// several declarations ask for the same Go name.
	wireNameOrder struct {
		kind                wireNameKind
		api                 string
		source              string
		target              string
		role                uint8
		preferred           string
		shape               string
		view                string
		request             bool
		pointer             bool
		arrayElementPointer bool
		defaults            bool
		required            bool
	}

	// wireAttributeScope chooses the Go type name for each copied HTTP field.
	wireAttributeScope struct {
		catalog         *wireTypeCatalog
		base            codegen.Attributor
		pkg             string
		policy          wireTypePolicy
		use             wireUnionUse
		viewRoot        *wireTypeRecord
		exactOccurrence bool
	}
)

const (
	wireRequestBody wireTypeRole = iota + 1
	wireResponseBody
	wireAttribute
	wireStreamPayload
)

const (
	wireTransformSource wireTransformSide = iota + 1
	wireTransformTarget
)

const (
	wireNameConstructor wireNameKind = iota + 1
	wireNameNestedValidator
	wireNameTransformHelper
	wireNameType
	wireNameValidator
)

// newWireTypeCatalog creates the type list for one generated Go package. Tests
// may omit the package when they only compare copied attributes.
func newWireTypeCatalog(pkg ...*codegen.GeneratedPackage) *wireTypeCatalog {
	catalog := &wireTypeCatalog{
		bindings:          make(map[*expr.AttributeExpr]*wireTypeRecord),
		unionBindings:     make(map[*expr.Union]*wireUnionRecord),
		transformBindings: make(map[codegen.TransformHelperID]*wireTransformHelperRecord),
		releasedSuffixes:  make(map[*expr.AttributeExpr]string),
	}
	if len(pkg) > 0 {
		catalog.pkg = pkg[0]
	}
	return catalog
}

// collect records attribute and every named type it contains.
func (c *wireTypeCatalog) collect(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, api ...string) *wireTypeRecord {
	return c.collectWithReleasedNames(attribute, role, policy, "", nil, api...)
}

// collectWithReleasedNames records a response while keeping the public names
// produced before view selection moved into this package.
func (c *wireTypeCatalog) collectWithReleasedNames(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string, releasedNames map[expr.UserType]string, api ...string) *wireTypeRecord {
	if c.declared {
		panic("cannot collect an HTTP type after its package declarations are submitted")
	}
	suffix := releasedWireTypeSuffix(attribute, role)
	c.releasedSuffixes[attribute] = suffix
	root := ""
	if len(api) > 0 {
		root = api[0]
	}
	use := wireUnionUse{role: role, view: policy.view}
	return c.collectRecursive(attribute, role, policy, use, preferred, suffix, root, true, releasedNames, make(map[expr.UserType]struct{}))
}

// collectChildren records named types inside attribute without recording its
// top-level named type a second time.
func (c *wireTypeCatalog) collectChildren(attribute *expr.AttributeExpr, use wireUnionUse, policy wireTypePolicy, api ...string) {
	c.collectChildrenWithReleasedNames(attribute, use, policy, nil, api...)
}

// collectChildrenWithReleasedNames records response fields with their released
// names when selecting a view changed the order of name suffixes.
func (c *wireTypeCatalog) collectChildrenWithReleasedNames(attribute *expr.AttributeExpr, use wireUnionUse, policy wireTypePolicy, releasedNames map[expr.UserType]string, api ...string) {
	suffix := c.releasedSuffixes[attribute]
	root := ""
	if len(api) > 0 {
		root = api[0]
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		c.collectRecursive(userType.Attribute(), wireAttribute, policy, use, "", suffix, root, false, releasedNames, make(map[expr.UserType]struct{}))
		return
	}
	c.collectRecursive(attribute, wireAttribute, policy, use, "", suffix, root, false, releasedNames, make(map[expr.UserType]struct{}))
}

// Declare requests every type and function name this HTTP package can write.
// The caller invokes it before Goa chooses names so every file writing to the
// same package can resolve conflicts together.
func (c *wireTypeCatalog) Declare() error {
	if c.declared {
		return nil
	}
	if c.pkg == nil {
		return fmt.Errorf("HTTP type declarations require a generated package")
	}
	c.planNestedValidators()
	for _, record := range c.records {
		record.declaration = codegen.NewPreferredName(
			codegen.NameType,
			record.preferredName(),
			codegen.ExportedName,
			record.identity.order(wireNameType),
		)
		if err := c.pkg.DeclareName(record.declaration); err != nil {
			return err
		}
		if record.needsValidator {
			declaration, err := c.pkg.DeclareDependentName(
				codegen.NameFunction,
				record.declaration,
				"Validate",
				"",
				record.identity.order(wireNameValidator),
			)
			if err != nil {
				return err
			}
			record.validator = declaration
			if record.needsNestedCall {
				declaration, err = c.pkg.DeclareDependentName(
					codegen.NameFunction,
					record.declaration,
					"validate",
					"",
					record.identity.order(wireNameNestedValidator),
				)
				if err != nil {
					return err
				}
				record.nestedValidator = declaration
			}
		}
		if record.needsConstructor {
			declaration, err := c.pkg.DeclareDependentName(
				codegen.NameFunction,
				record.declaration,
				"New",
				"",
				record.identity.order(wireNameConstructor),
			)
			if err != nil {
				return err
			}
			record.constructor = declaration
		}
	}
	for _, occurrence := range c.unionOccurrences {
		if err := validateHTTPUnionBranchGoNames(occurrence.attribute.Type.(*expr.Union)); err != nil {
			return err
		}
		identity := c.unionIdentity(occurrence.attribute, occurrence.use, occurrence.role, occurrence.policy)
		if record := c.findUnion(identity); record != nil {
			continue
		}
		if record := c.findUnionOwner(identity.owner); record != nil {
			return fmt.Errorf(
				"HTTP OneOf %q produces different %s definitions; use separate OneOf declarations",
				occurrence.attribute.Type.Name(),
				occurrence.use.description(),
			)
		}
		c.unions = append(c.unions, &wireUnionRecord{identity: identity, attribute: occurrence.attribute})
	}
	for _, union := range c.unions {
		actual := union.attribute.Type.(*expr.Union)
		name := union.identity.owner.use.nameFor(actual)
		union.declaration = codegen.NewExactName(codegen.NameType, name)
		if err := c.declareUnionName(actual, union.identity.owner.use, union.declaration); err != nil {
			return err
		}
		kind := codegen.NewExactName(codegen.NameType, name+"Kind")
		if err := c.declareUnionName(actual, union.identity.owner.use, kind); err != nil {
			return err
		}
		union.kind = kind
		union.kindDecls = make([]*codegen.NameDeclaration, len(actual.Values))
		union.ctorDecls = make([]*codegen.NameDeclaration, len(actual.Values))
		union.storageNames = make([]string, len(actual.Values))
		storageNames := codegen.NewNameScope()
		// Reserve kind for the selector so a branch named kind uses another field.
		storageNames.Unique("kind")
		for index, branch := range actual.Values {
			branchName := codegen.Goify(branch.Name, true)
			kindDeclaration := codegen.NewExactName(codegen.NameConstant, name+"Kind"+branchName)
			if err := c.declareUnionName(actual, union.identity.owner.use, kindDeclaration); err != nil {
				return err
			}
			constructor := codegen.NewExactName(codegen.NameFunction, "New"+name+branchName)
			if err := c.declareUnionName(actual, union.identity.owner.use, constructor); err != nil {
				return err
			}
			union.kindDecls[index] = kindDeclaration
			union.ctorDecls[index] = constructor
			union.storageNames[index] = storageNames.Unique(codegen.Goify(branch.Name, false))
		}
	}
	for _, transform := range c.transforms {
		for _, helper := range transform.plan.Helpers() {
			identity, err := c.transformHelperIdentity(transform, helper)
			if err != nil {
				return err
			}
			preferred, err := c.transformHelperPreferredName(transform.prefix, identity)
			if err != nil {
				return err
			}
			order := wireNameOrder{
				kind:      wireNameTransformHelper,
				source:    identity.source.orderKey(),
				target:    identity.target.orderKey(),
				preferred: preferred,
				required:  identity.required,
			}
			record := c.findTransformHelper(identity)
			if record == nil {
				record = &wireTransformHelperRecord{
					identity:  identity,
					prefix:    transform.prefix,
					preferred: preferred,
					order:     order,
				}
				c.transformHelpers = append(c.transformHelpers, record)
			} else if order.ComparePackageName(record.order) < 0 {
				record.prefix = transform.prefix
				record.preferred = preferred
				record.order = order
			}
			c.transformBindings[helper.ID] = record
		}
	}
	for _, helper := range c.transformHelpers {
		declaration, err := c.declareTransformHelper(helper)
		if err != nil {
			return err
		}
		helper.declaration = declaration
	}
	for _, transform := range c.transforms {
		for _, planned := range transform.plan.Helpers() {
			helper := c.transformBindings[planned.ID]
			if helper == nil {
				return fmt.Errorf("HTTP conversion function declaration was not recorded")
			}
			if err := transform.plan.BindHelperDeclaration(planned.ID, helper.declaration); err != nil {
				return err
			}
		}
	}
	c.declared = true
	return nil
}

// collectTransform records one request or response conversion before Goa
// chooses the names of any extra conversion functions. The returned handle
// selects this record when the caller later writes the conversion.
func (c *wireTypeCatalog) collectTransform(source, target *expr.AttributeExpr, prefix, owner string, layout wireTransformLayout) wireTransformHandle {
	if c.declared {
		panic("cannot collect an HTTP conversion after package declarations are submitted")
	}
	source = expr.DupAtt(source)
	target = expr.DupAtt(target)
	plan, err := codegen.NewTransformPlan(source, target, "", nil)
	if err != nil {
		panic(err)
	}
	record := &wireTransformRecord{
		source: source,
		target: target,
		prefix: prefix,
		owner:  owner,
		layout: layout,
		plan:   plan,
	}
	c.transforms = append(c.transforms, record)
	return wireTransformHandle{catalog: c, record: record}
}

// transformHelperIdentity resolves the exact generated declarations and field
// layouts used by one planned conversion function.
func (c *wireTypeCatalog) transformHelperIdentity(transform *wireTransformRecord, helper codegen.TransformHelper) (wireTransformHelperIdentity, error) {
	sourceWire := transform.layout.wireSide == wireTransformSource
	targetWire := transform.layout.wireSide == wireTransformTarget
	source, err := c.transformTypeIdentity(helper.Source, sourceWire, transform.layout.wirePolicy, transform.layout.servicePointer, transform.layout.servicePackage)
	if err != nil {
		return wireTransformHelperIdentity{}, err
	}
	target, err := c.transformTypeIdentity(helper.Target, targetWire, transform.layout.wirePolicy, transform.layout.servicePointer, transform.layout.servicePackage)
	if err != nil {
		return wireTransformHelperIdentity{}, err
	}
	return wireTransformHelperIdentity{
		source:   source,
		target:   target,
		required: helper.Required,
	}, nil
}

// transformTypeIdentity returns the declaration and field rules that determine
// a generated conversion function's parameter or result type.
func (c *wireTypeCatalog) transformTypeIdentity(attribute *expr.AttributeExpr, wire bool, policy wireTypePolicy, servicePointer bool, servicePackage codegen.ImportSpec) (wireTransformTypeIdentity, error) {
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return wireTransformTypeIdentity{}, fmt.Errorf("HTTP conversion function type %q is not named", attribute.Type.Name())
	}
	if !wire {
		if location := codegen.UserTypeLocation(userType); location != nil {
			servicePackage = codegen.ImportSpec{Name: location.PackageName(), Path: location.RelImportPath}
		}
		if servicePackage.Name == "" || servicePackage.Path == "" {
			return wireTransformTypeIdentity{}, fmt.Errorf("HTTP conversion function type %q has no service package", userType.Name())
		}
		return wireTransformTypeIdentity{
			origin:         userType.Origin(),
			attribute:      attribute,
			servicePackage: servicePackage,
			layout: codegen.GoLayoutPolicy{
				Pointer:    servicePointer,
				UseDefault: true,
				SumType:    true,
			},
		}, nil
	}
	policy.view = ""
	preferred := wireTypePreferredName(userType, policy)
	record := c.find(newWireTypeIdentity(attribute, wireAttribute, policy, preferred))
	if record == nil {
		return wireTransformTypeIdentity{}, fmt.Errorf("HTTP conversion function type %q was not recorded", preferred)
	}
	return wireTransformTypeIdentity{
		wire:      record,
		attribute: attribute,
		layout: codegen.GoLayoutPolicy{
			Pointer:             policy.pointer,
			UseDefault:          policy.useDefault,
			UnionPointer:        true,
			ArrayElementPointer: policy.arrayElementPointer,
			SumType:             true,
		},
	}, nil
}

// transformHelperPreferredName describes both generated types converted by
// one helper before their package selects final declaration names.
func (c *wireTypeCatalog) transformHelperPreferredName(prefix string, identity wireTransformHelperIdentity) (string, error) {
	source, err := c.transformTypeRoleName(identity.source)
	if err != nil {
		return "", err
	}
	target, err := c.transformTypeRoleName(identity.target)
	if err != nil {
		return "", err
	}
	return prefix + codegen.Goify(source, true) + "To" + codegen.Goify(target, true) + identity.behaviorSuffix(), nil
}

// transformTypeRoleName returns the package and type role used in a helper
// name. Wire values use their planned request or response declaration.
func (c *wireTypeCatalog) transformTypeRoleName(identity wireTransformTypeIdentity) (string, error) {
	if identity.wire != nil {
		return identity.wire.preferredName(), nil
	}
	userType, ok := identity.attribute.Type.(expr.UserType)
	if !ok {
		return "", fmt.Errorf("HTTP conversion function type %q is not named", identity.attribute.Type.Name())
	}
	typeName := codegen.Goify(wireTypeDeclaredName(userType), true)
	return identity.servicePackage.Name + typeName, nil
}

// declareTransformHelper makes the function name depend on the wire type it
// converts. If that type receives a suffix, the helper receives it too.
func (c *wireTypeCatalog) declareTransformHelper(helper *wireTransformHelperRecord) (*codegen.NameDeclaration, error) {
	sourceWire := helper.identity.source.wire
	targetWire := helper.identity.target.wire
	if (sourceWire == nil) == (targetWire == nil) {
		return nil, fmt.Errorf("HTTP conversion function must convert between one wire type and one service type")
	}
	if sourceWire != nil {
		target, err := c.transformTypeRoleName(helper.identity.target)
		if err != nil {
			return nil, err
		}
		return c.pkg.DeclareDependentName(
			codegen.NameFunction,
			sourceWire.declaration,
			helper.prefix,
			"To"+codegen.Goify(target, true)+helper.identity.behaviorSuffix(),
			helper.order,
		)
	}
	source, err := c.transformTypeRoleName(helper.identity.source)
	if err != nil {
		return nil, err
	}
	return c.pkg.DeclareDependentName(
		codegen.NameFunction,
		targetWire.declaration,
		helper.prefix+codegen.Goify(source, true)+"To",
		helper.identity.behaviorSuffix(),
		helper.order,
	)
}

// behaviorSuffix distinguishes a conversion that preserves a missing source
// value from one whose caller guarantees that the source is present.
func (i wireTransformHelperIdentity) behaviorSuffix() string {
	if !i.required {
		return "Optional"
	}
	return ""
}

// findTransformHelper returns the declaration for an equivalent conversion.
func (c *wireTypeCatalog) findTransformHelper(identity wireTransformHelperIdentity) *wireTransformHelperRecord {
	for _, record := range c.transformHelpers {
		if wireTransformHelperIdentitiesEqual(record.identity, identity) {
			return record
		}
	}
	return nil
}

// wireTransformHelperIdentitiesEqual reports whether two functions have the
// same generated parameter, result, field layout, and nil behavior.
func wireTransformHelperIdentitiesEqual(left, right wireTransformHelperIdentity) bool {
	return left.required == right.required &&
		wireTransformTypeIdentitiesEqual(left.source, right.source) &&
		wireTransformTypeIdentitiesEqual(left.target, right.target)
}

// wireTransformTypeIdentitiesEqual compares exact generated declarations and
// the concrete fields written inside service types.
func wireTransformTypeIdentitiesEqual(left, right wireTransformTypeIdentity) bool {
	if left.wire != right.wire || left.origin != right.origin || left.layout != right.layout || left.servicePackage != right.servicePackage {
		return false
	}
	if left.wire != nil {
		return true
	}
	leftType := left.attribute.Type.(expr.UserType)
	rightType := right.attribute.Type.(expr.UserType)
	return wireAttributesEqual(leftType.Attribute(), rightType.Attribute(), make(map[wireAttributePair]struct{}))
}

// orderKey describes every generated fact that changes a conversion
// function's parameter or result type.
func (i wireTransformTypeIdentity) orderKey() string {
	if i.wire != nil {
		order := i.wire.identity.order(wireNameType)
		return fmt.Sprintf(
			"wire:%q:%d:%q:%q:%q:%t:%t:%t:%t",
			order.source,
			order.role,
			order.preferred,
			order.shape,
			order.view,
			order.request,
			order.pointer,
			order.arrayElementPointer,
			order.defaults,
		)
	}
	return fmt.Sprintf(
		"service:%q:%q:%q:%q:%t:%t:%t:%t:%t",
		i.servicePackage.Path,
		i.servicePackage.Name,
		wireTypeDeclaredName(i.origin),
		expr.Hash(i.attribute.Type, false, false, false),
		i.layout.Pointer,
		i.layout.IgnoreRequired,
		i.layout.UseDefault,
		i.layout.UnionPointer,
		i.layout.ArrayElementPointer,
	)
}

// renderTransform writes the conversion selected when wire types were
// collected. The handle prevents two structurally identical conversions from
// being exchanged when callers render them in a different order.
func (c *wireTypeCatalog) renderTransform(
	handle wireTransformHandle,
	wireAttribute *expr.AttributeExpr,
	sourceVar, targetVar string,
	sourceContext, targetContext *codegen.AttributeContext,
) (string, []*codegen.TransformFunctionData, error) {
	if handle.catalog != c || handle.record == nil {
		return "", nil, fmt.Errorf("HTTP conversion handle belongs to a different generated package")
	}
	transform := handle.record
	if transform.used {
		return "", nil, fmt.Errorf("HTTP %s conversion for %s was already rendered", transform.prefix, transform.owner)
	}
	if err := transform.plan.BindContexts(sourceContext, targetContext); err != nil {
		return "", nil, err
	}
	if transform.layout.wireSide == wireTransformSource {
		if err := c.bindTransformOccurrence(transform.source, wireAttribute, sourceContext, transform.layout.wireUse); err != nil {
			return "", nil, err
		}
	} else {
		if err := c.bindTransformOccurrence(transform.target, wireAttribute, targetContext, transform.layout.wireUse); err != nil {
			return "", nil, err
		}
	}
	for _, helper := range transform.plan.Helpers() {
		c.bindTransformHelper(helper.Source, sourceContext, transform.layout.wireUse)
		c.bindTransformHelper(helper.Target, targetContext, transform.layout.wireUse)
	}
	code, helpers, err := transform.plan.Render(sourceVar, targetVar, true)
	if err != nil {
		return "", nil, err
	}
	if err := c.retainTransformDefinitions(helpers); err != nil {
		return "", nil, err
	}
	transform.used = true
	return code, helpers, nil
}

// checkTransformUsed rejects a planned conversion that was never written. A
// generated package may contain records from several transport plans, so the
// caller checks only handles owned by the plan currently linking.
func (c *wireTypeCatalog) checkTransformUsed(handle wireTransformHandle) error {
	if handle.record == nil {
		return nil
	}
	if handle.catalog != c {
		return fmt.Errorf("HTTP conversion handle belongs to a different generated package")
	}
	if !handle.record.used {
		return fmt.Errorf("HTTP %s conversion for %s was planned but not rendered", handle.record.prefix, handle.record.owner)
	}
	return nil
}

// retainTransformDefinitions verifies that every independently planned use of
// one package function has the same parameter type, result type, and body.
func (c *wireTypeCatalog) retainTransformDefinitions(helpers []*codegen.TransformFunctionData) error {
	if c.transformDefinitions == nil {
		c.transformDefinitions = make(map[*codegen.NameDeclaration]*codegen.TransformFunctionData)
	}
	pending := make(map[*codegen.NameDeclaration]*codegen.TransformFunctionData)
	for _, helper := range helpers {
		previous := c.transformDefinitions[helper.Declaration]
		if previous == nil {
			previous = pending[helper.Declaration]
		}
		if previous != nil && !wireTransformDefinitionsEqual(previous, helper) {
			return fmt.Errorf("HTTP transform helper declaration %q has different definitions", helper.Declaration.Name())
		}
		pending[helper.Declaration] = helper
	}
	for declaration, helper := range pending {
		c.transformDefinitions[declaration] = helper
	}
	return nil
}

// wireTransformDefinitionsEqual reports whether two planned conversions emit
// the same package-level function.
func wireTransformDefinitionsEqual(left, right *codegen.TransformFunctionData) bool {
	return left.ParamTypeRef == right.ParamTypeRef &&
		left.ResultTypeRef == right.ResultTypeRef &&
		left.Code == right.Code
}

// bindTransformHelper gives a nested copied field the same Go type name used by
// its generated conversion function. Service values already receive names from
// the service generator.
func (c *wireTypeCatalog) bindTransformHelper(attribute *expr.AttributeExpr, context *codegen.AttributeContext, use wireUnionUse) {
	scope, ok := context.Scope.(*wireAttributeScope)
	if !ok || scope.catalog != c {
		return
	}
	policy := scope.policy
	policy.view = ""
	c.applyNamesRecursive(attribute, wireAttribute, policy, use, make(map[expr.UserType]struct{}))
}

// bindTransformOccurrence records the Go type used by one copied conversion
// value and each named field inside it.
func (c *wireTypeCatalog) bindTransformOccurrence(planned, rendered *expr.AttributeExpr, context *codegen.AttributeContext, use wireUnionUse) error {
	scope, ok := context.Scope.(*wireAttributeScope)
	if !ok || scope.catalog != c {
		return nil
	}
	// The transform planner works on its own copy. Carry each union's planned
	// declaration from the rendered HTTP body to that copy before rendering.
	if err := c.bindCopiedUnionOccurrences(planned, rendered, "body", make(map[wireAttributePair]struct{})); err != nil {
		return err
	}
	record := c.bindings[rendered]
	if record == nil {
		c.applyNamesRecursive(planned, wireAttribute, scope.policy, use, make(map[expr.UserType]struct{}))
		return nil
	}
	c.bindOccurrence(planned, record)
	c.applyNamesRecursive(planned, record.identity.role, record.identity.policy, use, make(map[expr.UserType]struct{}))
	return nil
}

// bindCopiedUnionOccurrences gives unions in a transform copy the declarations
// already selected for the matching fields in the rendered HTTP body. It
// rejects a mismatch because a declaration from another field would generate
// a conversion with the wrong public type.
func (c *wireTypeCatalog) bindCopiedUnionOccurrences(planned, rendered *expr.AttributeExpr, path string, seen map[wireAttributePair]struct{}) error {
	if planned == nil || rendered == nil {
		switch {
		case planned == nil && rendered == nil:
			return nil
		case planned == nil:
			return fmt.Errorf("planned attribute is missing at %s", path)
		default:
			return fmt.Errorf("rendered attribute is missing at %s", path)
		}
	}
	if planned.Type == expr.Empty || rendered.Type == expr.Empty {
		if planned.Type == rendered.Type {
			return nil
		}
		return copiedUnionTypeMismatch(planned, rendered, path)
	}
	pair := wireAttributePair{left: planned, right: rendered}
	if _, ok := seen[pair]; ok {
		return nil
	}
	seen[pair] = struct{}{}

	if plannedType, ok := planned.Type.(expr.UserType); ok {
		renderedType, ok := rendered.Type.(expr.UserType)
		if !ok {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		if plannedType.Name() != renderedType.Name() {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		return c.bindCopiedUnionOccurrences(plannedType.Attribute(), renderedType.Attribute(), path, seen)
	}
	if _, ok := rendered.Type.(expr.UserType); ok {
		return copiedUnionTypeMismatch(planned, rendered, path)
	}
	switch plannedType := planned.Type.(type) {
	case *expr.Object:
		renderedType, ok := rendered.Type.(*expr.Object)
		if !ok {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		for _, field := range *plannedType {
			renderedField := renderedType.Attribute(field.Name)
			fieldPath := path + "." + field.Name
			if renderedField == nil {
				return fmt.Errorf("rendered object has no field %q at %s", field.Name, fieldPath)
			}
			if err := c.bindCopiedUnionOccurrences(field.Attribute, renderedField, fieldPath, seen); err != nil {
				return err
			}
		}
		for _, field := range *renderedType {
			if plannedType.Attribute(field.Name) == nil {
				return fmt.Errorf("planned object has no field %q at %s.%s", field.Name, path, field.Name)
			}
		}
	case *expr.Array:
		renderedType, ok := rendered.Type.(*expr.Array)
		if !ok {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		return c.bindCopiedUnionOccurrences(plannedType.ElemType, renderedType.ElemType, path+"[]", seen)
	case *expr.Map:
		renderedType, ok := rendered.Type.(*expr.Map)
		if !ok {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		if err := c.bindCopiedUnionOccurrences(plannedType.KeyType, renderedType.KeyType, path+" key", seen); err != nil {
			return err
		}
		return c.bindCopiedUnionOccurrences(plannedType.ElemType, renderedType.ElemType, path+" value", seen)
	case *expr.Union:
		renderedType, ok := rendered.Type.(*expr.Union)
		if !ok {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		if plannedType.Name() != renderedType.Name() || plannedType.TypeKey != renderedType.TypeKey || plannedType.ValueKey != renderedType.ValueKey {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
		if record := c.unionBindings[renderedType]; record != nil {
			c.unionBindings[plannedType] = record
		}
		for _, branch := range plannedType.Values {
			renderedBranch := unionBranchAttribute(renderedType, branch.Name)
			branchPath := path + "." + branch.Name
			if renderedBranch == nil {
				return fmt.Errorf("rendered OneOf %q has no branch %q at %s", renderedType.Name(), branch.Name, branchPath)
			}
			if err := c.bindCopiedUnionOccurrences(branch.Attribute, renderedBranch, branchPath, seen); err != nil {
				return err
			}
		}
		for _, branch := range renderedType.Values {
			if unionBranchAttribute(plannedType, branch.Name) == nil {
				return fmt.Errorf("planned OneOf %q has no branch %q at %s.%s", plannedType.Name(), branch.Name, path, branch.Name)
			}
		}
	default:
		if planned.Type != rendered.Type {
			return copiedUnionTypeMismatch(planned, rendered, path)
		}
	}
	return nil
}

// copiedUnionTypeMismatch describes two expression nodes that cannot represent
// the same field in a planned and rendered HTTP conversion.
func copiedUnionTypeMismatch(planned, rendered *expr.AttributeExpr, path string) error {
	return fmt.Errorf(
		"planned %s does not match rendered %s at %s",
		copiedUnionTypeName(planned.Type),
		copiedUnionTypeName(rendered.Type),
		path,
	)
}

// copiedUnionTypeName returns a short concrete name for a mismatch diagnostic.
func copiedUnionTypeName(dataType expr.DataType) string {
	if dataType == expr.Empty {
		return "empty type"
	}
	if userType, ok := dataType.(expr.UserType); ok {
		return fmt.Sprintf("user type %q", userType.Name())
	}
	if union, ok := dataType.(*expr.Union); ok {
		return fmt.Sprintf("OneOf %q", union.Name())
	}
	return dataType.Name()
}

// unionBranchAttribute returns the branch with name, or nil when two transform
// copies do not have the same fields.
func unionBranchAttribute(union *expr.Union, name string) *expr.AttributeExpr {
	for _, branch := range union.Values {
		if branch.Name == name {
			return branch.Attribute
		}
	}
	return nil
}

// Link reads the assigned package names and builds the type definitions,
// references, unions, and validation functions written to files.
func (c *wireTypeCatalog) Link() {
	if c.linked {
		return
	}
	if !c.declared {
		panic("cannot link HTTP types before declaring their package names")
	}
	c.scope = c.pkg.Scope()
	for _, union := range c.unions {
		union.name = union.declaration.Name()
		union.kindName = union.kind.Name()
		union.kindConsts = make([]string, len(union.kindDecls))
		union.constructors = make([]string, len(union.ctorDecls))
		for index := range union.kindDecls {
			union.kindConsts[index] = union.kindDecls[index].Name()
			union.constructors[index] = union.ctorDecls[index].Name()
		}
		c.applyUnionRecord(union.attribute, union)
	}
	for _, record := range c.records {
		record.name = record.declaration.Name()
		use := wireUnionUse{
			role: record.identity.role,
			view: record.identity.policy.view,
		}
		resolver := c.rootResolver(c.scope, record.identity.policy, use, record)
		layout, err := resolver.(*wireAttributeScope).goTypeLayout(
			record.identity.attribute,
			wireGoLayoutPolicy(record.identity.policy),
			true,
		)
		if err != nil {
			panic(fmt.Sprintf("link HTTP type %q: %v", record.name, err))
		}
		record.ref = layout.Ref()
	}
	for _, union := range c.unions {
		actual := union.attribute.Type.(*expr.Union)
		union.data = buildHTTPUnionTypeData(actual, c.occurrenceResolver(c.scope, union.identity.owner.use), union)
	}
	c.linked = true
}

// optionalWireTypeRef returns the exact transport type used to preserve an
// optional top-level value. The booleans report whether command-line decoding
// must preserve absence and whether conversion must dereference an added
// pointer.
func optionalWireTypeRef(layout codegen.LinkedGoType, optional bool) (typeRef string, preserveAbsence, dereference bool) {
	if optional && !layout.ReferenceCanBeNil() {
		return layout.RefWithPointer(true), true, layout.Kind() != codegen.GoStruct
	}
	return layout.Ref(), optional && layout.ReferenceIsPointer(), false
}

// wireGoLayoutPolicy converts the HTTP package choices into the shared Go type
// plan used by definitions, references, and default values.
func wireGoLayoutPolicy(policy wireTypePolicy) codegen.GoLayoutPolicy {
	return codegen.GoLayoutPolicy{
		Pointer:             policy.pointer,
		UseDefault:          policy.useDefault,
		UnionPointer:        true,
		ArrayElementPointer: policy.arrayElementPointer,
		SumType:             true,
	}
}

// lookup returns the chosen Go names for an equivalent copied type. It panics
// when the type was not recorded before names were assigned.
func (c *wireTypeCatalog) lookup(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) *wireTypeRecord {
	if !c.linked {
		panic("cannot resolve an HTTP type before its generated package freezes")
	}
	identity := newWireTypeIdentity(attribute, role, policy, preferred)
	record := c.find(identity)
	if record != nil {
		c.bindOccurrence(attribute, record)
		return record
	}
	panic(fmt.Sprintf("HTTP type %q was not submitted before package names were assigned", preferred))
}

// lookupUser returns the generated name information for a named design type.
// It returns nil for inline and primitive values because they define no named
// type at the top level.
func (c *wireTypeCatalog) lookupUser(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy) *wireTypeRecord {
	if attribute.Type == expr.Empty {
		return nil
	}
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return nil
	}
	return c.lookup(attribute, role, policy, wireTypePreferredName(userType, policy))
}

// applyNames associates every copied nested attribute with the Go type name
// used by its definition and conversions.
func (c *wireTypeCatalog) applyNames(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy) {
	use := wireUnionUse{role: role, view: policy.view}
	c.applyNamesRecursive(attribute, role, policy, use, make(map[expr.UserType]struct{}))
}

// applyNamesRecursive follows each named field once, including fields that
// refer back to an outer type.
func (c *wireTypeCatalog) applyNamesRecursive(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, use wireUnionUse, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		c.lookupUser(attribute, role, policy)
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.applyNamesRecursive(userType.Attribute(), wireAttribute, nestedPolicy, use, seen)
		delete(seen, origin)
		return
	}
	nestedPolicy := policy
	nestedPolicy.view = ""
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, named := range *actual {
			c.applyNamesRecursive(named.Attribute, wireAttribute, nestedPolicy, use, seen)
		}
	case *expr.Array:
		c.applyNamesRecursive(actual.ElemType, wireAttribute, nestedPolicy, use, seen)
	case *expr.Map:
		c.applyNamesRecursive(actual.KeyType, wireAttribute, nestedPolicy, use, seen)
		c.applyNamesRecursive(actual.ElemType, wireAttribute, nestedPolicy, use, seen)
	case *expr.Union:
		if record := c.unionBindings[actual]; record != nil {
			c.applyUnionRecord(attribute, record)
			return
		}
		identity := c.unionIdentity(attribute, use, role, policy)
		record := c.findUnion(identity)
		if record == nil {
			panic(fmt.Sprintf("HTTP union %q was not submitted before package names were assigned", actual.Name()))
		}
		c.applyUnionRecord(attribute, record)
	}
}

// unionTypes returns the generated union definitions in Go name order.
func (c *wireTypeCatalog) unionTypes() []*service.UnionTypeData {
	unions := make([]*service.UnionTypeData, len(c.unions))
	for index, record := range c.unions {
		unions[index] = record.data
	}
	slices.SortFunc(unions, func(left, right *service.UnionTypeData) int { return strings.Compare(left.Name, right.Name) })
	return unions
}

// bind associates the data used to write a type with its chosen Go name. When
// several equivalent copies need validation, it stores their shared validator.
func (c *wireTypeCatalog) bind(record *wireTypeRecord, data *TypeData) *TypeData {
	data.declaration = record
	data.Declaration = record.declaration
	data.VarName = record.declaration.Name()
	data.ValidatorDeclaration = record.validator
	data.NestedValidatorDeclaration = record.nestedValidator
	if record.validator != nil {
		data.ValidatorName = record.validator.Name()
	}
	if record.nestedValidator != nil {
		data.NestedValidatorName = record.nestedValidator.Name()
	}
	if data.Init != nil {
		data.Init.Declaration = record.constructor
		data.Init.Name = record.constructor.Name()
	}
	if description := record.errorDescription(); description != "" {
		data.Description = description
	}
	if record.data == nil {
		if data.Def == "" && data.ValidateDef == "" {
			return data
		}
		declaration := *data
		declaration.Init = nil
		record.data = &declaration
		return data
	}
	if data.Def != "" {
		if record.data.Def == "" {
			record.data.Def = data.Def
		} else if record.data.Def != data.Def {
			panic(fmt.Sprintf("HTTP type %q produced conflicting declarations", record.name))
		}
	}
	if data.ValidateDef != "" {
		if record.data.ValidateDef == "" {
			record.data.ValidateDef = data.ValidateDef
			record.data.ValidateRef = data.ValidateRef
			record.data.ValidatorName = data.ValidatorName
		} else if record.data.ValidateDef != data.ValidateDef || record.data.ValidateRef != data.ValidateRef {
			panic(fmt.Sprintf("HTTP type %q produced conflicting validators", record.name))
		}
	}
	if data.NestedValidateDef != "" {
		if record.data.NestedValidateDef == "" {
			record.data.NestedValidateDef = data.NestedValidateDef
			record.data.NestedValidatorName = data.NestedValidatorName
		} else if record.data.NestedValidateDef != data.NestedValidateDef {
			panic(fmt.Sprintf("HTTP type %q produced conflicting nested validators", record.name))
		}
	}
	return data
}

// addErrorUse records one error body role and keeps all roles in a stable
// order before generated source is written.
func (r *wireTypeRecord) addErrorUse(use wireErrorUse) {
	for _, existing := range r.errorUses {
		if existing == use {
			return
		}
	}
	r.errorUses = append(r.errorUses, use)
	slices.SortFunc(r.errorUses, func(left, right wireErrorUse) int {
		for _, compared := range []int{
			cmp.Compare(left.service, right.service),
			cmp.Compare(left.method, right.method),
			cmp.Compare(left.name, right.name),
		} {
			if compared != 0 {
				return compared
			}
		}
		return 0
	})
}

// errorDescription describes every designed error that uses the generated
// type. Two errors on one endpoint remain short enough for one sentence.
func (r *wireTypeRecord) errorDescription() string {
	if len(r.errorUses) == 0 {
		return ""
	}
	first := r.errorUses[0]
	if len(r.errorUses) == 1 {
		return fmt.Sprintf(
			"%s is the type of the %q service %q endpoint HTTP response body for the %q error.",
			r.name,
			first.service,
			first.method,
			first.name,
		)
	}
	if len(r.errorUses) == 2 && first.service == r.errorUses[1].service && first.method == r.errorUses[1].method {
		return fmt.Sprintf(
			"%s is the type of the %q service %q endpoint HTTP response body for the %q and %q errors.",
			r.name,
			first.service,
			first.method,
			first.name,
			r.errorUses[1].name,
		)
	}
	var description strings.Builder
	fmt.Fprintf(&description, "%s is the HTTP response body type for these service errors:", r.name)
	for _, use := range r.errorUses {
		fmt.Fprintf(
			&description,
			"\n- %q service %q endpoint: %q error",
			use.service,
			use.method,
			use.name,
		)
	}
	return description.String()
}

// collectRecursive records named types and stops when a type refers back to one
// it is already reading. Released names are kept separately from current type
// identity so several old declarations can share one current declaration.
func (c *wireTypeCatalog) collectRecursive(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, use wireUnionUse, preferred, releasedSuffix, api string, root bool, releasedNames map[expr.UserType]string, seen map[expr.UserType]struct{}) *wireTypeRecord {
	if attribute.Type == expr.Empty {
		return nil
	}
	var record *wireTypeRecord
	if userType, ok := attribute.Type.(expr.UserType); ok {
		preferred = wireTypePreferredName(userType, policy)
		identity := newWireTypeIdentity(attribute, role, policy, preferred)
		identity.api = api
		record = c.findOrAppend(identity)
		released := preferred
		if name := releasedNames[userType]; name != "" {
			released = name
		} else if !root {
			released += releasedSuffix
		}
		record.addReleasedName(released)
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return record
		}
		seen[origin] = struct{}{}
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectRecursive(userType.Attribute(), wireAttribute, nestedPolicy, use, "", releasedSuffix, api, false, releasedNames, seen)
		delete(seen, origin)
		return record
	}
	if preferred != "" {
		identity := newWireTypeIdentity(attribute, role, policy, preferred)
		identity.api = api
		record = c.findOrAppend(identity)
		record.addReleasedName(preferred)
	}
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		nestedPolicy := policy
		nestedPolicy.view = ""
		for _, named := range sortedWireAttributes(*actual) {
			c.collectRecursive(named.Attribute, wireAttribute, nestedPolicy, use, "", releasedSuffix, api, false, releasedNames, seen)
		}
	case *expr.Array:
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectRecursive(actual.ElemType, wireAttribute, nestedPolicy, use, "", releasedSuffix, api, false, releasedNames, seen)
	case *expr.Map:
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectRecursive(actual.KeyType, wireAttribute, nestedPolicy, use, "", releasedSuffix, api, false, releasedNames, seen)
		c.collectRecursive(actual.ElemType, wireAttribute, nestedPolicy, use, "", releasedSuffix, api, false, releasedNames, seen)
	case *expr.Union:
		nestedPolicy := policy
		nestedPolicy.view = ""
		unionAttribute := expr.DupAtt(attribute)
		c.unionOccurrences = append(c.unionOccurrences, wireUnionOccurrence{
			attribute: unionAttribute,
			use:       use,
			role:      role,
			policy:    policy,
		})
		for _, named := range actual.Values {
			c.collectRecursive(named.Attribute, wireAttribute, nestedPolicy, use, "", releasedSuffix, api, false, releasedNames, seen)
		}
	}
	return record
}

// addReleasedName records one spelling used before HTTP types were retained.
func (r *wireTypeRecord) addReleasedName(name string) {
	if name == "" || slices.Contains(r.releasedNames, name) {
		return
	}
	r.releasedNames = append(r.releasedNames, name)
	slices.Sort(r.releasedNames)
}

// preferredName keeps a released spelling only when it still names exactly one
// retained declaration. Shared declarations use their current designed name.
func (r *wireTypeRecord) preferredName() string {
	if len(r.releasedNames) == 1 {
		return r.releasedNames[0]
	}
	return r.identity.preferred
}

// findOrAppend reuses a record with the same generated type definition or adds
// a new record.
func (c *wireTypeCatalog) findOrAppend(identity wireTypeIdentity) *wireTypeRecord {
	if record := c.find(identity); record != nil {
		record.needsValidator = record.needsValidator || identity.policy.validate
		return record
	}
	record := &wireTypeRecord{
		identity:       identity,
		needsValidator: identity.policy.validate,
	}
	c.records = append(c.records, record)
	return record
}

// addValidationRoot records an inline HTTP value whose generated decoder or
// constructor runs validation.
func (c *wireTypeCatalog) addValidationRoot(attribute *expr.AttributeExpr, policy wireTypePolicy) {
	c.validationRoots = append(c.validationRoots, wireValidationRoot{
		attribute: expr.DupAtt(attribute),
		policy:    policy,
	})
}

// planNestedValidators marks the named validators called by generated public
// validators and inline validation code.
func (c *wireTypeCatalog) planNestedValidators() {
	for _, record := range c.records {
		record.needsNestedCall = false
	}
	for _, record := range c.records {
		if record.needsValidator {
			c.markNestedValidatorCalls(record.identity.attribute, record.identity.policy)
		}
	}
	for _, root := range c.validationRoots {
		c.markNestedValidatorCalls(root.attribute, root.policy)
	}
}

// markNestedValidatorCalls follows inline validation until it reaches a named
// type. A named type gets one private helper because the caller supplies its
// complete error path.
func (c *wireTypeCatalog) markNestedValidatorCalls(attribute *expr.AttributeExpr, policy wireTypePolicy) {
	if userType, ok := attribute.Type.(expr.UserType); ok && !expr.IsAlias(userType) {
		attribute = userType.Attribute()
	}
	policy.view = ""
	c.markInlineValidationCalls(attribute, policy, policy.pointer)
}

// markInlineValidationCalls records named calls inside one generated
// validation body. Anonymous values and aliases remain inside their caller.
func (c *wireTypeCatalog) markInlineValidationCalls(attribute *expr.AttributeExpr, policy wireTypePolicy, pointer bool) {
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if expr.IsAlias(userType) {
			c.markInlineValidationCalls(userType.Attribute(), policy, pointer)
			return
		}
		layout := codegen.GoLayoutPolicy{
			Pointer:             pointer,
			UseDefault:          policy.useDefault,
			UnionPointer:        true,
			ArrayElementPointer: policy.arrayElementPointer,
			SumType:             true,
		}
		if !codegen.NeedsValidation(userType.Attribute(), layout) {
			return
		}
		preferred := wireTypePreferredName(userType, policy)
		record := c.find(newWireTypeIdentity(attribute, wireAttribute, policy, preferred))
		if record == nil {
			panic(fmt.Sprintf("HTTP nested validator for %q was not collected", preferred))
		}
		if !record.needsValidator {
			panic(fmt.Sprintf("HTTP nested validator for %q has no public validator", record.name))
		}
		record.needsNestedCall = true
		return
	}

	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, field := range *actual {
			c.markInlineValidationCalls(field.Attribute, policy, pointer)
		}
	case *expr.Array:
		c.markInlineValidationCalls(actual.ElemType, policy, pointer)
	case *expr.Map:
		c.markInlineValidationCalls(actual.KeyType, policy, false)
		c.markInlineValidationCalls(actual.ElemType, policy, false)
	case *expr.Union:
		for _, branch := range actual.Values {
			branchPointer := pointer && expr.IsObject(branch.Attribute.Type)
			c.markInlineValidationCalls(branch.Attribute, policy, branchPointer)
		}
	}
}

// find returns the record for the same generated type definition.
func (c *wireTypeCatalog) find(identity wireTypeIdentity) *wireTypeRecord {
	for _, record := range c.records {
		if wireTypeIdentitiesEqual(record.identity, identity) {
			return record
		}
	}
	return nil
}

// unionIdentity returns the authored declaration, HTTP use, and Go type used
// by every named branch without changing attribute.
func (c *wireTypeCatalog) unionIdentity(attribute *expr.AttributeExpr, use wireUnionUse, role wireTypeRole, policy wireTypePolicy) wireUnionIdentity {
	identity := wireUnionIdentity{
		declaration: codegen.NewUnionDeclarationID(attribute),
		owner: wireUnionOwner{
			authored: attribute.AuthoredAttribute(),
			use:      use,
		},
	}
	c.collectUnionDeclarations(attribute, role, policy, &identity.declarations, make(map[expr.UserType]struct{}))
	return identity
}

// collectUnionDeclarations records generated branch types in branch order.
func (c *wireTypeCatalog) collectUnionDeclarations(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, declarations *[]*wireTypeRecord, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		preferred := wireTypePreferredName(userType, policy)
		record := c.find(newWireTypeIdentity(attribute, role, policy, preferred))
		if record == nil {
			panic(fmt.Sprintf("HTTP union branch type %q was not submitted before package names were assigned", preferred))
		}
		*declarations = append(*declarations, record)
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectUnionDeclarations(userType.Attribute(), wireAttribute, nestedPolicy, declarations, seen)
		delete(seen, origin)
		return
	}
	nestedPolicy := policy
	nestedPolicy.view = ""
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, named := range *actual {
			c.collectUnionDeclarations(named.Attribute, wireAttribute, nestedPolicy, declarations, seen)
		}
	case *expr.Array:
		c.collectUnionDeclarations(actual.ElemType, wireAttribute, nestedPolicy, declarations, seen)
	case *expr.Map:
		c.collectUnionDeclarations(actual.KeyType, wireAttribute, nestedPolicy, declarations, seen)
		c.collectUnionDeclarations(actual.ElemType, wireAttribute, nestedPolicy, declarations, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			c.collectUnionDeclarations(named.Attribute, wireAttribute, nestedPolicy, declarations, seen)
		}
	}
}

// applyUnionRecord gives one copied union the exact branch type names stored in
// record.
func (c *wireTypeCatalog) applyUnionRecord(attribute *expr.AttributeExpr, record *wireUnionRecord) {
	union := attribute.Type.(*expr.Union)
	c.unionBindings[union] = record
	index := 0
	seen := make(map[expr.UserType]struct{})
	for _, branch := range union.Values {
		c.applyResolvedDeclarations(branch.Attribute, record.identity.owner.use, record.identity.declarations, &index, seen)
	}
	if index != len(record.identity.declarations) {
		panic(fmt.Sprintf("HTTP union %q did not use every submitted branch name", record.name))
	}
}

// applyResolvedDeclarations gives each named branch its previously chosen Go
// type in the same order those types were recorded.
func (c *wireTypeCatalog) applyResolvedDeclarations(attribute *expr.AttributeExpr, use wireUnionUse, declarations []*wireTypeRecord, index *int, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if *index >= len(declarations) {
			panic(fmt.Sprintf("HTTP union branch %q has no submitted Go type name", wireTypeDeclaredName(userType)))
		}
		record := declarations[*index]
		(*index)++
		c.bindOccurrence(attribute, record)
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		c.applyResolvedDeclarations(userType.Attribute(), use, declarations, index, seen)
		delete(seen, origin)
		return
	}
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, named := range *actual {
			c.applyResolvedDeclarations(named.Attribute, use, declarations, index, seen)
		}
	case *expr.Array:
		c.applyResolvedDeclarations(actual.ElemType, use, declarations, index, seen)
	case *expr.Map:
		c.applyResolvedDeclarations(actual.KeyType, use, declarations, index, seen)
		c.applyResolvedDeclarations(actual.ElemType, use, declarations, index, seen)
	case *expr.Union:
		start := *index
		for _, branch := range actual.Values {
			c.applyResolvedDeclarations(branch.Attribute, use, declarations, index, seen)
		}
		identity := wireUnionIdentity{
			declaration: codegen.NewUnionDeclarationID(attribute),
			owner: wireUnionOwner{
				authored: attribute.AuthoredAttribute(),
				use:      use,
			},
			declarations: declarations[start:*index],
		}
		record := c.findUnion(identity)
		if record == nil {
			panic(fmt.Sprintf("HTTP nested union %q has no submitted Go type name", actual.Name()))
		}
		c.unionBindings[actual] = record
	}
}

// resolver returns the chosen HTTP type names. The supplied name list is used
// only to keep local variable names unique.
func (c *wireTypeCatalog) resolver(scope *codegen.NameScope, policy wireTypePolicy) codegen.Attributor {
	return &wireAttributeScope{catalog: c, base: codegen.NewAttributeScope(scope), policy: policy}
}

// resolverForUse names unions from copies made while rendering one exact HTTP
// body role and response view.
func (c *wireTypeCatalog) resolverForUse(scope *codegen.NameScope, policy wireTypePolicy, use wireUnionUse) codegen.Attributor {
	return &wireAttributeScope{catalog: c, base: codegen.NewAttributeScope(scope), policy: policy, use: use}
}

// occurrenceResolver uses only the exact type records attached to one copied
// expression. Union declarations use it after their branch records are fixed.
func (c *wireTypeCatalog) occurrenceResolver(scope *codegen.NameScope, use wireUnionUse) codegen.Attributor {
	return &wireAttributeScope{
		catalog:         c,
		base:            codegen.NewAttributeScope(scope),
		use:             use,
		exactOccurrence: true,
	}
}

// rootResolver applies a selected result view only to root. Nested fields use
// their normal type definitions without that view.
func (c *wireTypeCatalog) rootResolver(scope *codegen.NameScope, policy wireTypePolicy, use wireUnionUse, root *wireTypeRecord) codegen.Attributor {
	return &wireAttributeScope{catalog: c, base: codegen.NewAttributeScope(scope), policy: policy, use: use, viewRoot: root}
}

// bindOccurrence records the Go type used by one copied named value and its fields.
func (c *wireTypeCatalog) bindOccurrence(attribute *expr.AttributeExpr, record *wireTypeRecord) {
	c.bindings[attribute] = record
	if userType, ok := attribute.Type.(expr.UserType); ok {
		c.bindings[userType.Attribute()] = record
	}
}

// applyReleasedNames gives a copied composite body the public nested type
// names used to build its released constructor name.
func (c *wireTypeCatalog) applyReleasedNames(attribute *expr.AttributeExpr, policy wireTypePolicy, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		preferred := wireTypePreferredName(userType, policy)
		record := c.find(newWireTypeIdentity(attribute, wireAttribute, policy, preferred))
		if record == nil {
			panic(fmt.Sprintf("HTTP type %q was not submitted before its constructor name was built", preferred))
		}
		userType.Attribute().AddMeta("struct:type:name", record.preferredName())
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.applyReleasedNames(userType.Attribute(), nestedPolicy, seen)
		delete(seen, origin)
		return
	}
	nestedPolicy := policy
	nestedPolicy.view = ""
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, named := range *actual {
			c.applyReleasedNames(named.Attribute, nestedPolicy, seen)
		}
	case *expr.Array:
		c.applyReleasedNames(actual.ElemType, nestedPolicy, seen)
	case *expr.Map:
		c.applyReleasedNames(actual.KeyType, nestedPolicy, seen)
		c.applyReleasedNames(actual.ElemType, nestedPolicy, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			c.applyReleasedNames(named.Attribute, nestedPolicy, seen)
		}
	}
}

// releasedCompositeName returns the old public name for an array or map after
// applying its nested public type names.
func (c *wireTypeCatalog) releasedCompositeName(body *expr.AttributeExpr, policy wireTypePolicy) string {
	body = expr.DupAtt(body)
	c.applyReleasedNames(body, policy, make(map[expr.UserType]struct{}))
	name := codegen.NewAttributeScope(codegen.NewNameScope()).Name(body, "", policy.pointer, policy.useDefault)
	return codegen.Goify(name, true)
}

// releasedCompositeConstructorName returns the old public constructor name for
// an array or map body.
func (c *wireTypeCatalog) releasedCompositeConstructorName(body *expr.AttributeExpr, policy wireTypePolicy) string {
	return "New" + c.releasedCompositeName(body, policy)
}

// Name returns the type name selected for this HTTP attribute copy.
func (s *wireAttributeScope) Name(attribute *expr.AttributeExpr, pkg string, pointer, useDefault bool) string {
	if record := s.record(attribute); record != nil {
		if pkg == "" {
			return record.name
		}
		return pkg + "." + record.name
	}
	if _, ok := attribute.Type.(*expr.Union); ok {
		if record := s.unionRecord(attribute); record != nil {
			if pkg == "" {
				return record.name
			}
			return pkg + "." + record.name
		}
	}
	switch actual := attribute.Type.(type) {
	case expr.Primitive:
		if name, _ := codegen.GetMetaType(attribute); name != "" {
			return name
		}
		return codegen.GoNativeTypeName(actual)
	case *expr.Array, *expr.Map, *expr.Object:
		context := &codegen.AttributeContext{
			Pointer:             pointer,
			UseDefault:          useDefault,
			Scope:               s,
			UnionPointer:        true,
			ArrayElementPointer: s.policy.arrayElementPointer,
		}
		return goTypeDefForContext(attribute, context)
	case expr.UserType:
		panic(fmt.Sprintf("HTTP type %q has no package declaration", wireTypeDeclaredName(actual)))
	default:
		return s.base.Name(attribute, pkg, pointer, useDefault)
	}
}

// Ref returns the pointer or value spelling for this HTTP attribute copy.
func (s *wireAttributeScope) Ref(attribute *expr.AttributeExpr, pkg string) string {
	layout, err := s.goTypeLayout(attribute, wireGoLayoutPolicy(s.policy), true)
	if err != nil {
		panic(fmt.Sprintf("resolve HTTP type %q: %v", attribute.Type.Name(), err))
	}
	return layout.Ref()
}

// GoTypeLayout returns the exact fields and references selected by this HTTP
// package for attribute.
func (s *wireAttributeScope) GoTypeLayout(attribute *expr.AttributeExpr, policy codegen.GoLayoutPolicy) (codegen.LinkedGoType, error) {
	return s.goTypeLayout(attribute, policy, false)
}

// goTypeLayout may stop at a named root when a caller only needs its reference.
func (s *wireAttributeScope) goTypeLayout(attribute *expr.AttributeExpr, policy codegen.GoLayoutPolicy, referenceOnly bool) (codegen.LinkedGoType, error) {
	owner := s.catalog.pkg.ImportPath()
	layout, err := codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
		Owner:            owner,
		Policy:           policy,
		RetainNamedValue: !referenceOnly,
		Bind: func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
			switch request.Kind {
			case codegen.GoNamed:
				record := s.record(request.Attribute)
				if record == nil || record.declaration == nil {
					return codegen.GoTypeBinding{}, fmt.Errorf("HTTP type %q has no planned declaration", request.Attribute.Type.Name())
				}
				return codegen.GoTypeBinding{Owner: owner, Declaration: record.declaration}, nil
			case codegen.GoUnion:
				record := s.unionRecord(request.Attribute)
				if record == nil || record.declaration == nil {
					return codegen.GoTypeBinding{}, fmt.Errorf("HTTP OneOf %q has no planned declaration", request.Attribute.Type.Name())
				}
				return codegen.GoTypeBinding{Owner: owner, Declaration: record.declaration}, nil
			default:
				return codegen.GoTypeBinding{}, fmt.Errorf("resolve unsupported HTTP Go type kind %s", request.Kind)
			}
		},
	})
	if err != nil {
		return codegen.LinkedGoType{}, err
	}
	return layout.Link(owner, s.catalog.pkg.ImportName), nil
}

// Field returns the generated Go field for an HTTP attribute.
func (*wireAttributeScope) Field(attribute *expr.AttributeExpr, name string, firstUpper bool) string {
	return codegen.GoifyAtt(attribute, name, firstUpper)
}

// Package returns the Go package name written before the type for attribute.
func (s *wireAttributeScope) Package(attribute *expr.AttributeExpr) string {
	if location := codegen.UserTypeLocation(attribute.Type); location != nil {
		return location.PackageName()
	}
	return s.pkg
}

// Enter returns type names with the Go package name needed by nested fields.
func (s *wireAttributeScope) Enter(attribute *expr.AttributeExpr) codegen.Attributor {
	pkg := s.pkg
	if location := codegen.UserTypeLocation(attribute.Type); location != nil {
		pkg = location.PackageName()
	}
	policy := s.policy
	viewRoot := s.viewRoot
	if policy.view != "" && (s.viewRoot == nil || !wireTypeIdentitiesEqual(
		s.viewRoot.identity,
		newWireTypeIdentity(attribute, s.viewRoot.identity.role, policy, s.viewRoot.identity.preferred),
	)) {
		policy.view = ""
		viewRoot = nil
	}
	return &wireAttributeScope{
		catalog:         s.catalog,
		base:            s.base.Enter(attribute),
		pkg:             pkg,
		policy:          policy,
		use:             s.use,
		viewRoot:        viewRoot,
		exactOccurrence: s.exactOccurrence,
	}
}

// IsSumType reports that HTTP unions use generated values that hold one branch.
func (*wireAttributeScope) IsSumType() bool {
	return true
}

// ValidatorCall returns the exact private call used for a named value inside
// another HTTP body value.
func (s *wireAttributeScope) ValidatorCall(attribute *expr.AttributeExpr, _, target, path string) string {
	if record := s.record(attribute); record != nil {
		if record.nestedValidator == nil {
			panic(fmt.Sprintf("HTTP type %q has no nested validator", record.name))
		}
		return fmt.Sprintf("%s(%s, %s)", record.nestedValidator.Name(), target, path)
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		panic(fmt.Sprintf("HTTP validator for %q has no package declaration", wireTypeDeclaredName(userType)))
	}
	return s.base.ValidatorCall(attribute, "", target, path)
}

// record returns the chosen type for attribute. A copied nested value may reuse
// a type when its pointer and default-value rules are the same.
func (s *wireAttributeScope) record(attribute *expr.AttributeExpr) *wireTypeRecord {
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return nil
	}
	if s.exactOccurrence {
		return s.catalog.bindings[attribute]
	}
	preferred := wireTypePreferredName(userType, s.policy)
	if s.viewRoot != nil {
		identity := newWireTypeIdentity(attribute, s.viewRoot.identity.role, s.policy, s.viewRoot.identity.preferred)
		if wireTypeIdentitiesEqual(s.viewRoot.identity, identity) {
			return s.viewRoot
		}
	}
	identity := newWireTypeIdentity(attribute, wireAttribute, s.policy, preferred)
	if record := s.catalog.bindings[attribute]; record != nil && wireTypeIdentitiesEqual(record.identity, identity) {
		return record
	}
	return s.catalog.find(identity)
}

// unionRecord returns the declaration assigned to this exact copied union.
func (s *wireAttributeScope) unionRecord(attribute *expr.AttributeExpr) *wireUnionRecord {
	union := attribute.Type.(*expr.Union)
	if record := s.catalog.unionBindings[union]; record != nil {
		return record
	}
	identity := s.catalog.unionIdentity(attribute, s.use, wireAttribute, s.policy)
	if record := s.catalog.findUnion(identity); record != nil {
		s.catalog.applyUnionRecord(attribute, record)
		return record
	}
	if s.catalog.linked {
		panic(fmt.Sprintf("HTTP union %q was not bound to its planned declaration", union.Name()))
	}
	return nil
}

// Scope returns the list used to keep local variable names unique.
func (s *wireAttributeScope) Scope() *codegen.NameScope {
	return s.base.Scope()
}

// findUnion returns the package record for the same generated union.
func (c *wireTypeCatalog) findUnion(identity wireUnionIdentity) *wireUnionRecord {
	for _, record := range c.unions {
		if wireUnionIdentitiesEqual(record.identity, identity) {
			return record
		}
	}
	return nil
}

// findUnionOwner returns the declaration for one authored OneOf in one HTTP
// body role and response view.
func (c *wireTypeCatalog) findUnionOwner(owner wireUnionOwner) *wireUnionRecord {
	for _, record := range c.unions {
		if record.identity.owner == owner {
			return record
		}
	}
	return nil
}

// wireUnionIdentitiesEqual reports whether two copies come from the same OneOf,
// serve the same HTTP use, and emit the same branch declarations.
func wireUnionIdentitiesEqual(left, right wireUnionIdentity) bool {
	return left.declaration == right.declaration && left.owner.use == right.owner.use && slices.Equal(left.declarations, right.declarations)
}

// nameFor returns the exact public name for one union in this HTTP body role.
func (u wireUnionUse) nameFor(union *expr.Union) string {
	return codegen.Goify(union.Name(), true) + u.suffix()
}

// suffix returns the fixed public suffix for one HTTP body role and response
// view.
func (u wireUnionUse) suffix() string {
	switch u.role {
	case wireRequestBody:
		return "RequestBody"
	case wireStreamPayload:
		return "StreamingBody"
	case wireResponseBody:
		return codegen.Goify(u.view, true) + "ResponseBody"
	default:
		return ""
	}
}

// description names one HTTP use in an error shown to a design author.
func (u wireUnionUse) description() string {
	if suffix := u.suffix(); suffix != "" {
		return suffix
	}
	return "direct attribute"
}

// declareUnionName reserves one exact public name and explains how the design
// author can resolve a collision.
func (c *wireTypeCatalog) declareUnionName(union *expr.Union, use wireUnionUse, declaration *codegen.NameDeclaration) error {
	if err := c.pkg.DeclareName(declaration); err != nil {
		return fmt.Errorf(
			"declare HTTP OneOf %q for %s: %w; set TypeName on the OneOf to a unique name",
			union.Name(),
			use.description(),
			err,
		)
	}
	return nil
}

// validateHTTPUnionBranchGoNames rejects branch spellings that would create
// the same exported Go identifier and therefore the same public union names.
func validateHTTPUnionBranchGoNames(union *expr.Union) error {
	branches := make(map[string]string, len(union.Values))
	for _, branch := range union.Values {
		goName := codegen.Goify(branch.Name, true)
		if existing, ok := branches[goName]; ok {
			return fmt.Errorf(
				"OneOf %q branches %q and %q both generate Go name %q; rename one of the branches",
				union.Name(),
				existing,
				branch.Name,
				goName,
			)
		}
		branches[goName] = branch.Name
	}
	return nil
}

// newWireTypeIdentity records the designed value and the rules that determine
// its generated Go type.
func newWireTypeIdentity(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) wireTypeIdentity {
	identity := wireTypeIdentity{role: role, preferred: preferred, attribute: expr.DupAtt(attribute), policy: policy}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if resultType, ok := userType.(*expr.ResultTypeExpr); ok {
			identity.resultID = resultType.Identifier
			identity.role = 0
			identity.policy.view = ""
		} else if policy.view == "" {
			identity.sourceID = userType.Origin().ID()
			identity.role = 0
		}
	}
	return identity
}

// wireTypePreferredName returns the Go name requested for one designed type.
// A result type created for one view already has the view in its name. Other
// values use their authored type name unless selecting a view changes the
// fields in the generated transport type.
func wireTypePreferredName(userType expr.UserType, policy wireTypePolicy) string {
	named := userType.Origin()
	if _, projected := userType.(*expr.ResultTypeExpr); projected || policy.view != "" {
		named = userType
	}
	return codegen.Goify(wireTypeDeclaredName(named), true)
}

// releasedWireTypeSuffix returns the suffix that HTTP copies added to named
// values nested inside one body. The body declaration itself already keeps its
// endpoint name.
func releasedWireTypeSuffix(attribute *expr.AttributeExpr, role wireTypeRole) string {
	switch role {
	case wireRequestBody:
		return "RequestBody"
	case wireStreamPayload:
		if expr.IsObject(attribute.Type) {
			return "StreamingBody"
		}
		return ""
	case wireResponseBody:
		if expr.IsObject(attribute.Type) {
			return "ResponseBody"
		}
		return "Response"
	default:
		return ""
	}
}

// wireTypeIdentitiesEqual reports whether two records produce the same Go type.
func wireTypeIdentitiesEqual(left, right wireTypeIdentity) bool {
	if left.sourceID != right.sourceID || left.resultID != right.resultID || left.role != right.role || left.preferred != right.preferred || !wireTypePoliciesEqual(left.policy, right.policy) {
		return false
	}
	if left.sourceID != "" {
		leftType := left.attribute.Type.(expr.UserType)
		rightType := right.attribute.Type.(expr.UserType)
		return wireAttributesEqual(leftType.Attribute(), rightType.Attribute(), make(map[wireAttributePair]struct{}))
	}
	if leftType, ok := left.attribute.Type.(expr.UserType); ok {
		rightType, ok := right.attribute.Type.(expr.UserType)
		return ok && wireAttributesEqual(leftType.Attribute(), rightType.Attribute(), make(map[wireAttributePair]struct{}))
	}
	return wireAttributesEqual(left.attribute, right.attribute, make(map[wireAttributePair]struct{}))
}

// order returns the designed values used to choose a stable suffix when several
// HTTP declarations ask for the same Go name.
func (i wireTypeIdentity) order(kind wireNameKind) wireNameOrder {
	return wireNameOrder{
		kind:                kind,
		api:                 i.api,
		source:              i.sourceID + i.resultID,
		role:                uint8(i.role),
		preferred:           i.preferred,
		shape:               expr.Hash(i.attribute.Type, false, false, false),
		view:                i.policy.view,
		request:             i.policy.request,
		pointer:             i.policy.pointer,
		arrayElementPointer: i.policy.arrayElementPointer,
		defaults:            i.policy.useDefault,
	}
}

// ComparePackageName orders HTTP declarations from designed values so memory
// addresses and design reading order cannot change generated names.
func (o wireNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(wireNameOrder)
	for _, compared := range []int{
		cmp.Compare(o.kind, right.kind),
		cmp.Compare(o.api, right.api),
		cmp.Compare(o.source, right.source),
		cmp.Compare(o.target, right.target),
		cmp.Compare(o.role, right.role),
		cmp.Compare(o.preferred, right.preferred),
		cmp.Compare(o.shape, right.shape),
		cmp.Compare(o.view, right.view),
		cmp.Compare(boolOrder(o.request), boolOrder(right.request)),
		cmp.Compare(boolOrder(o.pointer), boolOrder(right.pointer)),
		cmp.Compare(boolOrder(o.arrayElementPointer), boolOrder(right.arrayElementPointer)),
		cmp.Compare(boolOrder(o.defaults), boolOrder(right.defaults)),
		cmp.Compare(boolOrder(o.required), boolOrder(right.required)),
	} {
		if compared != 0 {
			return compared
		}
	}
	return 0
}

// boolOrder converts false to zero and true to one for name ordering.
func boolOrder(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}

// wireTypePoliciesEqual compares only rules that change a Go type definition.
// A validation function does not create a second type when every field is the
// same.
func wireTypePoliciesEqual(left, right wireTypePolicy) bool {
	left.validate = false
	right.validate = false
	return left == right
}

// wireAttributesEqual compares the designed facts that change a generated type
// or its validation function. It handles types that refer back to themselves.
func wireAttributesEqual(left, right *expr.AttributeExpr, seen map[wireAttributePair]struct{}) bool {
	if left == right {
		return true
	}
	pair := wireAttributePair{left: left, right: right}
	if _, ok := seen[pair]; ok {
		return true
	}
	seen[pair] = struct{}{}
	if !reflect.DeepEqual(left.DefaultValue, right.DefaultValue) || !reflect.DeepEqual(left.Validation, right.Validation) || !wireMetadataEqual(left.Meta, right.Meta) {
		return false
	}
	switch ltype := left.Type.(type) {
	case expr.UserType:
		rtype, ok := right.Type.(expr.UserType)
		if !ok {
			return false
		}
		if lresult, ok := ltype.(*expr.ResultTypeExpr); ok {
			rresult, ok := rtype.(*expr.ResultTypeExpr)
			return ok && lresult.Identifier == rresult.Identifier && lresult.Name() == rresult.Name() && wireAttributesEqual(lresult.Attribute(), rresult.Attribute(), seen)
		}
		return ltype.Origin() == rtype.Origin() && wireAttributesEqual(ltype.Attribute(), rtype.Attribute(), seen)
	case *expr.Object:
		rtype, ok := right.Type.(*expr.Object)
		if !ok || len(*ltype) != len(*rtype) {
			return false
		}
		for index, field := range *ltype {
			other := (*rtype)[index]
			if field.Name != other.Name || !wireAttributesEqual(field.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case *expr.Array:
		rtype, ok := right.Type.(*expr.Array)
		return ok && ltype.NonNullableElems == rtype.NonNullableElems && wireAttributesEqual(ltype.ElemType, rtype.ElemType, seen)
	case *expr.Map:
		rtype, ok := right.Type.(*expr.Map)
		return ok && wireAttributesEqual(ltype.KeyType, rtype.KeyType, seen) && wireAttributesEqual(ltype.ElemType, rtype.ElemType, seen)
	case *expr.Union:
		rtype, ok := right.Type.(*expr.Union)
		if !ok || ltype.Name() != rtype.Name() || ltype.GetTypeKey() != rtype.GetTypeKey() || ltype.GetValueKey() != rtype.GetValueKey() || len(ltype.Values) != len(rtype.Values) {
			return false
		}
		for index, branch := range ltype.Values {
			other := rtype.Values[index]
			if branch.Name != other.Name || !wireAttributesEqual(branch.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	default:
		return left.Type.Kind() == right.Type.Kind() && left.Type.Name() == right.Type.Name()
	}
}

// wireMetadataEqual compares design metadata while ignoring the Go name added later.
func wireMetadataEqual(left, right expr.MetaExpr) bool {
	keys := make([]string, 0, len(left))
	for key := range left {
		if key != "struct:type:name" {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	otherKeys := make([]string, 0, len(right))
	for key := range right {
		if key != "struct:type:name" {
			otherKeys = append(otherKeys, key)
		}
	}
	slices.Sort(otherKeys)
	if !slices.Equal(keys, otherKeys) {
		return false
	}
	for _, key := range keys {
		if !slices.Equal(left[key], right[key]) {
			return false
		}
	}
	return true
}

// sortedWireAttributes keeps generated Go names stable when object fields are
// listed in a different order.
func sortedWireAttributes(attributes []*expr.NamedAttributeExpr) []*expr.NamedAttributeExpr {
	sorted := slices.Clone(attributes)
	slices.SortFunc(sorted, func(left, right *expr.NamedAttributeExpr) int {
		return strings.Compare(left.Name, right.Name)
	})
	return sorted
}

// wireTypeDeclaredName returns the type name written in the design instead of a
// generated Go name saved in metadata.
func wireTypeDeclaredName(userType expr.UserType) string {
	switch actual := userType.(type) {
	case *expr.UserTypeExpr:
		return actual.TypeName
	case *expr.ResultTypeExpr:
		return actual.TypeName
	default:
		panic(fmt.Sprintf("unsupported HTTP wire user type %T", userType))
	}
}
