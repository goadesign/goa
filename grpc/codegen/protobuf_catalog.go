// This file owns the protobuf declarations and validation helpers emitted by
// one generated gRPC protobuf package. It separates declaration identity from
// traversal identity and freezes names before conversion data refers to them.
package codegen

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// protobufPackageCatalog owns every protobuf message and validator emitted
	// into one generated protobuf package.
	protobufPackageCatalog struct {
		packageName       string
		messages          []*protobufMessageRecord
		messageUses       map[*expr.AttributeExpr]*protobufMessageRecord
		unions            []*protobufUnionRecord
		unionUses         map[*expr.AttributeExpr]*protobufUnionRecord
		syntheticSources  map[expr.UserType]protobufMessageSource
		reservedNames     []string
		validators        []*protobufValidationRecord
		frozen            bool
		messagesRendered  bool
		validationsFrozen bool
	}

	// protobufEndpointMessages contains every detached protobuf-shaped value for
	// one endpoint before conversion and render records are built.
	protobufEndpointMessages struct {
		request          *expr.AttributeExpr
		streamingRequest *expr.AttributeExpr
		requestEnvelope  *expr.AttributeExpr
		response         *expr.AttributeExpr
		errors           map[string]*expr.AttributeExpr
	}

	// protobufMessageRecord is the canonical declaration selected for one typed
	// protobuf wire contract.
	protobufMessageRecord struct {
		identity protobufMessageIdentity
		uses     []*expr.AttributeExpr
		name     string
		goRef    string
		data     *service.UserTypeData
	}

	// protobufMessageIdentity contains the source declaration and the wire facts
	// that can change the emitted protobuf message.
	protobufMessageIdentity struct {
		source        protobufMessageSource
		preferredName string
		explicitName  bool
		userType      expr.UserType
		attribute     *expr.AttributeExpr
	}

	// protobufUnionRecord identifies one oneof declaration nested in an owning
	// protobuf message. Generated transformation helpers use this typed owner
	// instead of allocating a name during lookup.
	protobufUnionRecord struct {
		owner     *protobufMessageRecord
		attribute *expr.AttributeExpr
		fieldName string
		uses      []*expr.AttributeExpr
		name      string
	}

	// protobufMessageSource identifies either an authored declaration or a
	// synthetic endpoint message whose declaration does not exist in the design.
	protobufMessageSource struct {
		origin    expr.UserType
		synthetic protobufSyntheticMessage
	}

	// protobufSyntheticMessage identifies a compiler-created endpoint message.
	protobufSyntheticMessage struct {
		endpoint *expr.GRPCEndpointExpr
		error    *expr.GRPCErrorExpr
		role     protobufSyntheticRole
	}

	// protobufSyntheticRole identifies which endpoint wire value a synthetic
	// protobuf message represents.
	protobufSyntheticRole uint8

	// protobufValidationRecord is the canonical validation helper emitted in one
	// generated client or server package.
	protobufValidationRecord struct {
		declaration *protobufMessageRecord
		attribute   *expr.AttributeExpr
		side        validateKind
		targetName  string
		contextName string
		uses        []*expr.AttributeExpr
		name        string
		data        *ValidationData
	}

	// protobufAttributePair breaks cycles while comparing two typed wire or
	// validation graphs.
	protobufAttributePair struct {
		left  *expr.AttributeExpr
		right *expr.AttributeExpr
	}

	// protobufValidationScope resolves nested validation calls through frozen
	// validator records while delegating fields and type references to protobuf.
	protobufValidationScope struct {
		*protoBufScope
		catalog *protobufPackageCatalog
		side    validateKind
	}
)

const (
	protobufRequestMessage protobufSyntheticRole = iota + 1
	protobufStreamingRequestMessage
	protobufStreamEnvelopeMessage
	protobufResponseMessage
	protobufErrorMessage
	protobufWrapperMessage
)

// newProtobufPackageCatalog constructs the declaration owner for one actual
// generated protobuf package.
func newProtobufPackageCatalog(packageName string) *protobufPackageCatalog {
	return &protobufPackageCatalog{
		packageName:      packageName,
		messageUses:      make(map[*expr.AttributeExpr]*protobufMessageRecord),
		unionUses:        make(map[*expr.AttributeExpr]*protobufUnionRecord),
		syntheticSources: make(map[expr.UserType]protobufMessageSource),
	}
}

