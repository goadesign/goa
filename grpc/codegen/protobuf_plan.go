// This file asks the supported protobuf tools for every Go name that Goa must
// reference, then stores those names before any generated file is built.
package codegen

import (
	"cmp"
	"fmt"
	"sort"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// protobufServicePlan stores one service's copied messages and the Go names
	// that protoc and its Go plugins produce for them.
	protobufServicePlan struct {
		expression   *expr.GRPCServiceExpr
		catalog      *protobufPackageCatalog
		messages     []*protobufEndpointMessages
		protoPackage string
		serviceName  string
		fileIndex    int
		order        protobufServiceOrder
		methods      map[*expr.GRPCEndpointExpr]string
		names        map[protocNameKey]*codegen.NameDeclaration
		localNames   map[protocNameKey]string
		fields       map[*expr.AttributeExpr]protocNameKey
		sourceFields map[*expr.AttributeExpr]string
		sourceOneofs map[*expr.AttributeExpr]string
		wrappers     map[*expr.AttributeExpr]protocNameKey
		oneofs       map[*expr.AttributeExpr]protocNameKey
	}

	// protobufNameGroup holds one name written to a .proto file and every Go name
	// generated from it. If a Go name is already used, Goa adds the same number
	// to the .proto name and asks the tools for the complete set again.
	protobufNameGroup struct {
		preferred string
		name      string
		suffix    int
		message   *protobufMessageRecord
		method    *expr.GRPCEndpointExpr
		service   bool
	}

	// protobufServiceOrder holds the two names used to place services in a
	// stable order.
	protobufServiceOrder struct {
		service string
		api     string
	}
)

// planProtobufServices records every protobuf declaration in the generated Go
// package that will contain it.
func planProtobufServices(generation *codegen.Generation, roots []*Plan) error {
	groups := make(map[string][]*protobufServicePlan)
	for _, rootPlan := range roots {
		for _, service := range rootPlan.expressions {
			pathName := rootPlan.packages[service].pathName
			packagePath := generation.GenPkg() + "/grpc/" + pathName + "/" + pbPkgName
			catalog := newProtobufPackageCatalog("")
			messages, err := collectProtobufPackage(service, catalog)
			if err != nil {
				return fmt.Errorf("service %q %w", service.Name(), err)
			}
			plan := &protobufServicePlan{
				expression:   service,
				catalog:      catalog,
				messages:     messages,
				protoPackage: pkgName(service, pathName),
				order: protobufServiceOrder{
					service: service.Name(),
					api:     rootPlan.root.API.Name,
				},
				methods:      make(map[*expr.GRPCEndpointExpr]string, len(service.GRPCEndpoints)),
				names:        make(map[protocNameKey]*codegen.NameDeclaration),
				localNames:   make(map[protocNameKey]string),
				fields:       make(map[*expr.AttributeExpr]protocNameKey),
				sourceFields: make(map[*expr.AttributeExpr]string),
				sourceOneofs: make(map[*expr.AttributeExpr]string),
				wrappers:     make(map[*expr.AttributeExpr]protocNameKey),
				oneofs:       make(map[*expr.AttributeExpr]protocNameKey),
			}
			catalog.plan = plan
			rootPlan.protobuf[service] = plan
			groups[packagePath] = append(groups[packagePath], plan)
		}
	}
	for packagePath, group := range groups {
		pkg, err := generation.ClaimPackage(packagePath)
		if err != nil {
			return err
		}
		sort.Slice(group, func(i, j int) bool {
			if group[i].order.service != group[j].order.service {
				return group[i].order.service < group[j].order.service
			}
			return group[i].order.api < group[j].order.api
		})
		for index := 1; index < len(group); index++ {
			if group[index-1].order == group[index].order {
				return fmt.Errorf(
					"generated protobuf package %q has two services named %q in API %q",
					packagePath,
					group[index].order.service,
					group[index].order.api,
				)
			}
		}
		for _, plan := range group[1:] {
			if plan.protoPackage != group[0].protoPackage {
				return fmt.Errorf("generated package %q cannot contain protobuf packages %q and %q", packagePath, group[0].protoPackage, plan.protoPackage)
			}
		}
		used := make(map[string]struct{})
		for index, plan := range group {
			plan.fileIndex = index + 1
			if err := plan.chooseNames(pkg, used); err != nil {
				return fmt.Errorf("plan protobuf names for service %q: %w", plan.expression.Name(), err)
			}
		}
	}
	return nil
}

