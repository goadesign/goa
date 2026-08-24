// This file records the protobuf messages and validation functions written for
// one gRPC service. It chooses every package-level name before conversion code
// refers to that name.
package codegen

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

type (
	// protobufPackageCatalog stores every message and validation function that
	// Goa writes for one service.
	protobufPackageCatalog struct {
		packageName       string
		plan              *protobufServicePlan
		messages          []*protobufMessageRecord
		messageUses       map[*expr.AttributeExpr]*protobufMessageRecord
		unions            []*protobufUnionRecord
		unionUses         map[*expr.AttributeExpr]*protobufUnionRecord
		rootSources       map[expr.UserType]protobufMessageSource
		validators        []*protobufValidationRecord
		validationUses    map[protobufValidationUse]*protobufValidationRecord
		frozen            bool
		messagesRendered  bool
		validationsFrozen bool
	}

	// protobufEndpointMessages contains the copied request, response, and error
	// values used to write one endpoint's protobuf code.
	protobufEndpointMessages struct {
		request          *expr.AttributeExpr
		streamingRequest *expr.AttributeExpr
		requestEnvelope  *expr.AttributeExpr
		response         *expr.AttributeExpr
		errors           map[string]*expr.AttributeExpr
	}

	// protobufMessageRecord stores one protobuf message and every place that
	// uses it.
	protobufMessageRecord struct {
		identity    protobufMessageIdentity
		uses        []*expr.AttributeExpr
		protoName   string
		plannedName string
		declaration *codegen.NameDeclaration
		name        string
		goRef       string
		data        *service.UserTypeData
	}

	// This record stores the source declaration, requested name, explicit-name
	// flag, and type fields used to decide whether two values can share one
	// protobuf message.
	protobufMessageIdentity struct {
		source        protobufMessageSource
		preferredName string
		explicitName  bool
		userType      expr.UserType
		attribute     *expr.AttributeExpr
	}

	// protobufUnionRecord stores one oneof inside its protobuf message.
	protobufUnionRecord struct {
		owner     *protobufMessageRecord
		attribute *expr.AttributeExpr
		fieldName string
		uses      []*expr.AttributeExpr
		name      string
	}

	// protobufMessageSource points to either an authored declaration or an
	// endpoint message created by the generator.
	protobufMessageSource struct {
		origin    expr.UserType
		synthetic protobufSyntheticMessage
	}

	// protobufSyntheticMessage stores a message that Goa creates for an endpoint.
	protobufSyntheticMessage struct {
		endpoint *expr.GRPCEndpointExpr
		error    *expr.GRPCErrorExpr
		role     protobufSyntheticRole
	}

	// protobufSyntheticRole says which endpoint value a Goa-created message holds.
	protobufSyntheticRole uint8

	// protobufValidationRecord stores one validation function written to a
	// client or server package.
	protobufValidationRecord struct {
		message     *protobufMessageRecord
		declaration *codegen.NameDeclaration
		attribute   *expr.AttributeExpr
		source      protobufValidationSource
		side        validateKind
		targetName  string
		contextName string
		uses        []*expr.AttributeExpr
		data        *ValidationData
	}

	// protobufValidationSource records the endpoint value and nested field that
	// first needs one validation function.
	protobufValidationSource struct {
		api     string
		service string
		method  string
		error   string
		path    string
		role    protobufValidationRole
	}

	// protobufValidationRole says which endpoint value is checked.
	protobufValidationRole uint8

	// protobufValidationUse records which function checks one copied value in a
	// generated client or server package.
	protobufValidationUse struct {
		attribute *expr.AttributeExpr
		side      validateKind
	}

	// protobufAttributePair stops a comparison when recursive values lead back
	// to the same pair.
	protobufAttributePair struct {
		left  *expr.AttributeExpr
		right *expr.AttributeExpr
	}

	// This value supplies the protobuf message names and validation function
	// names used while Goa writes validation code.
	protobufValidationScope struct {
		*protoBufScope
		catalog *protobufPackageCatalog
		side    validateKind
		message *protobufMessageRecord
		parent  expr.UserType
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

var protobufExactNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

const (
	protobufRequestValidation protobufValidationRole = iota + 1
	protobufResponseValidation
	protobufErrorValidation
	protobufStreamingRequestValidation
)

// newProtobufPackageCatalog creates the protobuf messages for one service and
// the functions that validate those messages in its client and server.
func newProtobufPackageCatalog(packageName string) *protobufPackageCatalog {
	return &protobufPackageCatalog{
		packageName:    packageName,
		messageUses:    make(map[*expr.AttributeExpr]*protobufMessageRecord),
		unionUses:      make(map[*expr.AttributeExpr]*protobufUnionRecord),
		rootSources:    make(map[expr.UserType]protobufMessageSource),
		validationUses: make(map[protobufValidationUse]*protobufValidationRecord),
	}
}

// bindRootSource records which service or endpoint value a generated top-level
// message carries. Copies of that message use the same information.
func (c *protobufPackageCatalog) bindRootSource(attribute *expr.AttributeExpr, source protobufMessageSource) {
	if attribute.Type == expr.Empty {
		return
	}
	userType, ok := attribute.Type.(expr.UserType)
	if !ok || (source.origin == nil && source.synthetic.role == 0) {
		return
	}
	c.rootSources[userType.Origin()] = source
}

// collectMessage records every protobuf message reachable from an endpoint's
// request, response, or error attribute. source identifies that top-level
// attribute. Nested user types keep their own declarations.
func (c *protobufPackageCatalog) collectMessage(attribute *expr.AttributeExpr, source protobufMessageSource) error {
	if c.frozen {
		panic("cannot collect a protobuf message after the package catalog is frozen")
	}
	return c.collectMessageRecursive(attribute, source, true, nil, "")
}

// freezeMessages chooses the final protobuf and Go names for every message. It
// then connects each copied value to its message and prepares template data.
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

// freezeMessageNames chooses the final package-level name for every message
// without creating its .proto definition. Conversion code uses this when it
// needs the Go type name but another service writes the message.
func (c *protobufPackageCatalog) freezeMessageNames() {
	if c.frozen {
		return
	}
	c.frozen = true
	if c.plan == nil {
		panic("protobuf message names were not planned")
	}
	for _, record := range c.messages {
		record.name = record.declaration.Name()
		record.goRef = "*" + c.packageName + "." + record.name
	}
	for _, record := range c.unions {
		record.name = c.plan.oneofInterfaceName(record)
	}
}

// collectValidation records the validation function needed for attribute and
// every nested protobuf message it can call.
func (c *protobufPackageCatalog) collectValidation(attribute *expr.AttributeExpr, side validateKind, source protobufValidationSource, targetName, contextName string) {
	if c.validationsFrozen {
		panic("cannot collect a protobuf validator after validators freeze")
	}
	c.collectValidationRecursive(attribute, side, source, targetName, contextName, make(map[*protobufValidationRecord]struct{}))
}

// freezeValidations builds each validation function with the name already
// chosen for its generated client or server package.
func (c *protobufPackageCatalog) freezeValidations(sd *ServiceData) []*ValidationData {
	if c.validationsFrozen {
		return c.validationData()
	}
	c.validationsFrozen = true
	for _, record := range c.validators {
		if record.declaration == nil {
			panic(fmt.Sprintf("protobuf validator for %q has no generated declaration", record.message.plannedName))
		}
		validationAttribute := expr.DupAtt(record.attribute)
		c.plan.bindAttributeCopy(record.attribute, validationAttribute)
		c.bindCopiedValidationUses(record.attribute, validationAttribute, record.side)
		removeMeta(validationAttribute)
		userType := validationAttribute.Type.(expr.UserType)
		context := protoBufTypeContext(c.packageName, sd, false)
		context.Scope = &protobufValidationScope{
			protoBufScope: context.Scope.(*protoBufScope),
			catalog:       c,
			side:          record.side,
			message:       record.message,
			parent:        userType,
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
			Declaration: record.declaration,
			Name:        record.declaration.Name(),
			Def:         definition,
			ArgName:     record.targetName,
			SrcName:     record.message.name,
			SrcRef:      record.message.goRef,
			Kind:        record.side,
		}
	}
	return c.validationData()
}

// message returns the completed message used for attribute.
func (c *protobufPackageCatalog) message(attribute *expr.AttributeExpr) *protobufMessageRecord {
	if !c.frozen {
		panic("cannot resolve a protobuf message before the package catalog freezes")
	}
	return c.messageRecord(attribute)
}

// messageRecord returns the protobuf message collected for attribute. It may
// be called before package names are fixed.
func (c *protobufPackageCatalog) messageRecord(attribute *expr.AttributeExpr) *protobufMessageRecord {
	return c.messageUses[attribute]
}

// unionName returns the chosen Go interface name for one oneof declaration.
func (c *protobufPackageCatalog) unionName(attribute *expr.AttributeExpr) string {
	if !c.frozen {
		panic("cannot resolve a protobuf oneof before the package catalog freezes")
	}
	record := c.unionUses[attribute]
	if record == nil {
		panic(fmt.Sprintf("protobuf oneof %q has no frozen declaration", attribute.Type.Name()))
	}
	return record.name
}

// validation returns the completed validation function used for attribute in
// the client or server package.
func (c *protobufPackageCatalog) validation(attribute *expr.AttributeExpr, side validateKind) *ValidationData {
	if !c.validationsFrozen {
		panic("cannot resolve a protobuf validator before validators freeze")
	}
	record := c.validationUses[protobufValidationUse{attribute: attribute, side: side}]
	if record == nil {
		return nil
	}
	return record.data
}

// Ref returns the current message name when union validation asks for its
// parent type. Other values use the protobuf package records.
func (s *protobufValidationScope) Ref(attribute *expr.AttributeExpr, pkg string) string {
	if attribute.Type == s.parent {
		name := s.message.name
		if pkg != "" {
			name = pkg + "." + name
		}
		return "*" + name
	}
	return s.protoBufScope.Ref(attribute, pkg)
}

// ValidatorCall returns a call to the validation function written to the
// current client or server package.
func (s *protobufValidationScope) ValidatorCall(attribute *expr.AttributeExpr, _, target, _ string) string {
	validator := s.catalog.validationRecord(attribute, s.side)
	if validator == nil {
		panic("protobuf validator was not retained")
	}
	return fmt.Sprintf("%s(%s)", validator.declaration.Name(), target)
}

// collectMessageRecursive records messages and oneofs. Existing message
// records stop recursive user types.
func (c *protobufPackageCatalog) collectMessageRecursive(attribute *expr.AttributeExpr, source protobufMessageSource, root bool, owner *protobufMessageRecord, fieldName string) error {
	if attribute == nil {
		return nil
	}
	if expr.IsPrimitive(attribute.Type) {
		return nil
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		origin := actual.Origin()
		identitySource := protobufMessageSource{origin: origin}
		if rootSource, ok := c.rootSources[origin]; ok {
			identitySource = rootSource
		}
		if !root && len(actual.Attribute().Meta[wrappedAttrMeta]) > 0 {
			identitySource = protobufMessageSource{synthetic: protobufSyntheticMessage{
				role: protobufWrapperMessage,
			}}
		}
		if root && (source.origin != nil || source.synthetic.role != 0) {
			identitySource = source
			c.rootSources[origin] = source
		}
		identity, err := protobufMessageIdentityFor(attribute, identitySource)
		if err != nil {
			return err
		}
		record := c.findMessage(identity)
		if record != nil {
			record.uses = append(record.uses, attribute)
			c.messageUses[attribute] = record
			c.bindCopiedMessageUses(record.uses[0], attribute)
			return nil
		}
		record = &protobufMessageRecord{identity: identity, uses: []*expr.AttributeExpr{attribute}}
		c.messages = append(c.messages, record)
		c.messageUses[attribute] = record
		return c.collectMessageRecursive(userTypeAttribute(actual), protobufMessageSource{}, false, record, "")
	case *expr.Object:
		for _, named := range *actual {
			if err := c.collectMessageRecursive(named.Attribute, protobufMessageSource{}, false, owner, named.Name); err != nil {
				return err
			}
		}
	case *expr.Array:
		return c.collectMessageRecursive(actual.ElemType, protobufMessageSource{}, false, owner, fieldName+"Elem")
	case *expr.Map:
		if err := c.collectMessageRecursive(actual.KeyType, protobufMessageSource{}, false, owner, fieldName+"Key"); err != nil {
			return err
		}
		return c.collectMessageRecursive(actual.ElemType, protobufMessageSource{}, false, owner, fieldName+"Elem")
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
			if err := c.collectMessageRecursive(named.Attribute, protobufMessageSource{}, false, owner, fieldName+named.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// bindCopiedMessageUses records the message, choice, field, and wrapper names
// for every part of a copied protobuf value.
func (c *protobufPackageCatalog) bindCopiedMessageUses(original, copy *expr.AttributeExpr) {
	walkProtobufCopy(original, copy, func(original, copy *expr.AttributeExpr) {
		if record := c.messageUses[original]; record != nil {
			if existing := c.messageUses[copy]; existing != nil && existing != record {
				panic("protobuf copy is connected to two message declarations")
			}
			if c.messageUses[copy] == nil {
				c.messageUses[copy] = record
				record.uses = append(record.uses, copy)
			}
		}
		if record := c.unionUses[original]; record != nil {
			if existing := c.unionUses[copy]; existing != nil && existing != record {
				panic("protobuf copy is connected to two oneof declarations")
			}
			if c.unionUses[copy] == nil {
				c.unionUses[copy] = record
				record.uses = append(record.uses, copy)
			}
		}
	})
}

// bindCopiedValidationUses records the validation function for each matching
// part of a copied protobuf value.
func (c *protobufPackageCatalog) bindCopiedValidationUses(original, copy *expr.AttributeExpr, side validateKind) {
	walkProtobufCopy(original, copy, func(original, copy *expr.AttributeExpr) {
		record := c.validationUses[protobufValidationUse{attribute: original, side: side}]
		if record != nil {
			c.bindValidationUse(copy, side, record)
		}
	})
}

// findUnion returns the oneof with the same parent message, field name, and
// branch types.
func (c *protobufPackageCatalog) findUnion(owner *protobufMessageRecord, fieldName string, attribute *expr.AttributeExpr) *protobufUnionRecord {
	for _, record := range c.unions {
		if record.owner == owner && record.fieldName == fieldName &&
			sameProtobufWireAttribute(record.attribute, attribute, make(map[protobufAttributePair]struct{})) {
			return record
		}
	}
	return nil
}

// collectValidationRecursive records one function per message, set of rules,
// and client or server package, then visits the nested messages it may call.
func (c *protobufPackageCatalog) collectValidationRecursive(attribute *expr.AttributeExpr, side validateKind, source protobufValidationSource, targetName, contextName string, seen map[*protobufValidationRecord]struct{}) {
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if expr.IsPrimitive(actual) {
			return
		}
		policy := codegen.GoLayoutPolicy{IgnoreRequired: true}
		if !codegen.NeedsValidation(userTypeAttribute(actual), policy) {
			return
		}
		message := c.messageRecord(attribute)
		if message == nil {
			panic(fmt.Sprintf("no protobuf declaration collected for validation type %q", actual.Name()))
		}
		record := c.findValidation(message, attribute, side)
		if record == nil {
			record = &protobufValidationRecord{
				message:     message,
				attribute:   attribute,
				source:      source,
				side:        side,
				targetName:  targetName,
				contextName: contextName,
			}
			c.validators = append(c.validators, record)
		} else if source.compare(record.source) < 0 {
			record.source = source
			record.targetName = targetName
			record.contextName = contextName
		}
		c.bindValidationUse(attribute, side, record)
		if _, ok := seen[record]; ok {
			return
		}
		seen[record] = struct{}{}
		c.collectValidationRecursive(userTypeAttribute(actual), side, source, targetName, contextName, seen)
	case *expr.Object:
		for _, named := range *actual {
			c.collectValidationRecursive(named.Attribute, side, source.child(named.Name), codegen.Goify(named.Name, false), named.Name, seen)
		}
	case *expr.Array:
		c.collectValidationRecursive(actual.ElemType, side, source.child("element"), "elem", "elem", seen)
	case *expr.Map:
		c.collectValidationRecursive(actual.KeyType, side, source.child("key"), "key", "key", seen)
		c.collectValidationRecursive(actual.ElemType, side, source.child("value"), "val", "val", seen)
	case *expr.Union:
		for _, named := range actual.Values {
			c.collectValidationRecursive(named.Attribute, side, source.child(named.Name), codegen.Goify(named.Name, false), named.Name, seen)
		}
	}
}

// findMessage returns the message written for the same source type, requested
// name, and protobuf fields.
func (c *protobufPackageCatalog) findMessage(identity protobufMessageIdentity) *protobufMessageRecord {
	for _, record := range c.messages {
		if sameProtobufMessageIdentity(record.identity, identity) {
			return record
		}
	}
	return nil
}

// findValidation returns the existing function that checks the same message
// rules in the same client or server package.
func (c *protobufPackageCatalog) findValidation(declaration *protobufMessageRecord, attribute *expr.AttributeExpr, side validateKind) *protobufValidationRecord {
	for _, record := range c.validators {
		if record.message == declaration && record.side == side &&
			sameProtobufValidationAttribute(record.attribute, attribute, make(map[protobufAttributePair]struct{})) {
			return record
		}
	}
	return nil
}

// validationRecord returns the validation function called for one copied
// message value.
func (c *protobufPackageCatalog) validationRecord(attribute *expr.AttributeExpr, side validateKind) *protobufValidationRecord {
	return c.validationUses[protobufValidationUse{attribute: attribute, side: side}]
}

// bindValidationUse records the function that checks one protobuf value in a
// generated client or server package.
func (c *protobufPackageCatalog) bindValidationUse(attribute *expr.AttributeExpr, side validateKind, record *protobufValidationRecord) {
	key := protobufValidationUse{attribute: attribute, side: side}
	if existing := c.validationUses[key]; existing != nil && existing != record {
		panic("protobuf value is connected to two validation functions")
	}
	if c.validationUses[key] == nil {
		c.validationUses[key] = record
		record.uses = append(record.uses, attribute)
	}
}

// walkProtobufCopy visits matching parts of an original protobuf value and its
// copy. Recursive user types are visited once.
func walkProtobufCopy(original, copy *expr.AttributeExpr, visit func(*expr.AttributeExpr, *expr.AttributeExpr)) {
	seen := make(map[protobufAttributePair]struct{})
	var walk func(*expr.AttributeExpr, *expr.AttributeExpr)
	walk = func(original, copy *expr.AttributeExpr) {
		pair := protobufAttributePair{left: original, right: copy}
		if _, ok := seen[pair]; ok {
			return
		}
		seen[pair] = struct{}{}
		visit(original, copy)
		switch originalType := original.Type.(type) {
		case expr.UserType:
			copyType := copy.Type.(expr.UserType)
			walk(userTypeAttribute(originalType), userTypeAttribute(copyType))
		case *expr.Object:
			copyType := copy.Type.(*expr.Object)
			for index, field := range *originalType {
				walk(field.Attribute, (*copyType)[index].Attribute)
			}
		case *expr.Array:
			walk(originalType.ElemType, copy.Type.(*expr.Array).ElemType)
		case *expr.Map:
			copyType := copy.Type.(*expr.Map)
			walk(originalType.KeyType, copyType.KeyType)
			walk(originalType.ElemType, copyType.ElemType)
		case *expr.Union:
			copyType := copy.Type.(*expr.Union)
			for index, branch := range originalType.Values {
				walk(branch.Attribute, copyType.Values[index].Attribute)
			}
		}
	}
	walk(original, copy)
}

// messageData returns only messages whose names and template data are complete.
func (c *protobufPackageCatalog) messageData() []*service.UserTypeData {
	data := make([]*service.UserTypeData, 0, len(c.messages))
	for _, record := range c.messages {
		if record.data != nil {
			data = append(data, record.data)
		}
	}
	return data
}

// protoMessageData returns message records with the names written to the
// protobuf source file.
func (c *protobufPackageCatalog) protoMessageData() []*service.UserTypeData {
	data := make([]*service.UserTypeData, 0, len(c.messages))
	for _, record := range c.messages {
		if record.data == nil {
			continue
		}
		message := *record.data
		message.Name = record.protoName
		message.VarName = record.protoName
		data = append(data, &message)
	}
	return data
}

// validationData returns only validation functions that contain checks.
func (c *protobufPackageCatalog) validationData() []*ValidationData {
	data := make([]*ValidationData, 0, len(c.validators))
	for _, record := range c.validators {
		if record.data != nil {
			data = append(data, record.data)
		}
	}
	return data
}

// child returns the same endpoint value at one nested field or collection
// part.
func (s protobufValidationSource) child(name string) protobufValidationSource {
	if s.path == "" {
		s.path = name
	} else {
		s.path += "." + name
	}
	return s
}

// compare orders endpoint values and their nested fields the same way even when
// the design lists them in a different order.
func (s protobufValidationSource) compare(other protobufValidationSource) int {
	if result := strings.Compare(s.api, other.api); result != 0 {
		return result
	}
	if result := strings.Compare(s.service, other.service); result != 0 {
		return result
	}
	if result := strings.Compare(s.method, other.method); result != 0 {
		return result
	}
	if result := strings.Compare(s.error, other.error); result != 0 {
		return result
	}
	if s.role < other.role {
		return -1
	}
	if s.role > other.role {
		return 1
	}
	return strings.Compare(s.path, other.path)
}

// protobufMessageIdentityFor records the source type, requested name, and
// protobuf fields that decide whether two uses share one message.
func protobufMessageIdentityFor(attribute *expr.AttributeExpr, source protobufMessageSource) (protobufMessageIdentity, error) {
	userType := attribute.Type.(expr.UserType)
	if source.origin == nil && source.synthetic.role == 0 {
		source.origin = userType.Origin()
	}
	preferred := codegen.ProtobufName(userType.Name())
	explicit := false
	names := attribute.Meta["struct:name:proto"]
	if len(names) == 0 {
		names = userType.Attribute().Meta["struct:name:proto"]
	}
	if len(names) > 0 {
		if !protobufExactNamePattern.MatchString(names[0]) {
			return protobufMessageIdentity{}, fmt.Errorf("protobuf message name %q from struct:name:proto is not a valid protobuf identifier", names[0])
		}
		preferred = names[0]
		explicit = true
	}
	return protobufMessageIdentity{
		source:        source,
		preferredName: preferred,
		explicitName:  explicit,
		userType:      userType,
		attribute:     userTypeAttribute(userType),
	}, nil
}

// sameProtobufMessageIdentity reports whether two requests have the same source
// type, requested name, and protobuf fields. It does not compare generated Go
// names.
func sameProtobufMessageIdentity(left, right protobufMessageIdentity) bool {
	if left.source != right.source || left.preferredName != right.preferredName || left.explicitName != right.explicitName {
		return false
	}
	return sameProtobufWireAttribute(left.attribute, right.attribute, make(map[protobufAttributePair]struct{}))
}

// sameProtobufWireAttribute reports whether two values produce the same fields
// and rules in a .proto message.
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

// sameProtobufWireType reports whether two types produce the same protobuf type,
// including ordered fields, choices, arrays, maps, and nested messages.
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

// sameProtobufValidationAttribute compares the source types and validation
// rules without comparing the protobuf message name.
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