// reserveName prevents message declarations from colliding with another
// package-level protobuf declaration such as the service interface.
func (c *protobufPackageCatalog) reserveName(name string) {
	if c.frozen {
		panic("cannot reserve a protobuf package name after the catalog freezes")
	}
	c.reservedNames = append(c.reservedNames, name)
}

// bindSyntheticSource associates a shaped synthetic declaration and all of
// its later copies with the endpoint role that created it.
func (c *protobufPackageCatalog) bindSyntheticSource(attribute *expr.AttributeExpr, source protobufMessageSource) {
	userType, ok := attribute.Type.(expr.UserType)
	if !ok || source.synthetic.role == 0 {
		return
	}
	c.syntheticSources[userType.Origin()] = source
}

// collectMessage records every protobuf declaration reachable from attribute.
// source identifies a synthetic root; nested authored declarations retain
// their own origins.
func (c *protobufPackageCatalog) collectMessage(attribute *expr.AttributeExpr, source protobufMessageSource, sd *ServiceData) []string {
	if c.frozen {
		panic("cannot collect a protobuf message after the package catalog is frozen")
	}
	return c.collectMessageRecursive(attribute, source, true, nil, "", sd)
}

// freezeMessages assigns every declaration its final protobuf and generated Go
// names, binds all occurrences, and builds immutable template records.
func (c *protobufPackageCatalog) freezeMessages(sd *ServiceData) []*service.UserTypeData {
	c.freezeMessageNames()
	if c.messagesRendered {
		return c.messageData()
	}
	c.messagesRendered = true
	for _, record := range c.messages {
		identity := record.identity
		userType := identity.userType
		definition := protoBufMessageDef(userTypeAttribute(userType), sd)
		for _, use := range record.uses[1:] {
			other := protoBufMessageDef(userTypeAttribute(use.Type.(expr.UserType)), sd)
			if other != definition {
				panic(fmt.Sprintf("protobuf declaration %q has one typed identity but different wire definitions", record.name))
			}
		}
		record.data = &service.UserTypeData{
			Name:        record.name,
			VarName:     record.name,
			Description: userType.Attribute().Description,
			Def:         definition,
			Ref:         record.goRef,
			Type:        userType,
		}
	}
	return c.messageData()
}

// freezeMessageNames assigns every declaration its final package-level name
// without rendering .proto definitions. Transformation-only consumers use
// this phase because they need references but do not emit declarations.
func (c *protobufPackageCatalog) freezeMessageNames() {
	if c.frozen {
		return
	}
	c.frozen = true
	used := make(map[string]struct{}, len(c.messages))
	counts := make(map[string]int, len(c.messages))
	for _, name := range c.reservedNames {
		used[name] = struct{}{}
		counts[name] = 1
	}
	for _, record := range c.messages {
		record.name = uniqueProtobufName(record.identity.preferredName, used, counts)
		record.goRef = "*" + c.packageName + "." + record.name
	}
	for _, record := range c.unions {
		record.name = record.owner.name + "_" + protoBufify(record.fieldName, true, true)
	}
}

// collectValidation records the validation helper needed for attribute and all
// nested protobuf message declarations on one generated side.
func (c *protobufPackageCatalog) collectValidation(attribute *expr.AttributeExpr, side validateKind, targetName, contextName string) {
	if !c.frozen {
		panic("cannot collect protobuf validators before message declarations freeze")
	}
	if c.validationsFrozen {
		panic("cannot collect a protobuf validator after validators freeze")
	}
	c.collectValidationRecursive(attribute, side, targetName, contextName, make(map[*protobufValidationRecord]struct{}))
}