// name returns one Go name produced by the supported protobuf tools.
func (p *protobufServicePlan) name(descriptor string, role protocNameRole) string {
	key := protocNameKey{descriptor: descriptor, role: role}
	if declaration := p.names[key]; declaration != nil {
		return declaration.Name()
	}
	name, ok := p.localNames[key]
	if !ok {
		panic(fmt.Sprintf("protobuf Go name %s for %q was not planned", role, descriptor))
	}
	return name
}

// fieldName returns the Go field name produced for one copied protobuf field.
func (p *protobufServicePlan) fieldName(attribute *expr.AttributeExpr) (string, bool) {
	key, ok := p.fields[attribute]
	if !ok {
		return "", false
	}
	return p.name(key.descriptor, key.role), true
}

// sourceFieldName returns the field name written to the protobuf file.
func (p *protobufServicePlan) sourceFieldName(attribute *expr.AttributeExpr) string {
	name, ok := p.sourceFields[attribute]
	if !ok {
		panic("protobuf source field was not planned")
	}
	return name
}

// sourceOneofName returns the oneof name written to the protobuf file.
func (p *protobufServicePlan) sourceOneofName(attribute *expr.AttributeExpr) string {
	name, ok := p.sourceOneofs[attribute]
	if !ok {
		panic("protobuf source oneof was not planned")
	}
	return name
}

// wrapperName returns the Go wrapper type for one branch in one parent
// message.
func (p *protobufServicePlan) wrapperName(attribute *expr.AttributeExpr) (string, bool) {
	key, ok := p.wrappers[attribute]
	if !ok {
		return "", false
	}
	return p.name(key.descriptor, key.role), true
}

// oneofInterfaceName returns the Go interface produced for one oneof.
func (p *protobufServicePlan) oneofInterfaceName(record *protobufUnionRecord) string {
	key, ok := p.oneofs[record.attribute]
	if !ok {
		panic("protobuf oneof interface name was not planned")
	}
	return p.name(key.descriptor, key.role)
}

// bindAttributeCopy records every message, choice, field, and wrapper name for
// the matching parts of a copied protobuf value.
func (p *protobufServicePlan) bindAttributeCopy(original, copy *expr.AttributeExpr) {
	p.catalog.bindCopiedMessageUses(original, copy)
	walkProtobufCopy(original, copy, func(original, copy *expr.AttributeExpr) {
		if key, ok := p.fields[original]; ok {
			p.fields[copy] = key
		}
		if name, ok := p.sourceFields[original]; ok {
			p.sourceFields[copy] = name
		}
		if name, ok := p.sourceOneofs[original]; ok {
			p.sourceOneofs[copy] = name
		}
		if key, ok := p.wrappers[original]; ok {
			p.wrappers[copy] = key
		}
		if key, ok := p.oneofs[original]; ok {
			p.oneofs[copy] = key
		}
	})
}

// chooseNames tries numbered protobuf names until every generated Go name is
// unique in the package.
func (p *protobufServicePlan) chooseNames(pkg *codegen.GeneratedPackage, used map[string]struct{}) error {
	groups, err := p.nameGroups()
	if err != nil {
		return err
	}
	p.assignInitialNames(groups)
	rules, err := newProtocNameRules(protocNameVersionGo1_36GRPC1_6)
	if err != nil {
		return err
	}
	for attempts := 0; attempts < 1000; attempts++ {
		descriptor, owners, err := p.namingDescriptor(groups)
		if err != nil {
			return err
		}
		generated, err := rules.file(descriptor)
		if err != nil {
			return err
		}
		colliding := collidingProtobufGroup(generated, owners, groups, used)
		if colliding == nil {
			if err := p.declareNames(pkg, generated); err != nil {
				return err
			}
			for key, name := range generated.values {
				if _, packageName := protocPackageNameKind(key.role); packageName {
					used[name] = struct{}{}
				}
			}
			return nil
		}
		colliding.name = nextAvailableProtobufName(colliding, groups)
	}
	return fmt.Errorf("could not choose unique protobuf Go names")
}

