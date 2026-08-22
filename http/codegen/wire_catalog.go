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
		pkg              *codegen.GeneratedPackage
		scope            *codegen.NameScope
		records          []*wireTypeRecord
		transforms       []*wireTransformRecord
		unionOccurrences []wireUnionOccurrence
		unions           []*wireUnionRecord
		declared         bool
		linked           bool
		bindings         map[*expr.AttributeExpr]*wireTypeRecord
		unionBindings    map[*expr.Union]*wireUnionRecord
	}

	// wireTransformRecord stores one value conversion and any extra functions it
	// needs.
	wireTransformRecord struct {
		source *expr.AttributeExpr
		target *expr.AttributeExpr
		prefix string
		owner  string
		plan   *codegen.TransformPlan
		used   bool
	}

	// wireUnionRecord stores one generated union and the Go names used for its
	// type, branches, constants, and functions.
	wireUnionRecord struct {
		identity     wireUnionIdentity
		union        *expr.Union
		declaration  *codegen.NameDeclaration
		kind         *codegen.NameDeclaration
		kindDecls    []*codegen.NameDeclaration
		ctorDecls    []*codegen.NameDeclaration
		name         string
		kindName     string
		kindConsts   []string
		constructors []string
		data         *service.UnionTypeData
	}

	// wireUnionOccurrence stores one copied union until Goa assigns names to its branches.
	wireUnionOccurrence struct {
		union  *expr.Union
		role   wireTypeRole
		policy wireTypePolicy
	}

	// wireUnionIdentity pairs a union definition with the Go type used by each branch.
	wireUnionIdentity struct {
		definition   codegen.UnionTypeID
		declarations []*wireTypeRecord
	}

	// wireTypeRecord stores one generated type and its optional functions.
	wireTypeRecord struct {
		identity         wireTypeIdentity
		declaration      *codegen.NameDeclaration
		validator        *codegen.NameDeclaration
		constructor      *codegen.NameDeclaration
		needsValidator   bool
		needsConstructor bool
		name             string
		ref              string
		data             *TypeData
	}

	// wireTypeIdentity contains a designed type and the rules that change its Go definition.
	wireTypeIdentity struct {
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
		request    bool
		pointer    bool
		useDefault bool
		validate   bool
		view       string
	}

	// wireTypeRole says whether an unnamed designed type is used for a request,
	// response, field, or stream value.
	wireTypeRole uint8

	// wireAttributePair remembers two attributes already compared so values that
	// refer back to themselves do not cause an endless loop.
	wireAttributePair struct {
		left  *expr.AttributeExpr
		right *expr.AttributeExpr
	}

	// wireNameOrder contains designed values used to choose stable suffixes when
	// several declarations ask for the same Go name.
	wireNameOrder struct {
		family    string
		source    string
		role      uint8
		preferred string
		shape     string
		view      string
		request   bool
		pointer   bool
		defaults  bool
	}

	// wireAttributeScope chooses the Go type name for each copied HTTP field.
	wireAttributeScope struct {
		catalog *wireTypeCatalog
		base    codegen.Attributor
		pkg     string
		policy  wireTypePolicy
	}
)

const (
	wireRequestBody wireTypeRole = iota + 1
	wireResponseBody
	wireAttribute
	wireStreamPayload
)

// newWireTypeCatalog creates the type list for one generated Go package. Tests
// may omit the package when they only compare copied attributes.
func newWireTypeCatalog(pkg ...*codegen.GeneratedPackage) *wireTypeCatalog {
	catalog := &wireTypeCatalog{
		bindings:      make(map[*expr.AttributeExpr]*wireTypeRecord),
		unionBindings: make(map[*expr.Union]*wireUnionRecord),
	}
	if len(pkg) > 0 {
		catalog.pkg = pkg[0]
	}
	return catalog
}

// collect records attribute and every named type it contains.
func (c *wireTypeCatalog) collect(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) *wireTypeRecord {
	if c.declared {
		panic("cannot collect an HTTP type after its package declarations are submitted")
	}
	return c.collectRecursive(attribute, role, policy, preferred, make(map[expr.UserType]struct{}))
}

