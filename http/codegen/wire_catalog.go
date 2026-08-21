// This file owns declaration identity and names for the wire types emitted by
// one generated HTTP or JSON-RPC client or server package. The catalog first
// collects detached shapes, then freezes names, and only then lets analysis
// build declarations, references, and validators from those records.
package codegen

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// wireTypeCatalog owns the names and declarations emitted by one transport output package.
	wireTypeCatalog struct {
		scope            *codegen.NameScope
		records          []*wireTypeRecord
		unionOccurrences []wireUnionOccurrence
		unions           []*wireUnionRecord
		names            map[string]int
		frozen           bool
	}

	// wireUnionRecord owns one emitted positional union in this output package.
	wireUnionRecord struct {
		identity     wireUnionIdentity
		union        *expr.Union
		name         string
		kindName     string
		kindConsts   []string
		constructors []string
		data         *service.UnionTypeData
	}

	// wireUnionOccurrence records one union use until branch declarations have names.
	wireUnionOccurrence struct {
		union  *expr.Union
		role   wireTypeRole
		policy wireTypePolicy
	}

	// wireUnionIdentity combines the authored wire shape with the exact frozen
	// declarations referenced by its branches.
	wireUnionIdentity struct {
		definition   codegen.UnionTypeID
		declarations []*wireTypeRecord
	}

	// wireTypeRecord is the canonical package-local declaration selected for a wire identity.
	wireTypeRecord struct {
		identity wireTypeIdentity
		name     string
		ref      string
		data     *TypeData
	}

	// wireTypeIdentity contains typed declaration provenance and every policy
	// fact that changes the emitted Go type.
	wireTypeIdentity struct {
		source    expr.UserType
		resultID  string
		role      wireTypeRole
		preferred string
		attribute *expr.AttributeExpr
		policy    wireTypePolicy
	}

	// wireTypePolicy describes the pointer, default, validation, and view rules applied to a wire shape.
	wireTypePolicy struct {
		request    bool
		pointer    bool
		useDefault bool
		validate   bool
		view       string
	}

	// wireTypeRole identifies synthetic declarations that have no authored Origin.
	wireTypeRole uint8

	// wireAttributePair identifies two recursive attributes already compared.
	wireAttributePair struct {
		left  *expr.AttributeExpr
		right *expr.AttributeExpr
	}
)

const (
	wireRequestBody wireTypeRole = iota + 1
	wireResponseBody
	wireAttribute
	wireStreamPayload
)

// newWireTypeCatalog constructs an empty output-package catalog.
func newWireTypeCatalog(reserved ...string) *wireTypeCatalog {
	scope := codegen.NewNameScope()
	names := make(map[string]int, len(reserved))
	for _, name := range reserved {
		scope.Unique(name)
		names[name] = 1
	}
	return &wireTypeCatalog{scope: scope, names: names}
}

// collect records attribute and every named type it contains.
func (c *wireTypeCatalog) collect(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) *wireTypeRecord {
	if c.frozen {
		panic("cannot collect HTTP wire type after catalog freeze")
	}
	return c.collectRecursive(attribute, role, policy, preferred, make(map[expr.UserType]struct{}))
}

// Freeze assigns final names and binds copied user types to the catalog scope.
func (c *wireTypeCatalog) Freeze() {
	if c.frozen {
		return
	}
	for _, record := range c.records {
		record.name = c.uniqueName(record.identity.preferred)
		record.ref = wireTypeRef(record.name, record.identity.attribute.Type)
		setWireTypeName(record.identity.attribute, record.name)
		if userType, ok := record.identity.attribute.Type.(expr.UserType); ok {
			c.scope.HashedUnique(userType, record.name)
		} else {
			c.scope.Unique(record.name)
		}
	}
	for _, occurrence := range c.unionOccurrences {
		identity := c.unionIdentity(occurrence.union, occurrence.role, occurrence.policy)
		if c.findUnion(identity) == nil {
			c.unions = append(c.unions, &wireUnionRecord{identity: identity, union: occurrence.union})
		}
	}
	for _, union := range c.unions {
		union.name = c.uniqueName(codegen.Goify(union.union.Name(), true))
		union.kindName = c.uniqueName(union.name + "Kind")
		union.kindConsts = make([]string, len(union.union.Values))
		union.constructors = make([]string, len(union.union.Values))
		for index, branch := range union.union.Values {
			fieldName := codegen.Goify(branch.Name, true)
			union.kindConsts[index] = c.uniqueName(union.kindName + fieldName)
			union.constructors[index] = c.uniqueName("New" + union.name + fieldName)
		}
	}
	for _, union := range c.unions {
		c.applyUnionRecord(union.union, union)
		c.scope.HashedUnique(codegen.NewUnionTypeID(union.union), union.name)
		c.scope.Unique(union.kindName)
		for index := range union.kindConsts {
			c.scope.Unique(union.kindConsts[index])
			c.scope.Unique(union.constructors[index])
		}
	}
	for _, union := range c.unions {
		union.data = buildHTTPUnionTypeData(union.union, c.scope, union)
	}
	c.scope.Freeze()
	c.frozen = true
}