// nameGroups puts the service and methods before messages. This keeps the
// requested service and method names when a message would generate the same Go
// name.
func (p *protobufServicePlan) nameGroups() ([]*protobufNameGroup, error) {
	service := &protobufNameGroup{
		preferred: codegen.ProtobufName(p.expression.Name()),
		service:   true,
	}
	methods := make([]*protobufNameGroup, 0, len(p.expression.GRPCEndpoints))
	for _, endpoint := range p.expression.GRPCEndpoints {
		methods = append(methods, &protobufNameGroup{
			preferred: codegen.ProtobufName(endpoint.Name()),
			method:    endpoint,
		})
	}
	sort.Slice(methods, func(i, j int) bool {
		return compareProtobufEndpointSource(methods[i].method, methods[j].method) < 0
	})
	messages := make([]*protobufNameGroup, 0, len(p.catalog.messages))
	for _, message := range p.catalog.messages {
		messages = append(messages, &protobufNameGroup{
			preferred: message.identity.preferredName,
			message:   message,
		})
	}
	if err := sortProtobufMessageGroups(messages); err != nil {
		return nil, err
	}
	groups := []*protobufNameGroup{service}
	groups = append(groups, methods...)
	return append(groups, messages...), nil
}

// assignInitialNames makes message and service names unique in the file. Method
// names use a separate list because protobuf allows the same method name in a
// different service.
func (p *protobufServicePlan) assignInitialNames(groups []*protobufNameGroup) {
	packageNames := make(map[string]struct{})
	methodNames := make(map[string]struct{})
	for _, group := range groups {
		used := packageNames
		if group.method != nil {
			used = methodNames
		}
		assignProtobufGroupName(group, used)
	}
}

// namingDescriptor builds a small protobuf file containing only declarations
// that can change generated Go names. A field's type cannot change its Go name,
// so these temporary fields all use string.
func (p *protobufServicePlan) namingDescriptor(groups []*protobufNameGroup) (*descriptorpb.FileDescriptorProto, map[protocNameKey]*protobufNameGroup, error) {
	owners := make(map[protocNameKey]*protobufNameGroup)
	file := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("goa_names.proto"),
		Package: proto.String(p.protoPackage),
		Syntax:  proto.String(ProtoVersion),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("/" + p.protoPackage + "pb")},
	}
	var serviceGroup *protobufNameGroup
	for _, group := range groups {
		switch {
		case group.service:
			serviceGroup = group
			p.serviceName = group.name
		case group.method != nil:
			p.methods[group.method] = group.name
		case group.message != nil:
			group.message.protoName = group.name
			message, keys := p.messageDescriptor(group.message)
			for _, use := range group.message.uses {
				useType := use.Type.(expr.UserType)
				p.bindAttributeCopy(group.message.identity.attribute, userTypeAttribute(useType))
			}
			file.MessageType = append(file.MessageType, message)
			for _, key := range keys {
				owners[key] = group
			}
		}
	}
	service, keys, err := p.serviceDescriptor(serviceGroup, groups)
	if err != nil {
		return nil, nil, err
	}
	file.Service = []*descriptorpb.ServiceDescriptorProto{service}
	for _, key := range keys {
		owner := serviceGroup
		for _, group := range groups {
			if group.method != nil && key.descriptor == p.serviceFullName()+"."+group.name {
				owner = group
				break
			}
		}
		owners[key] = owner
	}
	return file, owners, nil
}