// collectChildren records named types inside attribute without recording its
// top-level named type a second time.
func (c *wireTypeCatalog) collectChildren(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy) {
	if userType, ok := attribute.Type.(expr.UserType); ok {
		c.collectRecursive(userType.Attribute(), role, policy, "", make(map[expr.UserType]struct{}))
		return
	}
	c.collectRecursive(attribute, role, policy, "", make(map[expr.UserType]struct{}))
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
	for _, record := range c.records {
		record.declaration = codegen.NewPreferredName(
			codegen.NameType,
			record.identity.preferred,
			codegen.ExportedName,
			record.identity.order("type"),
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
				record.identity.order("validator"),
			)
			if err != nil {
				return err
			}
			record.validator = declaration
		}
		if record.needsConstructor {
			declaration, err := c.pkg.DeclareDependentName(
				codegen.NameFunction,
				record.declaration,
				"New",
				"",
				record.identity.order("constructor"),
			)
			if err != nil {
				return err
			}
			record.constructor = declaration
		}
	}
	for _, occurrence := range c.unionOccurrences {
		identity := c.unionIdentity(occurrence.union, occurrence.role, occurrence.policy)
		if c.findUnion(identity) == nil {
			c.unions = append(c.unions, &wireUnionRecord{identity: identity, union: occurrence.union})
		}
	}
	for _, union := range c.unions {
		union.declaration = codegen.NewPreferredName(
			codegen.NameType,
			union.union.Name(),
			codegen.ExportedName,
			union.identity.order("union", union.union.Name(), ""),
		)
		if err := c.pkg.DeclareName(union.declaration); err != nil {
			return err
		}
		kind, err := c.pkg.DeclareDependentName(
			codegen.NameType,
			union.declaration,
			"",
			"Kind",
			union.identity.order("union kind", union.union.Name(), ""),
		)
		if err != nil {
			return err
		}
		union.kind = kind
		union.kindDecls = make([]*codegen.NameDeclaration, len(union.union.Values))
		union.ctorDecls = make([]*codegen.NameDeclaration, len(union.union.Values))
		for index, branch := range union.union.Values {
			kindDeclaration, err := c.pkg.DeclareDependentName(
				codegen.NameConstant,
				union.kind,
				"",
				codegen.Goify(branch.Name, true),
				union.identity.order("union constant", union.union.Name(), branch.Name),
			)
			if err != nil {
				return err
			}
			constructor, err := c.pkg.DeclareDependentName(
				codegen.NameFunction,
				union.declaration,
				"New",
				codegen.Goify(branch.Name, true),
				union.identity.order("union constructor", union.union.Name(), branch.Name),
			)
			if err != nil {
				return err
			}
			union.kindDecls[index] = kindDeclaration
			union.ctorDecls[index] = constructor
		}
	}
	for _, transform := range c.transforms {
		for _, helper := range transform.plan.Helpers() {
			preferred := transform.prefix + codegen.Goify(wireTransformTypeName(helper.Source), true) + "To" + codegen.Goify(wireTransformTypeName(helper.Target), true)
			declaration := codegen.NewPreferredName(
				codegen.NameFunction,
				preferred,
				codegen.UnexportedName,
				wireNameOrder{
					family:    "transform helper",
					source:    expr.Hash(transform.source.Type, false, false, false),
					preferred: preferred,
					shape:     expr.Hash(transform.target.Type, false, false, false),
					role:      uint8(helper.Occurrence),
					view:      transform.owner,
				},
			)
			if err := c.pkg.DeclareName(declaration); err != nil {
				return err
			}
			if err := transform.plan.BindHelperDeclaration(helper.ID, declaration); err != nil {
				return err
			}
		}
	}
	c.declared = true
	return nil
}

// collectTransform records one request or response conversion before Goa
// chooses the names of any extra conversion functions. Calls use these records
// in the same order.
func (c *wireTypeCatalog) collectTransform(source, target *expr.AttributeExpr, prefix, owner string) {
	if c.declared {
		panic("cannot collect an HTTP conversion after package declarations are submitted")
	}
	source = expr.DupAtt(source)
	target = expr.DupAtt(target)
	plan, err := codegen.NewTransformPlan(source, target)
	if err != nil {
		panic(err)
	}
	c.transforms = append(c.transforms, &wireTransformRecord{source: source, target: target, prefix: prefix, owner: owner, plan: plan})
}