// freezeValidations assigns helper names independently in the generated client
// and server packages, then renders definitions through those frozen records.
func (c *protobufPackageCatalog) freezeValidations(sd *ServiceData) []*ValidationData {
	if c.validationsFrozen {
		return c.validationData()
	}
	c.validationsFrozen = true
	used := map[validateKind]map[string]struct{}{
		validateServer: {},
		validateClient: {},
	}
	counts := map[validateKind]map[string]int{
		validateServer: {},
		validateClient: {},
	}
	for _, record := range c.validators {
		base := "Validate" + record.declaration.name
		record.name = uniqueProtobufName(base, used[record.side], counts[record.side])
	}
	for _, record := range c.validators {
		validationAttribute := expr.DupAtt(record.attribute)
		c.bindEquivalentMessageUses(record.attribute, validationAttribute, make(map[protobufAttributePair]struct{}))
		removeMeta(validationAttribute)
		userType := validationAttribute.Type.(expr.UserType)
		context := protoBufTypeContext(c.packageName, sd, false)
		context.Scope = &protobufValidationScope{
			protoBufScope: context.Scope.(*protoBufScope),
			catalog:       c,
			side:          record.side,
		}
		definition := codegen.AttributeValidationCode(
			userTypeAttribute(userType),
			userType,
			context,
			true,
			false,
			record.targetName,
			record.contextName,
		)
		if definition == "" {
			continue
		}
		record.data = &ValidationData{
			Name:    record.name,
			Def:     definition,
			ArgName: record.targetName,
			SrcName: record.declaration.name,
			SrcRef:  record.declaration.goRef,
			Kind:    record.side,
		}
	}
	return c.validationData()
}

// message returns the frozen declaration bound to attribute.
func (c *protobufPackageCatalog) message(attribute *expr.AttributeExpr) *protobufMessageRecord {
	if !c.frozen {
		panic("cannot resolve a protobuf message before the package catalog freezes")
	}
	if record := c.messageUses[attribute]; record != nil {
		return record
	}
	if _, ok := attribute.Type.(expr.UserType); !ok {
		return nil
	}
	userType := attribute.Type.(expr.UserType)
	source := protobufMessageSource{origin: userType.Origin()}
	if synthetic, ok := c.syntheticSources[userType.Origin()]; ok {
		source = synthetic
	}
	identity := protobufMessageIdentityFor(attribute, source)
	for _, record := range c.messages {
		if sameProtobufMessageIdentity(record.identity, identity) {
			return record
		}
	}
	if len(userType.Attribute().Meta[wrappedAttrMeta]) > 0 {
		identity.source = protobufMessageSource{synthetic: protobufSyntheticMessage{role: protobufWrapperMessage}}
		for _, record := range c.messages {
			if sameProtobufMessageIdentity(record.identity, identity) {
				return record
			}
		}
	}
	return nil
}

// unionName returns the frozen helper identity for one oneof declaration.
func (c *protobufPackageCatalog) unionName(attribute *expr.AttributeExpr) string {
	if !c.frozen {
		panic("cannot resolve a protobuf oneof before the package catalog freezes")
	}
	record := c.unionUses[attribute]
	if record == nil {
		for _, candidate := range c.unions {
			if !sameProtobufWireAttribute(candidate.attribute, attribute, make(map[protobufAttributePair]struct{})) {
				continue
			}
			if record != nil && record != candidate {
				panic(fmt.Sprintf("protobuf oneof %q matches multiple frozen declarations", attribute.Type.Name()))
			}
			record = candidate
		}
	}
	if record == nil {
		panic(fmt.Sprintf("protobuf oneof %q has no frozen declaration", attribute.Type.Name()))
	}
	return record.name
}

// validation returns the frozen validator bound to attribute on side.
func (c *protobufPackageCatalog) validation(attribute *expr.AttributeExpr, side validateKind) *ValidationData {
	if !c.validationsFrozen {
		panic("cannot resolve a protobuf validator before validators freeze")
	}
	declaration := c.message(attribute)
	if declaration == nil {
		return nil
	}
	for _, record := range c.validators {
		if record.side == side && record.declaration == declaration &&
			sameProtobufValidationAttribute(record.attribute, attribute, make(map[protobufAttributePair]struct{})) {
			return record.data
		}
	}
	return nil
}

// Name returns the frozen protobuf declaration name, or the validator-specific
// source name while validation code is being rendered.
func (s *protobufValidationScope) Name(attribute *expr.AttributeExpr, pkg string, pointer, useDefault bool) string {
	if validator := s.catalog.validationRecord(attribute, s.side); validator != nil {
		return strings.TrimPrefix(validator.name, "Validate")
	}
	return s.protoBufScope.Name(attribute, pkg, pointer, useDefault)
}