// messageDescriptor records one message's fields and oneofs for protogen.
func (p *protobufServicePlan) messageDescriptor(record *protobufMessageRecord) (*descriptorpb.DescriptorProto, []protocNameKey) {
	message := &descriptorpb.DescriptorProto{Name: proto.String(record.protoName)}
	fullName := p.protoPackage + "." + record.protoName
	keys := []protocNameKey{{descriptor: fullName, role: protocMessageName}}
	usedFieldNames := make(map[string]struct{})
	fieldNames := make(map[*expr.AttributeExpr]string)
	fieldNumber := int32(1)
	attribute := record.identity.attribute
	if userType, ok := attribute.Type.(expr.UserType); ok {
		attribute = userType.Attribute()
	}
	object := expr.AsObject(attribute.Type)
	if object == nil {
		if union, ok := attribute.Type.(*expr.Union); ok {
			for _, branch := range union.Values {
				fieldNames[branch.Attribute] = uniqueProtobufSourceName(protobufSourceFieldName(branch.Name), usedFieldNames)
			}
			oneofName := uniqueProtobufOneofSourceName(union.Name(), usedFieldNames)
			p.addOneofDescriptor(message, fullName, oneofName, attribute, union, fieldNames, &fieldNumber, &keys)
		}
		return message, keys
	}
	for _, named := range *object {
		if union, ok := named.Attribute.Type.(*expr.Union); ok {
			for _, branch := range union.Values {
				fieldNames[branch.Attribute] = uniqueProtobufSourceName(protobufSourceFieldName(branch.Name), usedFieldNames)
			}
			continue
		}
		fieldNames[named.Attribute] = uniqueProtobufSourceName(protobufSourceFieldName(named.Name), usedFieldNames)
	}
	oneofNames := make(map[*expr.AttributeExpr]string)
	for _, named := range *object {
		if _, ok := named.Attribute.Type.(*expr.Union); ok {
			oneofNames[named.Attribute] = uniqueProtobufOneofSourceName(named.Name, usedFieldNames)
		}
	}
	for _, named := range *object {
		if union, ok := named.Attribute.Type.(*expr.Union); ok {
			p.addOneofDescriptor(message, fullName, oneofNames[named.Attribute], named.Attribute, union, fieldNames, &fieldNumber, &keys)
			continue
		}
		fieldName := fieldNames[named.Attribute]
		message.Field = append(message.Field, namingField(fieldName, fieldNumber, nil))
		fieldNumber++
		key := protocNameKey{descriptor: fullName + "." + fieldName, role: protocFieldName}
		p.fields[named.Attribute] = key
		p.sourceFields[named.Attribute] = fieldName
		keys = append(keys, key)
	}
	return message, keys
}

// addOneofDescriptor records one oneof and each branch field after every source
// name in the message has been selected.
func (p *protobufServicePlan) addOneofDescriptor(message *descriptorpb.DescriptorProto, messageName, oneofName string, attribute *expr.AttributeExpr, union *expr.Union, fieldNames map[*expr.AttributeExpr]string, fieldNumber *int32, keys *[]protocNameKey) {
	index := int32(len(message.OneofDecl))
	message.OneofDecl = append(message.OneofDecl, &descriptorpb.OneofDescriptorProto{Name: proto.String(oneofName)})
	fieldKey := protocNameKey{descriptor: messageName + "." + oneofName, role: protocOneofFieldName}
	interfaceKey := protocNameKey{descriptor: messageName + "." + oneofName, role: protocOneofInterfaceName}
	p.fields[attribute] = fieldKey
	p.sourceOneofs[attribute] = oneofName
	p.oneofs[attribute] = interfaceKey
	*keys = append(*keys, fieldKey, interfaceKey)
	for _, branch := range union.Values {
		name := fieldNames[branch.Attribute]
		message.Field = append(message.Field, namingField(name, *fieldNumber, &index))
		*fieldNumber++
		fieldKey := protocNameKey{descriptor: messageName + "." + name, role: protocFieldName}
		wrapperKey := protocNameKey{descriptor: messageName + "." + name, role: protocOneofWrapperName}
		p.fields[branch.Attribute] = fieldKey
		p.sourceFields[branch.Attribute] = name
		p.wrappers[branch.Attribute] = wrapperKey
		*keys = append(*keys, fieldKey, wrapperKey)
	}
}