// renderTransform writes the next matching conversion with the function names
// chosen by Declare. It returns an error when collectTransform did not record
// the conversion.
func (c *wireTypeCatalog) renderTransform(source, target *expr.AttributeExpr, sourceVar, targetVar, prefix string, sourceContext, targetContext *codegen.AttributeContext) (string, []*codegen.TransformFunctionData, error) {
	for _, transform := range c.transforms {
		if transform.used || transform.prefix != prefix ||
			!wireAttributesEqual(transform.source, source, make(map[wireAttributePair]struct{})) ||
			!wireAttributesEqual(transform.target, target, make(map[wireAttributePair]struct{})) {
			continue
		}
		if err := transform.plan.BindContexts(sourceContext, targetContext); err != nil {
			return "", nil, err
		}
		c.bindTransformOccurrence(transform.source, source, sourceContext)
		c.bindTransformOccurrence(transform.target, target, targetContext)
		for _, helper := range transform.plan.Helpers() {
			c.bindTransformHelper(helper.Source, sourceContext)
			c.bindTransformHelper(helper.Target, targetContext)
		}
		transform.used = true
		return transform.plan.Render(sourceVar, targetVar, true)
	}
	return "", nil, fmt.Errorf("HTTP %s conversion was not submitted before package names were assigned", prefix)
}

// bindTransformHelper gives a nested copied field the same Go type name used by
// its generated conversion function. Service values already receive names from
// the service generator.
func (c *wireTypeCatalog) bindTransformHelper(attribute *expr.AttributeExpr, context *codegen.AttributeContext) {
	scope, ok := context.Scope.(*wireAttributeScope)
	if !ok || scope.catalog != c {
		return
	}
	policy := scope.policy
	policy.view = ""
	c.applyNamesRecursive(attribute, wireAttribute, policy, make(map[expr.UserType]struct{}))
}

// bindTransformOccurrence records the Go type used by one copied conversion
// value and each named field inside it.
func (c *wireTypeCatalog) bindTransformOccurrence(planned, rendered *expr.AttributeExpr, context *codegen.AttributeContext) {
	scope, ok := context.Scope.(*wireAttributeScope)
	if !ok || scope.catalog != c {
		return
	}
	record := c.bindings[rendered]
	if record == nil {
		c.applyNamesRecursive(planned, wireAttribute, scope.policy, make(map[expr.UserType]struct{}))
		return
	}
	c.bindOccurrence(planned, record)
	c.applyNamesRecursive(planned, record.identity.role, record.identity.policy, make(map[expr.UserType]struct{}))
}

// wireTransformTypeName returns the designed type name used in a generated
// conversion function name.
func wireTransformTypeName(attribute *expr.AttributeExpr) string {
	if userType, ok := attribute.Type.(expr.UserType); ok {
		name := wireTypeDeclaredName(userType)
		if location := codegen.UserTypeLocation(userType); location != nil {
			return location.PackageName() + codegen.Goify(name, true)
		}
		return name
	}
	if union, ok := attribute.Type.(*expr.Union); ok {
		return union.Name()
	}
	return attribute.Type.Name()
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
	for _, record := range c.records {
		record.name = record.declaration.Name()
		record.ref = wireTypeRef(record.name, record.identity.attribute.Type)
	}
	for _, union := range c.unions {
		union.name = union.declaration.Name()
		union.kindName = union.kind.Name()
		union.kindConsts = make([]string, len(union.kindDecls))
		union.constructors = make([]string, len(union.ctorDecls))
		for index := range union.kindDecls {
			union.kindConsts[index] = union.kindDecls[index].Name()
			union.constructors[index] = union.ctorDecls[index].Name()
		}
		c.applyUnionRecord(union.union, union)
	}
	for _, union := range c.unions {
		union.data = buildHTTPUnionTypeData(union.union, c.resolver(c.scope, wireTypePolicy{}), union)
	}
	c.linked = true
}

// wireTypeRef adds a pointer when the generated Go type requires one.
func wireTypeRef(name string, dataType expr.DataType) string {
	if _, inline := dataType.(*expr.Object); inline {
		return name
	}
	if expr.IsObject(dataType) || expr.IsUnion(dataType) {
		return "*" + name
	}
	return name
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
	preferred := wireTypeDeclaredName(userType.Origin())
	if policy.view != "" {
		preferred = wireTypeDeclaredName(userType)
	}
	return c.lookup(attribute, role, policy, codegen.Goify(preferred, true))
}