// wireTypeRef returns the Go reference owned by a frozen declaration record.
func wireTypeRef(name string, dataType expr.DataType) string {
	if _, inline := dataType.(*expr.Object); inline {
		return name
	}
	if expr.IsObject(dataType) || expr.IsUnion(dataType) {
		return "*" + name
	}
	return name
}

// lookup returns the frozen record and applies its name to an equivalent occurrence.
func (c *wireTypeCatalog) lookup(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) *wireTypeRecord {
	if !c.frozen {
		panic("cannot resolve HTTP wire type before catalog freeze")
	}
	identity := newWireTypeIdentity(attribute, role, policy, preferred)
	record := c.find(identity)
	if record != nil {
		setWireTypeName(attribute, record.name)
		return record
	}
	panic(fmt.Sprintf("HTTP wire type %q was not collected before catalog freeze", preferred))
}

// lookupUser returns the frozen record for a named user type and nil for an
// inline or primitive occurrence that has no top-level declaration.
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

// applyNames writes every frozen nested declaration name onto one detached
// occurrence before type definitions or transforms traverse it.
func (c *wireTypeCatalog) applyNames(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy) {
	c.applyNamesRecursive(attribute, role, policy, make(map[expr.UserType]struct{}))
}

// applyNamesRecursive follows named fields once per authored origin.
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
			panic(fmt.Sprintf("HTTP union %q was not collected before catalog freeze", actual.Name()))
		}
		c.applyUnionRecord(actual, record)
	}
}

// unionTypes returns the frozen union declarations in deterministic name order.
func (c *wireTypeCatalog) unionTypes() []*service.UnionTypeData {
	unions := make([]*service.UnionTypeData, len(c.unions))
	for index, record := range c.unions {
		unions[index] = record.data
	}
	slices.SortFunc(unions, func(left, right *service.UnionTypeData) int { return strings.Compare(left.Name, right.Name) })
	return unions
}

// bind attaches occurrence-specific TypeData to its canonical declaration and
// merges the validator generated by any occurrence of that declaration.
func (c *wireTypeCatalog) bind(record *wireTypeRecord, data *TypeData) *TypeData {
	data.declaration = record
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
			panic(fmt.Sprintf("HTTP wire type %q produced conflicting declarations", record.name))
		}
	}
	if data.ValidateDef != "" {
		if record.data.ValidateDef == "" {
			record.data.ValidateDef = data.ValidateDef
			record.data.ValidateRef = data.ValidateRef
		} else if record.data.ValidateDef != data.ValidateDef || record.data.ValidateRef != data.ValidateRef {
			panic(fmt.Sprintf("HTTP wire type %q produced conflicting validators", record.name))
		}
	}
	return data
}

// collectRecursive records named declarations and terminates cycles by source Origin.
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

// findOrAppend reuses a structurally equal typed record or appends a new one.
func (c *wireTypeCatalog) findOrAppend(identity wireTypeIdentity) *wireTypeRecord {
	if record := c.find(identity); record != nil {
		return record
	}
	record := &wireTypeRecord{identity: identity}
	c.records = append(c.records, record)
	return record
}

// find returns the declaration record equal to identity.
func (c *wireTypeCatalog) find(identity wireTypeIdentity) *wireTypeRecord {
	for _, record := range c.records {
		if wireTypeIdentitiesEqual(record.identity, identity) {
			return record
		}
	}
	return nil
}

// unionIdentity resolves every named declaration referenced by union without
// changing the detached occurrence.
func (c *wireTypeCatalog) unionIdentity(union *expr.Union, role wireTypeRole, policy wireTypePolicy) wireUnionIdentity {
	identity := wireUnionIdentity{definition: codegen.NewUnionTypeID(union)}
	attribute := &expr.AttributeExpr{Type: union}
	c.collectUnionDeclarations(attribute, role, policy, &identity.declarations, make(map[expr.UserType]struct{}))
	return identity
}

// collectUnionDeclarations records package declarations in branch traversal order.
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
			panic(fmt.Sprintf("HTTP union branch type %q was not collected before catalog freeze", preferred))
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