// serviceDescriptor records the service methods and their stream directions.
func (p *protobufServicePlan) serviceDescriptor(serviceGroup *protobufNameGroup, groups []*protobufNameGroup) (*descriptorpb.ServiceDescriptorProto, []protocNameKey, error) {
	service := &descriptorpb.ServiceDescriptorProto{Name: proto.String(serviceGroup.name)}
	serviceName := p.protoPackage + "." + serviceGroup.name
	keys := serviceNameKeys(serviceName)
	for _, group := range groups {
		if group.method == nil {
			continue
		}
		endpoint := group.method
		index := slicesIndexEndpoint(p.expression.GRPCEndpoints, endpoint)
		if index < 0 {
			return nil, nil, fmt.Errorf("method %q is not part of service %q", endpoint.Name(), p.expression.Name())
		}
		messages := p.messages[index]
		request := p.catalog.messageUses[messages.request]
		if messages.requestEnvelope != nil {
			request = p.catalog.messageUses[messages.requestEnvelope]
		} else if endpoint.MethodExpr.StreamingPayload.Type != expr.Empty {
			request = p.catalog.messageUses[messages.streamingRequest]
		}
		response := p.catalog.messageUses[messages.response]
		if request == nil || response == nil {
			return nil, nil, fmt.Errorf("method %q has no protobuf request or response message", endpoint.Name())
		}
		method := &descriptorpb.MethodDescriptorProto{
			Name:            proto.String(group.name),
			InputType:       proto.String("." + p.protoPackage + "." + request.protoName),
			OutputType:      proto.String("." + p.protoPackage + "." + response.protoName),
			ClientStreaming: proto.Bool(endpoint.MethodExpr.IsPayloadStreaming()),
			ServerStreaming: proto.Bool(endpoint.MethodExpr.IsResultStreaming()),
		}
		service.Method = append(service.Method, method)
		keys = append(keys, methodNameKeys(serviceName+"."+group.name, endpoint.MethodExpr.IsStreaming())...)
	}
	return service, keys, nil
}

// declareNames stores package-level Go names with the package that writes them.
// It stores field and method names directly because they cannot collide with
// names outside their message or service.
func (p *protobufServicePlan) declareNames(pkg *codegen.GeneratedPackage, generated *protocNames) error {
	for key, name := range generated.values {
		kind, packageName := protocPackageNameKind(key.role)
		if !packageName {
			p.localNames[key] = name
			continue
		}
		declaration := codegen.NewExactName(kind, name)
		if err := pkg.DeclareName(declaration); err != nil {
			return err
		}
		p.names[key] = declaration
	}
	for _, record := range p.catalog.messages {
		key := protocNameKey{
			descriptor: p.protoPackage + "." + record.protoName,
			role:       protocMessageName,
		}
		record.plannedName = generated.values[key]
		record.declaration = p.names[key]
		if record.declaration == nil {
			return fmt.Errorf("protobuf message %q has no generated Go declaration", record.protoName)
		}
	}
	return nil
}

// collidingProtobufGroup returns the first group that would generate a Go name
// already used in the package.
func collidingProtobufGroup(names *protocNames, owners map[protocNameKey]*protobufNameGroup, groups []*protobufNameGroup, occupied map[string]struct{}) *protobufNameGroup {
	byGroup := make(map[*protobufNameGroup][]string)
	for key, name := range names.values {
		_, packageName := protocPackageNameKind(key.role)
		if packageName {
			byGroup[owners[key]] = append(byGroup[owners[key]], name)
		}
	}
	used := make(map[string]struct{}, len(occupied))
	for name := range occupied {
		used[name] = struct{}{}
	}
	for _, group := range groups {
		for _, name := range byGroup[group] {
			if _, ok := used[name]; ok {
				return group
			}
		}
		for _, name := range byGroup[group] {
			used[name] = struct{}{}
		}
	}
	return nil
}