// collectMessageRecursive gathers imports and declarations while using record
// identity itself as the cycle guard.
func (c *protobufPackageCatalog) collectMessageRecursive(attribute *expr.AttributeExpr, source protobufMessageSource, root bool, owner *protobufMessageRecord, fieldName string, sd *ServiceData) []string {
	if attribute == nil {
		return nil
	}
	imports := protobufAttributeImports(attribute, sd)
	if expr.IsPrimitive(attribute.Type) {
		if attribute.Type.Kind() == expr.AnyKind {
			imports = append(imports, "google/protobuf/struct.proto")
		}
		return imports
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		origin := actual.Origin()
		identitySource := protobufMessageSource{origin: origin}
		if synthetic, ok := c.syntheticSources[origin]; ok {
			identitySource = synthetic
		}
		if !root && len(actual.Attribute().Meta[wrappedAttrMeta]) > 0 {
			identitySource = protobufMessageSource{synthetic: protobufSyntheticMessage{
				role: protobufWrapperMessage,
			}}
		}
		if root && source.synthetic.role != 0 {
			identitySource = source
			c.syntheticSources[origin] = source
		}
		identity := protobufMessageIdentityFor(attribute, identitySource)
		record := c.findMessage(identity)
		if record != nil {
			record.uses = append(record.uses, attribute)
			c.messageUses[attribute] = record
			c.bindEquivalentMessageUses(record.uses[0], attribute, make(map[protobufAttributePair]struct{}))
			return imports
		}
		record = &protobufMessageRecord{identity: identity, uses: []*expr.AttributeExpr{attribute}}
		c.messages = append(c.messages, record)
		c.messageUses[attribute] = record
		imports = append(imports, c.collectMessageRecursive(userTypeAttribute(actual), protobufMessageSource{}, false, record, "", sd)...)
	case *expr.Object:
		for _, named := range *actual {
			imports = append(imports, c.collectMessageRecursive(named.Attribute, protobufMessageSource{}, false, owner, named.Name, sd)...)
		}
	case *expr.Array:
		imports = append(imports, c.collectMessageRecursive(actual.ElemType, protobufMessageSource{}, false, owner, fieldName+"Elem", sd)...)
	case *expr.Map:
		imports = append(imports, c.collectMessageRecursive(actual.KeyType, protobufMessageSource{}, false, owner, fieldName+"Key", sd)...)
		imports = append(imports, c.collectMessageRecursive(actual.ElemType, protobufMessageSource{}, false, owner, fieldName+"Elem", sd)...)
	case *expr.Union:
		if owner == nil {
			panic(fmt.Sprintf("protobuf oneof %q has no owning message", actual.Name()))
		}
		if fieldName == "" {
			fieldName = actual.Name()
		}
		record := c.findUnion(owner, fieldName, attribute)
		if record == nil {
			record = &protobufUnionRecord{
				owner:     owner,
				attribute: attribute,
				fieldName: fieldName,
			}
			c.unions = append(c.unions, record)
		}
		record.uses = append(record.uses, attribute)
		c.unionUses[attribute] = record
		for _, named := range actual.Values {
			imports = append(imports, c.collectMessageRecursive(named.Attribute, protobufMessageSource{}, false, owner, fieldName+named.Name, sd)...)
		}
	}
	return imports
}

// bindEquivalentMessageUses associates every nested declaration occurrence in
// a reused wire graph with the canonical records already collected for the
// first occurrence.
func (c *protobufPackageCatalog) bindEquivalentMessageUses(canonical, duplicate *expr.AttributeExpr, seen map[protobufAttributePair]struct{}) {
	pair := protobufAttributePair{left: canonical, right: duplicate}
	if _, ok := seen[pair]; ok {
		return
	}
	seen[pair] = struct{}{}
	switch canonicalType := canonical.Type.(type) {
	case expr.UserType:
		if record := c.messageUses[canonical]; record != nil {
			if c.messageUses[duplicate] == nil {
				c.messageUses[duplicate] = record
				record.uses = append(record.uses, duplicate)
			}
		}
		duplicateType := duplicate.Type.(expr.UserType)
		c.bindEquivalentMessageUses(userTypeAttribute(canonicalType), userTypeAttribute(duplicateType), seen)
	case *expr.Object:
		duplicateType := duplicate.Type.(*expr.Object)
		for index, named := range *canonicalType {
			c.bindEquivalentMessageUses(named.Attribute, (*duplicateType)[index].Attribute, seen)
		}
	case *expr.Array:
		c.bindEquivalentMessageUses(canonicalType.ElemType, duplicate.Type.(*expr.Array).ElemType, seen)
	case *expr.Map:
		duplicateType := duplicate.Type.(*expr.Map)
		c.bindEquivalentMessageUses(canonicalType.KeyType, duplicateType.KeyType, seen)
		c.bindEquivalentMessageUses(canonicalType.ElemType, duplicateType.ElemType, seen)
	case *expr.Union:
		if record := c.unionUses[canonical]; record != nil {
			c.unionUses[duplicate] = record
			record.uses = append(record.uses, duplicate)
		}
		duplicateType := duplicate.Type.(*expr.Union)
		for index, named := range canonicalType.Values {
			c.bindEquivalentMessageUses(named.Attribute, duplicateType.Values[index].Attribute, seen)
		}
	}
}