// applyUnionRecord writes exactly the branch declarations captured by record
// onto one equivalent union occurrence.
func (c *wireTypeCatalog) applyUnionRecord(union *expr.Union, record *wireUnionRecord) {
	union.TypeName = record.name
	index := 0
	seen := make(map[expr.UserType]struct{})
	for _, branch := range union.Values {
		c.applyResolvedDeclarations(branch.Attribute, record.identity.declarations, &index, seen)
	}
	if index != len(record.identity.declarations) {
		panic(fmt.Sprintf("HTTP union %q did not consume its frozen branch declarations", record.name))
	}
}

// applyResolvedDeclarations consumes the typed declaration sequence captured
// while the union identity was built.
func (c *wireTypeCatalog) applyResolvedDeclarations(attribute *expr.AttributeExpr, declarations []*wireTypeRecord, index *int, seen map[expr.UserType]struct{}) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if *index >= len(declarations) {
			panic(fmt.Sprintf("HTTP union branch %q has no frozen declaration", wireTypeDeclaredName(userType)))
		}
		record := declarations[*index]
		*index = *index + 1
		setWireTypeName(attribute, record.name)
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
			panic(fmt.Sprintf("HTTP nested union %q has no frozen declaration", actual.Name()))
		}
		actual.TypeName = record.name
	}
}

// findUnion returns the package union record equal to identity.
func (c *wireTypeCatalog) findUnion(identity wireUnionIdentity) *wireUnionRecord {
	for _, record := range c.unions {
		if wireUnionIdentitiesEqual(record.identity, identity) {
			return record
		}
	}
	return nil
}

// wireUnionIdentitiesEqual compares the typed declaration sequence referenced by a wire union.
func wireUnionIdentitiesEqual(left, right wireUnionIdentity) bool {
	return left.definition == right.definition && slices.Equal(left.declarations, right.declarations)
}

// uniqueName allocates a package declaration without creating a second scope identity.
func (c *wireTypeCatalog) uniqueName(preferred string) string {
	count := c.names[preferred]
	if count == 0 {
		c.names[preferred] = 1
		return preferred
	}
	for index := count + 1; ; index++ {
		name := preferred + strconv.Itoa(index)
		if c.names[name] == 0 {
			c.names[preferred] = index
			c.names[name] = 1
			return name
		}
	}
}

// newWireTypeIdentity builds a typed identity from a detached occurrence.
func newWireTypeIdentity(attribute *expr.AttributeExpr, role wireTypeRole, policy wireTypePolicy, preferred string) wireTypeIdentity {
	identity := wireTypeIdentity{role: role, preferred: preferred, attribute: expr.DupAtt(attribute), policy: policy}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		if resultType, ok := userType.(*expr.ResultTypeExpr); ok {
			identity.resultID = resultType.Identifier
			identity.role = 0
		} else if policy.view == "" {
			identity.source = userType.Origin()
			identity.role = 0
		}
	}
	return identity
}

// wireTypeIdentitiesEqual compares provenance, policy, and the detached attribute contract.
func wireTypeIdentitiesEqual(left, right wireTypeIdentity) bool {
	if left.source != right.source || left.resultID != right.resultID || left.role != right.role || left.preferred != right.preferred || !wireTypePoliciesEqual(left.policy, right.policy) {
		return false
	}
	if left.source != nil {
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

// wireTypePoliciesEqual compares only facts that change a Go declaration.
// Validation helpers are separate package records and do not create a second
// type when the wire representation is otherwise identical.
func wireTypePoliciesEqual(left, right wireTypePolicy) bool {
	left.validate = false
	right.validate = false
	return left == right
}

// wireAttributesEqual compares facts that change a declaration or validator and terminates cycles.
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

// wireMetadataEqual compares authored metadata while ignoring the name written during Freeze.
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

// sortedWireAttributes makes declaration allocation independent of authored object field order.
func sortedWireAttributes(attributes []*expr.NamedAttributeExpr) []*expr.NamedAttributeExpr {
	sorted := slices.Clone(attributes)
	slices.SortFunc(sorted, func(left, right *expr.NamedAttributeExpr) int {
		return strings.Compare(left.Name, right.Name)
	})
	return sorted
}

// setWireTypeName records the package-owned name on a detached user type.
func setWireTypeName(attribute *expr.AttributeExpr, name string) {
	if attribute.Type == expr.Empty {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		userType.Attribute().AddMeta("struct:type:name", name)
	}
}

// wireTypeDeclaredName returns the stable expression declaration name rather
// than a package name previously assigned through struct:type:name metadata.
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