// applyNames associates every copied nested attribute with the Go type name
// used by its definition and conversions.
func (c *wireTypeCatalog) applyNames(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy) {
	c.applyNamesRecursive(attribute, role, policy, make(map[expr.UserType]struct{}))
}

// applyNamesRecursive follows each named field once, including fields that
// refer back to an outer type.
func (c *wireTypeCatalog) applyNamesRecursive(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, seen map[expr.UserType]struct{}) {
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
		c.applyNamesRecursive(userType.Attribute(), wireAttribute, nestedPolicy, seen)
		delete(seen, origin)
		return
	}
	nestedPolicy := policy
	nestedPolicy.view = ""
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, named := range *actual {
			c.applyNamesRecursive(named.Attribute, wireAttribute, nestedPolicy, seen)
		}
	case *expr.Array:
		c.applyNamesRecursive(actual.ElemType, wireAttribute, nestedPolicy, seen)
	case *expr.Map:
		c.applyNamesRecursive(actual.KeyType, wireAttribute, nestedPolicy, seen)
		c.applyNamesRecursive(actual.ElemType, wireAttribute, nestedPolicy, seen)
	case *expr.Union:
		identity := c.unionIdentity(actual, role, policy)
		record := c.findUnion(identity)
		if record == nil {
			panic(fmt.Sprintf("HTTP union %q was not submitted before package names were assigned", actual.Name()))
		}
		c.applyUnionRecord(actual, record)
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
	if data.ValidateDef != "" {
		data.ValidatorName = record.validator.Name()
		data.ValidateRef = strings.Replace(data.ValidateRef, "Validate"+record.name, data.ValidatorName, 1)
	}
	if data.Init != nil {
		data.Init.Declaration = record.constructor
		data.Init.Name = record.constructor.Name()
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
		} else if record.data.ValidateDef != data.ValidateDef || record.data.ValidateRef != data.ValidateRef {
			panic(fmt.Sprintf("HTTP type %q produced conflicting validators", record.name))
		}
	}
	return data
}

// collectRecursive records named types and stops when a type refers back to one
// it is already reading.
func (c *wireTypeCatalog) collectRecursive(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string, seen map[expr.UserType]struct{}) *wireTypeRecord {
	if attribute.Type == expr.Empty {
		return nil
	}
	var record *wireTypeRecord
	if userType, ok := attribute.Type.(expr.UserType); ok {
		name := wireTypeDeclaredName(userType.Origin())
		if policy.view != "" {
			name = wireTypeDeclaredName(userType)
		}
		preferred = codegen.Goify(name, true)
		record = c.findOrAppend(newWireTypeIdentity(attribute, role, policy, preferred))
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return record
		}
		seen[origin] = struct{}{}
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectRecursive(userType.Attribute(), wireAttribute, nestedPolicy, "", seen)
		delete(seen, origin)
		return record
	}
	if preferred != "" {
		record = c.findOrAppend(newWireTypeIdentity(attribute, role, policy, preferred))
	}
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		nestedPolicy := policy
		nestedPolicy.view = ""
		for _, named := range sortedWireAttributes(*actual) {
			c.collectRecursive(named.Attribute, wireAttribute, nestedPolicy, "", seen)
		}
	case *expr.Array:
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectRecursive(actual.ElemType, wireAttribute, nestedPolicy, "", seen)
	case *expr.Map:
		nestedPolicy := policy
		nestedPolicy.view = ""
		c.collectRecursive(actual.KeyType, wireAttribute, nestedPolicy, "", seen)
		c.collectRecursive(actual.ElemType, wireAttribute, nestedPolicy, "", seen)
	case *expr.Union:
		nestedPolicy := policy
		nestedPolicy.view = ""
		union := expr.Dup(actual).(*expr.Union)
		c.unionOccurrences = append(c.unionOccurrences, wireUnionOccurrence{union: union, role: role, policy: policy})
		for _, named := range actual.Values {
			c.collectRecursive(named.Attribute, wireAttribute, nestedPolicy, "", seen)
		}
	}
	return record
}