// findUnion returns the oneof declaration with the same owning message, field,
// and typed wire schema.
func (c *protobufPackageCatalog) findUnion(owner *protobufMessageRecord, fieldName string, attribute *expr.AttributeExpr) *protobufUnionRecord {
	for _, record := range c.unions {
		if record.owner == owner && record.fieldName == fieldName &&
			sameProtobufWireAttribute(record.attribute, attribute, make(map[protobufAttributePair]struct{})) {
			return record
		}
	}
	return nil
}

// collectValidationRecursive declares one validator per message, rule graph,
// and generated side, then descends to the nested declarations it may call.
func (c *protobufPackageCatalog) collectValidationRecursive(attribute *expr.AttributeExpr, side validateKind, targetName, contextName string, seen map[*protobufValidationRecord]struct{}) {
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if expr.IsPrimitive(actual) {
			return
		}
		declaration := c.message(attribute)
		if declaration == nil {
			panic(fmt.Sprintf("no protobuf declaration collected for validation type %q", actual.Name()))
		}
		record := c.findValidation(declaration, attribute, side)
		if record == nil {
			record = &protobufValidationRecord{
				declaration: declaration,
				attribute:   attribute,
				side:        side,
				targetName:  targetName,
				contextName: contextName,
				uses:        []*expr.AttributeExpr{attribute},
			}
			c.validators = append(c.validators, record)
		} else {
			record.uses = append(record.uses, attribute)
		}
		if _, ok := seen[record]; ok {
			return
		}
		seen[record] = struct{}{}
		c.collectValidationRecursive(userTypeAttribute(actual), side, targetName, contextName, seen)
	case *expr.Object:
		for _, named := range *actual {
			c.collectValidationRecursive(named.Attribute, side, codegen.Goify(named.Name, false), named.Name, seen)
		}
	case *expr.Array:
		c.collectValidationRecursive(actual.ElemType, side, "elem", "elem", seen)
	case *expr.Map:
		c.collectValidationRecursive(actual.KeyType, side, "key", "key", seen)
		c.collectValidationRecursive(actual.ElemType, side, "val", "val", seen)
	case *expr.Union:
		for _, named := range actual.Values {
			c.collectValidationRecursive(named.Attribute, side, codegen.Goify(named.Name, false), named.Name, seen)
		}
	}
}

// findMessage returns the existing declaration with the same typed identity.
func (c *protobufPackageCatalog) findMessage(identity protobufMessageIdentity) *protobufMessageRecord {
	for _, record := range c.messages {
		if sameProtobufMessageIdentity(record.identity, identity) {
			return record
		}
	}
	return nil
}

// findValidation returns the existing validator with the same declaration,
// validation contract, and generated side.
func (c *protobufPackageCatalog) findValidation(declaration *protobufMessageRecord, attribute *expr.AttributeExpr, side validateKind) *protobufValidationRecord {
	for _, record := range c.validators {
		if record.declaration == declaration && record.side == side &&
			sameProtobufValidationAttribute(record.attribute, attribute, make(map[protobufAttributePair]struct{})) {
			return record
		}
	}
	return nil
}

// validationRecord resolves the validator called for a nested message use.
func (c *protobufPackageCatalog) validationRecord(attribute *expr.AttributeExpr, side validateKind) *protobufValidationRecord {
	declaration := c.message(attribute)
	if declaration == nil {
		return nil
	}
	return c.findValidation(declaration, attribute, side)
}