// protocPackageNameKind identifies names declared at package level.
func protocPackageNameKind(role protocNameRole) (codegen.PackageNameKind, bool) {
	switch role {
	case protocMessageName, protocEnumName, protocOneofInterfaceName, protocOneofWrapperName,
		protocServiceClientName, protocServiceClientStructName, protocServiceServerName,
		protocServiceUnimplementedServerName, protocServiceUnsafeServerName,
		protocMethodClientStreamName, protocMethodServerStreamName:
		return codegen.NameType, true
	case protocServiceClientConstructorName, protocServiceRegisterName, protocMethodHandlerName:
		return codegen.NameFunction, true
	case protocMethodFullName:
		return codegen.NameConstant, true
	case protocServiceDescriptorName:
		return codegen.NameVariable, true
	default:
		return 0, false
	}
}

// sortProtobufMessageGroups puts messages in the same order for every input
// order. It reports separate declarations when their source, requested name,
// and protobuf fields cannot choose which one comes first.
func sortProtobufMessageGroups(groups []*protobufNameGroup) error {
	sort.Slice(groups, func(i, j int) bool {
		return compareProtobufMessageIdentity(groups[i].message.identity, groups[j].message.identity) < 0
	})
	for index := 1; index < len(groups); index++ {
		left := groups[index-1].message.identity
		right := groups[index].message.identity
		if compareProtobufMessageIdentity(left, right) == 0 && !sameProtobufMessageIdentity(left, right) {
			return fmt.Errorf("protobuf messages named %q have the same source, name, and fields but come from separate declarations", left.preferredName)
		}
	}
	return nil
}

// compareProtobufMessageIdentity compares the source, requested name, and
// protobuf fields that decide whether two values use one message.
func compareProtobufMessageIdentity(left, right protobufMessageIdentity) int {
	if order := compareProtobufMessageSource(left.source, right.source); order != 0 {
		return order
	}
	if order := cmp.Compare(left.preferredName, right.preferredName); order != 0 {
		return order
	}
	if order := compareBool(left.explicitName, right.explicitName); order != 0 {
		return order
	}
	return compareProtobufWireAttribute(left.attribute, right.attribute, make(map[protobufAttributePair]struct{}))
}

// compareProtobufMessageSource compares the design declaration or method value
// that produced a protobuf message.
func compareProtobufMessageSource(left, right protobufMessageSource) int {
	leftAuthored := left.origin != nil
	rightAuthored := right.origin != nil
	if order := compareBool(leftAuthored, rightAuthored); order != 0 {
		return order
	}
	if leftAuthored {
		if left.origin == right.origin {
			return 0
		}
		if order := cmp.Compare(left.origin.ID(), right.origin.ID()); order != 0 {
			return order
		}
		return cmp.Compare(left.origin.Name(), right.origin.Name())
	}
	if order := cmp.Compare(left.synthetic.role, right.synthetic.role); order != 0 {
		return order
	}
	if order := compareProtobufEndpointSource(left.synthetic.endpoint, right.synthetic.endpoint); order != 0 {
		return order
	}
	return compareProtobufErrorSource(left.synthetic.error, right.synthetic.error)
}

// compareProtobufEndpointSource compares the service and method names that
// produced a generated request or response message.
func compareProtobufEndpointSource(left, right *expr.GRPCEndpointExpr) int {
	if order := compareBool(left != nil, right != nil); order != 0 {
		return order
	}
	if left == nil || left == right {
		return 0
	}
	if order := cmp.Compare(left.Service.Name(), right.Service.Name()); order != 0 {
		return order
	}
	return cmp.Compare(left.Name(), right.Name())
}

// compareProtobufErrorSource compares the error names that produced generated
// error messages.
func compareProtobufErrorSource(left, right *expr.GRPCErrorExpr) int {
	if order := compareBool(left != nil, right != nil); order != 0 {
		return order
	}
	if left == nil || left == right {
		return 0
	}
	return cmp.Compare(left.Name, right.Name)
}

// compareProtobufWireAttribute orders values by the description, protobuf
// settings, and type written to the .proto file.
func compareProtobufWireAttribute(left, right *expr.AttributeExpr, seen map[protobufAttributePair]struct{}) int {
	if left == right {
		return 0
	}
	if order := compareBool(left != nil, right != nil); order != 0 {
		return order
	}
	if order := cmp.Compare(left.Description, right.Description); order != 0 {
		return order
	}
	pair := protobufAttributePair{left: left, right: right}
	if _, ok := seen[pair]; ok {
		return 0
	}
	seen[pair] = struct{}{}
	if order := compareProtobufMeta(left.Meta, right.Meta); order != 0 {
		return order
	}
	if order := compareProtobufWireType(left.Type, right.Type, seen); order != 0 {
		return order
	}
	return 0
}