// findOrAppend reuses a record with the same generated type definition or adds
// a new record.
func (c *wireTypeCatalog) findOrAppend(identity wireTypeIdentity) *wireTypeRecord {
	if record := c.find(identity); record != nil {
		record.needsValidator = record.needsValidator || identity.policy.validate
		return record
	}
	record := &wireTypeRecord{identity: identity, needsValidator: identity.policy.validate}
	c.records = append(c.records, record)
	return record
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

// unionIdentity returns the Go type used by every named branch of
// union without changing union.
func (c *wireTypeCatalog) unionIdentity(union *expr.Union, role wireTypeRole, policy wireTypePolicy) wireUnionIdentity {
	identity := wireUnionIdentity{definition: codegen.NewUnionTypeID(union)}
	attribute := &expr.AttributeExpr{Type: union}
	c.collectUnionDeclarations(attribute, role, policy, &identity.declarations, make(map[expr.UserType]struct{}))
	return identity
}

// collectUnionDeclarations records generated branch types in branch order.
func (c *wireTypeCatalog) collectUnionDeclarations(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, declarations *[]*wireTypeRecord, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		preferred := wireTypeDeclaredName(userType.Origin())
		if policy.view != "" {
			preferred = wireTypeDeclaredName(userType)
		}
		record := c.find(newWireTypeIdentity(attribute, role, policy, codegen.Goify(preferred, true)))
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
func (c *wireTypeCatalog) applyUnionRecord(union *expr.Union, record *wireUnionRecord) {
	c.unionBindings[union] = record
	index := 0
	seen := make(map[expr.UserType]struct{})
	for _, branch := range union.Values {
		c.applyResolvedDeclarations(branch.Attribute, record.identity.declarations, &index, seen)
	}
	if index != len(record.identity.declarations) {
		panic(fmt.Sprintf("HTTP union %q did not use every submitted branch name", record.name))
	}
}

// applyResolvedDeclarations gives each named branch its previously chosen Go
// type in the same order those types were recorded.
func (c *wireTypeCatalog) applyResolvedDeclarations(attribute *expr.AttributeExpr, declarations []*wireTypeRecord, index *int, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if *index >= len(declarations) {
			panic(fmt.Sprintf("HTTP union branch %q has no submitted Go type name", wireTypeDeclaredName(userType)))
		}
		record := declarations[*index]
		*index = *index + 1
		c.bindOccurrence(attribute, record)
		origin := userType.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		c.applyResolvedDeclarations(userType.Attribute(), declarations, index, seen)
		delete(seen, origin)
		return
	}
	switch actual := attribute.Type.(type) {
	case *expr.Object:
		for _, named := range *actual {
			c.applyResolvedDeclarations(named.Attribute, declarations, index, seen)
		}
	case *expr.Array:
		c.applyResolvedDeclarations(actual.ElemType, declarations, index, seen)
	case *expr.Map:
		c.applyResolvedDeclarations(actual.KeyType, declarations, index, seen)
		c.applyResolvedDeclarations(actual.ElemType, declarations, index, seen)
	case *expr.Union:
		definition := codegen.NewUnionTypeID(actual)
		start := *index
		for _, branch := range actual.Values {
			c.applyResolvedDeclarations(branch.Attribute, declarations, index, seen)
		}
		identity := wireUnionIdentity{definition: definition, declarations: declarations[start:*index]}
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

// bindOccurrence records the Go type used by one copied named value and its fields.
func (c *wireTypeCatalog) bindOccurrence(attribute *expr.AttributeExpr, record *wireTypeRecord) {
	c.bindings[attribute] = record
	if userType, ok := attribute.Type.(expr.UserType); ok {
		c.bindings[userType.Attribute()] = record
	}
}

// Name returns the type name selected for this HTTP attribute copy.
func (s *wireAttributeScope) Name(attribute *expr.AttributeExpr, pkg string, pointer, useDefault bool) string {
	if record := s.record(attribute); record != nil {
		if pkg == "" {
			return record.name
		}
		return pkg + "." + record.name
	}
	if union, ok := attribute.Type.(*expr.Union); ok {
		if record := s.unionRecord(union); record != nil {
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
			Pointer:      pointer,
			UseDefault:   useDefault,
			Scope:        s,
			UnionPointer: true,
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
	return wireTypeRef(s.Name(attribute, pkg, s.policy.pointer, s.policy.useDefault), attribute.Type)
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
	return &wireAttributeScope{catalog: s.catalog, base: s.base.Enter(attribute), pkg: pkg, policy: s.policy}
}

// IsSumType reports that HTTP unions use generated sum-type structs.
func (*wireAttributeScope) IsSumType() bool {
	return true
}

// ValidatorName returns the validation function chosen for this copied type.
func (s *wireAttributeScope) ValidatorName(attribute *expr.AttributeExpr, view string) string {
	if record := s.record(attribute); record != nil {
		if record.validator == nil {
			panic(fmt.Sprintf("HTTP type %q has no validator declaration", record.name))
		}
		return record.validator.Name()
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		panic(fmt.Sprintf("HTTP validator for %q has no package declaration", wireTypeDeclaredName(userType)))
	}
	return s.base.ValidatorName(attribute, view)
}

// record returns the chosen type for attribute. A copied nested value may reuse
// a type when its pointer and default-value rules are the same.
func (s *wireAttributeScope) record(attribute *expr.AttributeExpr) *wireTypeRecord {
	if record := s.catalog.bindings[attribute]; record != nil {
		return record
	}
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return nil
	}
	preferred := wireTypeDeclaredName(userType.Origin())
	if s.policy.view != "" {
		preferred = wireTypeDeclaredName(userType)
	}
	preferred = codegen.Goify(preferred, true)
	return s.catalog.find(newWireTypeIdentity(attribute, wireAttribute, s.policy, preferred))
}

// unionRecord returns the chosen union for union. A copied nested union may
// reuse it when its pointer and default-value rules are the same.
func (s *wireAttributeScope) unionRecord(union *expr.Union) *wireUnionRecord {
	if record := s.catalog.unionBindings[union]; record != nil {
		return record
	}
	return s.catalog.findUnion(s.catalog.unionIdentity(union, wireAttribute, s.policy))
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

// wireUnionIdentitiesEqual reports whether two unions use the same definition
// and the same generated branch types.
func wireUnionIdentitiesEqual(left, right wireUnionIdentity) bool {
	return left.definition == right.definition && slices.Equal(left.declarations, right.declarations)
}

// newWireTypeIdentity records the designed value and the rules that determine
// its generated Go type.
func newWireTypeIdentity(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) wireTypeIdentity {
	identity := wireTypeIdentity{role: role, preferred: preferred, attribute: expr.DupAtt(attribute), policy: policy}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if resultType, ok := userType.(*expr.ResultTypeExpr); ok {
			identity.resultID = resultType.Identifier
			identity.role = 0
		} else if policy.view == "" {
			identity.sourceID = userType.Origin().ID()
			identity.role = 0
		}
	}
	return identity
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
func (i wireTypeIdentity) order(family string) wireNameOrder {
	return wireNameOrder{
		family:    family,
		source:    i.sourceID + i.resultID,
		role:      uint8(i.role),
		preferred: i.preferred,
		shape:     expr.Hash(i.attribute.Type, false, false, false),
		view:      i.policy.view,
		request:   i.policy.request,
		pointer:   i.policy.pointer,
		defaults:  i.policy.useDefault,
	}
}

// order returns the designed values used to choose stable suffixes for a union,
// its constants, and its functions.
func (i wireUnionIdentity) order(family, name, branch string) wireNameOrder {
	declarations := make([]string, len(i.declarations))
	for index, declaration := range i.declarations {
		order := declaration.identity.order("type")
		declarations[index] = fmt.Sprintf(
			"%q:%d:%q:%q:%q:%t:%t:%t",
			order.source,
			order.role,
			order.preferred,
			order.shape,
			order.view,
			order.request,
			order.pointer,
			order.defaults,
		)
	}
	return wireNameOrder{
		family:    family,
		source:    strings.Join(declarations, "\x00"),
		preferred: name,
		shape:     string(i.definition),
		view:      branch,
	}
}

// ComparePackageName orders HTTP declarations from designed values so memory
// addresses and design reading order cannot change generated names.
func (o wireNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(wireNameOrder)
	for _, compared := range []int{
		cmp.Compare(o.family, right.family),
		cmp.Compare(o.source, right.source),
		cmp.Compare(o.role, right.role),
		cmp.Compare(o.preferred, right.preferred),
		cmp.Compare(o.shape, right.shape),
		cmp.Compare(o.view, right.view),
		cmp.Compare(boolOrder(o.request), boolOrder(right.request)),
		cmp.Compare(boolOrder(o.pointer), boolOrder(right.pointer)),
		cmp.Compare(boolOrder(o.defaults), boolOrder(right.defaults)),
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