// messageData returns only declarations with completed immutable render data.
func (c *protobufPackageCatalog) messageData() []*service.UserTypeData {
	data := make([]*service.UserTypeData, 0, len(c.messages))
	for _, record := range c.messages {
		if record.data != nil {
			data = append(data, record.data)
		}
	}
	return data
}

// validationData returns only validators whose typed rules emit code.
func (c *protobufPackageCatalog) validationData() []*ValidationData {
	data := make([]*ValidationData, 0, len(c.validators))
	for _, record := range c.validators {
		if record.data != nil {
			data = append(data, record.data)
		}
	}
	return data
}

// protobufMessageIdentityFor derives message identity without consulting a
// naming scope or rendered source.
func protobufMessageIdentityFor(attribute *expr.AttributeExpr, source protobufMessageSource) protobufMessageIdentity {
	userType := attribute.Type.(expr.UserType)
	if source.origin == nil && source.synthetic.role == 0 {
		source.origin = userType.Origin()
	}
	preferred := protoBufify(userType.Name(), true, true)
	explicit := false
	names := attribute.Meta["struct:name:proto"]
	if len(names) == 0 {
		names = userType.Attribute().Meta["struct:name:proto"]
	}
	if len(names) > 0 {
		preferred = names[0]
		explicit = true
	}
	return protobufMessageIdentity{
		source:        source,
		preferredName: preferred,
		explicitName:  explicit,
		userType:      userType,
		attribute:     userTypeAttribute(userType),
	}
}

// sameProtobufMessageIdentity compares the typed source and every protobuf
// schema fact rather than expression hashes or generated names.
func sameProtobufMessageIdentity(left, right protobufMessageIdentity) bool {
	if left.preferredName != right.preferredName || left.explicitName != right.explicitName {
		return false
	}
	if left.source != right.source {
		// An explicit protobuf name declares intentional reuse across endpoint
		// roles or authored origins when the complete wire schemas also match.
		if !left.explicitName {
			return false
		}
	}
	return sameProtobufWireAttribute(left.attribute, right.attribute, make(map[protobufAttributePair]struct{}))
}

// sameProtobufWireAttribute compares facts that affect a protobuf declaration.
func sameProtobufWireAttribute(left, right *expr.AttributeExpr, seen map[protobufAttributePair]struct{}) bool {
	if left == right {
		return true
	}
	if left == nil || right == nil || left.Description != right.Description {
		return false
	}
	pair := protobufAttributePair{left: left, right: right}
	if _, ok := seen[pair]; ok {
		return true
	}
	seen[pair] = struct{}{}
	if !sameProtobufMeta(left.Meta, right.Meta) {
		return false
	}
	if leftObject, rightObject := expr.AsObject(left.Type), expr.AsObject(right.Type); leftObject != nil && rightObject != nil {
		for _, named := range *leftObject {
			if expr.IsPrimitive(named.Attribute.Type) && left.IsRequired(named.Name) != right.IsRequired(named.Name) {
				return false
			}
		}
	}
	return sameProtobufWireType(left.Type, right.Type, seen)
}