// compareProtobufWireType orders values by the protobuf type and each nested
// field in the order Goa writes them.
func compareProtobufWireType(left, right expr.DataType, seen map[protobufAttributePair]struct{}) int {
	if order := cmp.Compare(left.Kind(), right.Kind()); order != 0 {
		return order
	}
	switch left := left.(type) {
	case expr.Primitive:
		return cmp.Compare(left, right.(expr.Primitive))
	case expr.UserType:
		right := right.(expr.UserType)
		leftWrapped := len(left.Attribute().Meta[wrappedAttrMeta]) > 0
		rightWrapped := len(right.Attribute().Meta[wrappedAttrMeta]) > 0
		if order := compareBool(leftWrapped, rightWrapped); order != 0 {
			return order
		}
		if !leftWrapped && left.Origin() != right.Origin() {
			if order := cmp.Compare(left.Origin().ID(), right.Origin().ID()); order != 0 {
				return order
			}
			if order := cmp.Compare(left.Origin().Name(), right.Origin().Name()); order != 0 {
				return order
			}
		}
		return compareProtobufWireAttribute(left.Attribute(), right.Attribute(), seen)
	case *expr.Object:
		right := right.(*expr.Object)
		if order := cmp.Compare(len(*left), len(*right)); order != 0 {
			return order
		}
		for index, field := range *left {
			other := (*right)[index]
			if order := cmp.Compare(field.Name, other.Name); order != 0 {
				return order
			}
			if order := compareProtobufWireAttribute(field.Attribute, other.Attribute, seen); order != 0 {
				return order
			}
		}
		return 0
	case *expr.Array:
		return compareProtobufWireAttribute(left.ElemType, right.(*expr.Array).ElemType, seen)
	case *expr.Map:
		right := right.(*expr.Map)
		if order := compareProtobufWireAttribute(left.KeyType, right.KeyType, seen); order != 0 {
			return order
		}
		return compareProtobufWireAttribute(left.ElemType, right.ElemType, seen)
	case *expr.Union:
		right := right.(*expr.Union)
		for _, order := range []int{
			cmp.Compare(left.TypeName, right.TypeName),
			cmp.Compare(left.TypeKey, right.TypeKey),
			cmp.Compare(left.ValueKey, right.ValueKey),
			cmp.Compare(len(left.Values), len(right.Values)),
		} {
			if order != 0 {
				return order
			}
		}
		for index, branch := range left.Values {
			other := right.Values[index]
			if order := cmp.Compare(branch.Name, other.Name); order != 0 {
				return order
			}
			if order := compareProtobufWireAttribute(branch.Attribute, other.Attribute, seen); order != 0 {
				return order
			}
		}
		return 0
	default:
		panic(fmt.Sprintf("unknown protobuf wire type %T", left))
	}
}

// compareProtobufMeta compares settings that change protobuf field numbers,
// names, external types, or generated wrappers.
func compareProtobufMeta(left, right expr.MetaExpr) int {
	for _, name := range []string{
		"rpc:tag",
		"struct:field:proto",
		"struct:name:proto",
		"proto:tag:json",
		wrappedAttrMeta,
	} {
		if order := compareStringList(left[name], right[name]); order != 0 {
			return order
		}
	}
	return 0
}

// compareStringList compares both presence and contents because a missing list
// and an empty list are separate metadata values.
func compareStringList(left, right []string) int {
	if order := compareBool(left != nil, right != nil); order != 0 {
		return order
	}
	if order := cmp.Compare(len(left), len(right)); order != 0 {
		return order
	}
	for index, value := range left {
		if order := cmp.Compare(value, right[index]); order != 0 {
			return order
		}
	}
	return 0
}