// sameProtobufWireType compares protobuf-native type, ordered field, oneof,
// collection, and nested declaration contracts.
func sameProtobufWireType(left, right expr.DataType, seen map[protobufAttributePair]struct{}) bool {
	if left.Kind() != right.Kind() {
		return false
	}
	switch left := left.(type) {
	case expr.Primitive:
		return left == right.(expr.Primitive)
	case expr.UserType:
		right := right.(expr.UserType)
		sameSource := left.Origin() == right.Origin()
		bothSyntheticWrappers := len(left.Attribute().Meta[wrappedAttrMeta]) > 0 &&
			len(right.Attribute().Meta[wrappedAttrMeta]) > 0
		return (sameSource || bothSyntheticWrappers) && sameProtobufWireAttribute(left.Attribute(), right.Attribute(), seen)
	case *expr.Object:
		right := right.(*expr.Object)
		if len(*left) != len(*right) {
			return false
		}
		for index, named := range *left {
			other := (*right)[index]
			if named.Name != other.Name || !sameProtobufWireAttribute(named.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case *expr.Array:
		right := right.(*expr.Array)
		return sameProtobufWireAttribute(left.ElemType, right.ElemType, seen)
	case *expr.Map:
		right := right.(*expr.Map)
		return sameProtobufWireAttribute(left.KeyType, right.KeyType, seen) && sameProtobufWireAttribute(left.ElemType, right.ElemType, seen)
	case *expr.Union:
		right := right.(*expr.Union)
		if left.TypeName != right.TypeName || left.TypeKey != right.TypeKey || left.ValueKey != right.ValueKey || len(left.Values) != len(right.Values) {
			return false
		}
		for index, named := range left.Values {
			other := right.Values[index]
			if named.Name != other.Name || !sameProtobufWireAttribute(named.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("unknown protobuf wire type %T", left))
	}
}

// sameProtobufValidationAttribute compares typed validation provenance and
// rules independently from protobuf wire declaration identity.
func sameProtobufValidationAttribute(left, right *expr.AttributeExpr, seen map[protobufAttributePair]struct{}) bool {
	if left == right {
		return true
	}
	if left == nil || right == nil || !reflect.DeepEqual(left.Validation, right.Validation) ||
		!reflect.DeepEqual(left.DefaultValue, right.DefaultValue) ||
		!reflect.DeepEqual(left.Meta["struct:field:type"], right.Meta["struct:field:type"]) {
		return false
	}
	pair := protobufAttributePair{left: left, right: right}
	if _, ok := seen[pair]; ok {
		return true
	}
	seen[pair] = struct{}{}
	if left.Type.Kind() != right.Type.Kind() {
		return false
	}
	switch left := left.Type.(type) {
	case expr.Primitive:
		return left == right.Type.(expr.Primitive)
	case expr.UserType:
		right := right.Type.(expr.UserType)
		return left.Origin() == right.Origin() && sameProtobufValidationAttribute(left.Attribute(), right.Attribute(), seen)
	case *expr.Object:
		right := right.Type.(*expr.Object)
		if len(*left) != len(*right) {
			return false
		}
		for index, named := range *left {
			other := (*right)[index]
			if named.Name != other.Name || !sameProtobufValidationAttribute(named.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case *expr.Array:
		right := right.Type.(*expr.Array)
		return left.NonNullableElems == right.NonNullableElems && sameProtobufValidationAttribute(left.ElemType, right.ElemType, seen)
	case *expr.Map:
		right := right.Type.(*expr.Map)
		return sameProtobufValidationAttribute(left.KeyType, right.KeyType, seen) && sameProtobufValidationAttribute(left.ElemType, right.ElemType, seen)
	case *expr.Union:
		right := right.Type.(*expr.Union)
		if len(left.Values) != len(right.Values) {
			return false
		}
		for index, named := range left.Values {
			other := right.Values[index]
			if named.Name != other.Name || !sameProtobufValidationAttribute(named.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("unknown protobuf validation type %T", left))
	}
}

// protobufAttributeImports returns imports declared directly on attribute.
func protobufAttributeImports(attribute *expr.AttributeExpr, sd *ServiceData) []string {
	proto := attribute.Meta["struct:field:proto"]
	if len(proto) <= 1 {
		return nil
	}
	protobufImport := proto[1]
	for _, spec := range sd.Service.ProtoImports {
		if spec.Path == protobufImport {
			return nil
		}
	}
	if len(proto) > 3 {
		elements := strings.Split(proto[3], "/")
		sd.Service.ProtoImports = append(sd.Service.ProtoImports, &codegen.ImportSpec{
			Path: proto[3],
			Name: elements[len(elements)-1],
		})
	}
	return []string{protobufImport}
}

// uniqueProtobufName reserves one deterministic package-level identifier.
func uniqueProtobufName(base string, used map[string]struct{}, counts map[string]int) string {
	if _, ok := used[base]; !ok {
		used[base] = struct{}{}
		counts[base] = 1
		return base
	}
	for index := counts[base] + 1; ; index++ {
		candidate := base + strconv.Itoa(index)
		if _, ok := used[candidate]; ok {
			continue
		}
		used[candidate] = struct{}{}
		counts[base] = index
		return candidate
	}
}

// sameProtobufMeta compares metadata that changes protobuf field numbers,
// external types, explicit names, wrapper layout, or JSON names.
func sameProtobufMeta(left, right expr.MetaExpr) bool {
	for _, name := range []string{
		"rpc:tag",
		"struct:field:proto",
		"struct:name:proto",
		"proto:tag:json",
		wrappedAttrMeta,
	} {
		if !reflect.DeepEqual(left[name], right[name]) {
			return false
		}
	}
	return true
}