// compareBool orders false before true.
func compareBool(left, right bool) int {
	switch {
	case left == right:
		return 0
	case left:
		return 1
	default:
		return -1
	}
}

// uniqueProtobufSourceName adds a number until the name is unused in the set.
func uniqueProtobufSourceName(preferred string, used map[string]struct{}) string {
	for index := 1; ; index++ {
		name := preferred
		if index > 1 {
			name += strconv.Itoa(index)
		}
		if _, ok := used[name]; ok {
			continue
		}
		used[name] = struct{}{}
		return name
	}
}

// assignProtobufGroupName stores both the selected name and its numeric suffix
// so later collision retries do not need to parse the name.
func assignProtobufGroupName(group *protobufNameGroup, used map[string]struct{}) {
	for suffix := 1; ; suffix++ {
		name := group.preferred
		if suffix > 1 {
			name += strconv.Itoa(suffix)
		}
		if _, ok := used[name]; ok {
			continue
		}
		used[name] = struct{}{}
		group.name = name
		group.suffix = suffix
		return
	}
}

// nextAvailableProtobufName returns the next numbered name that is not used by
// another message, service, or method in the same protobuf file.
func nextAvailableProtobufName(changed *protobufNameGroup, groups []*protobufNameGroup) string {
	preferred := changed.preferred
	index := max(changed.suffix, 1)
	for {
		index++
		candidate := preferred + strconv.Itoa(index)
		available := true
		for _, group := range groups {
			if group == changed || group.name != candidate {
				continue
			}
			if (group.method == nil) == (changed.method == nil) {
				available = false
				break
			}
		}
		if available {
			changed.suffix = index
			return candidate
		}
	}
}

// namingField creates one field whose type is sufficient for Go name planning.
func namingField(name string, number int32, oneof *int32) *descriptorpb.FieldDescriptorProto {
	field := &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: &number,
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
	}
	if oneof != nil {
		field.OneofIndex = oneof
	}
	return field
}

// protobufSourceFieldName returns the field spelling written to .proto.
func protobufSourceFieldName(name string) string {
	return codegen.ProtobufFieldName(name)
}

// uniqueProtobufOneofSourceName adds a suffix until the oneof name differs
// from every field, branch, and earlier oneof in the message.
func uniqueProtobufOneofSourceName(fieldName string, used map[string]struct{}) string {
	name := codegen.ProtobufFieldName(fieldName)
	for {
		if _, ok := used[name]; !ok {
			used[name] = struct{}{}
			return name
		}
		name += "_oneof"
	}
}

// serviceNameKeys returns every package name written for one service.
func serviceNameKeys(descriptor string) []protocNameKey {
	roles := []protocNameRole{
		protocServiceClientName,
		protocServiceClientStructName,
		protocServiceClientConstructorName,
		protocServiceServerName,
		protocServiceUnimplementedServerName,
		protocServiceUnsafeServerName,
		protocServiceRegisterName,
		protocServiceDescriptorName,
	}
	keys := make([]protocNameKey, len(roles))
	for index, role := range roles {
		keys[index] = protocNameKey{descriptor: descriptor, role: role}
	}
	return keys
}

// methodNameKeys returns every name written for one method.
func methodNameKeys(descriptor string, streaming bool) []protocNameKey {
	roles := []protocNameRole{protocMethodName, protocMethodFullName, protocMethodHandlerName}
	if streaming {
		roles = append(roles, protocMethodClientStreamName, protocMethodServerStreamName)
	}
	keys := make([]protocNameKey, len(roles))
	for index, role := range roles {
		keys[index] = protocNameKey{descriptor: descriptor, role: role}
	}
	return keys
}

// serviceFullName returns the current service descriptor name.
func (p *protobufServicePlan) serviceFullName() string {
	return p.protoPackage + "." + p.serviceName
}

// slicesIndexEndpoint returns endpoint's position in endpoints.
func slicesIndexEndpoint(endpoints []*expr.GRPCEndpointExpr, endpoint *expr.GRPCEndpointExpr) int {
	for index, candidate := range endpoints {
		if candidate == endpoint {
			return index
		}
	}
	return -1
}
