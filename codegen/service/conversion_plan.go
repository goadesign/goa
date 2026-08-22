// This file retains external Go type conversions before package names freeze.
// The root plan assigns each operation to its generated receiver package,
// declares recursive helpers, and records every reflected package import.
package service

import (
	"cmp"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// externalConversionDirection identifies which side of a retained mapping
	// owns the generated receiver method.
	externalConversionDirection uint8

	// externalConversionNameOrder gives every recursive helper a stable place
	// in its generated package without depending on service traversal.
	externalConversionNameOrder struct {
		receiverID  string
		externalPkg string
		external    string
		direction   externalConversionDirection
		source      string
		target      string
		occurrence  int
		required    bool
	}

	// externalConversionFacts retains one exact reflected mapping and transform
	// graph from collection through linked render data.
	externalConversionFacts struct {
		direction         externalConversionDirection
		serviceName       string
		servicePath       string
		receiverID        string
		receiverAttribute *expr.AttributeExpr
		receiverType      *codegen.TypeDeclaration
		externalType      reflect.Type
		externalPath      string
		externalAlias     string
		externalAttribute *expr.AttributeExpr
		externalPackages  map[expr.UserType]string
		externalScope     *codegen.NameScope
		plan              *codegen.TransformPlan
		methodName        string
		data              *convertData
		helpers           []*codegen.TransformFunctionData
	}

	// externalConversionFileFacts groups the conversion operations and imports
	// emitted by one generated receiver package's convert.go file.
	externalConversionFileFacts struct {
		owner      *codegen.GeneratedPackage
		operations []*externalConversionFacts
		imports    retainedFileImports
	}

	// externalConversionIdentity identifies one receiver method contract across
	// every root in a service planning batch.
	externalConversionIdentity struct {
		receiver     *codegen.TypeDeclaration
		direction    externalConversionDirection
		externalType reflect.Type
		externalPath string
	}

	// externalConversionResolver qualifies each reflected named type with the
	// frozen alias for that type's own Go package.
	externalConversionResolver struct {
		scope    *codegen.AttributeScope
		packages map[expr.UserType]string
	}
)

const (
	externalConvertTo externalConversionDirection = iota + 1
	externalCreateFrom
)

// collectExternalConversions plans every mapping across the complete service
// run once per exact generated receiver package. A relocated receiver shared
// by roots therefore receives one method namespace and one convert.go file.
func collectExternalConversions(roots []*rootFacts, generation *codegen.Generation) error {
	files := make(map[*codegen.GeneratedPackage]*externalConversionFileFacts)
	fileRoots := make(map[*codegen.GeneratedPackage]*rootFacts)
	serviceRoots := make(map[*serviceFacts]*rootFacts)
	operations := make(map[externalConversionIdentity]struct{})
	for _, root := range roots {
		root.externalConversions = nil
		for _, service := range root.services {
			serviceRoots[service] = root
		}
	}
	collect := func(mappings []*expr.TypeMap, direction externalConversionDirection) error {
		for _, mapping := range mappings {
			owners := make(map[*codegen.GeneratedPackage]*serviceFacts)
			for _, candidate := range roots {
				for _, service := range candidate.services {
					if !typeMapMatchesFacts(mapping, service) {
						continue
					}
					owner := generation.Package(generatedPackagePath(
						generation.GenPkg(), service.service, codegen.UserTypeLocation(mapping.User),
					))
					selected := owners[owner]
					if selected == nil || service.packagePath < selected.packagePath {
						owners[owner] = service
					}
				}
			}
			orderedOwners := make([]*codegen.GeneratedPackage, 0, len(owners))
			for owner := range owners {
				orderedOwners = append(orderedOwners, owner)
			}
			sort.Slice(orderedOwners, func(i, j int) bool {
				return orderedOwners[i].ImportPath() < orderedOwners[j].ImportPath()
			})
			for _, owner := range orderedOwners {
				identity, externalAlias, err := identifyExternalConversion(mapping, owner, direction)
				if err != nil {
					return err
				}
				if _, exists := operations[identity]; exists {
					return fmt.Errorf(
						"duplicate external conversion for receiver %q in package %q and external type %q",
						mapping.User.ID(),
						identity.receiver.PackagePath(),
						identity.externalType.String(),
					)
				}
				operations[identity] = struct{}{}
				operation, err := planExternalConversion(
					owners[owner], mapping, owner, identity, externalAlias, generation,
				)
				if err != nil {
					return err
				}
				file := files[owner]
				if file == nil {
					file = &externalConversionFileFacts{owner: owner}
					files[owner] = file
				}
				file.operations = append(file.operations, operation)
				candidateRoot := serviceRoots[owners[owner]]
				selectedRoot := fileRoots[owner]
				if selectedRoot == nil || rootFactsOrder(candidateRoot) < rootFactsOrder(selectedRoot) {
					fileRoots[owner] = candidateRoot
				}
			}
		}
		return nil
	}
	for _, root := range roots {
		if err := collect(root.root.Conversions, externalConvertTo); err != nil {
			return err
		}
		if err := collect(root.root.Creations, externalCreateFrom); err != nil {
			return err
		}
	}

	for _, file := range files {
		if err := finishExternalConversionFile(file, generation); err != nil {
			return err
		}
		owner := fileRoots[file.owner]
		owner.externalConversions = append(owner.externalConversions, file)
	}
	for _, root := range roots {
		sort.Slice(root.externalConversions, func(i, j int) bool {
			return root.externalConversions[i].owner.ImportPath() < root.externalConversions[j].owner.ImportPath()
		})
	}
	return nil
}

// identifyExternalConversion resolves the complete run-wide receiver method
// identity before planning can declare helpers or imports for the operation.
func identifyExternalConversion(mapping *expr.TypeMap, owner *codegen.GeneratedPackage, direction externalConversionDirection) (externalConversionIdentity, string, error) {
	externalType := reflect.TypeOf(mapping.External)
	if externalType == nil {
		return externalConversionIdentity{}, "", fmt.Errorf("external conversion type must not be nil")
	}
	externalPath, externalAlias, err := getExternalReflectTypeInfo(externalType)
	if err != nil {
		return externalConversionIdentity{}, "", err
	}
	receiver, err := owner.Type(mapping.User)
	if err != nil {
		return externalConversionIdentity{}, "", err
	}
	return externalConversionIdentity{
		receiver:     receiver,
		direction:    direction,
		externalType: externalType,
		externalPath: externalPath,
	}, externalAlias, nil
}

// rootFactsOrder returns the stable service identity that owns shared file
// contributions selected from a root.
func rootFactsOrder(facts *rootFacts) string {
	paths := make([]string, len(facts.services))
	for index, service := range facts.services {
		paths[index] = service.packagePath
	}
	sort.Strings(paths)
	return facts.apiName + "\x00" + strings.Join(paths, "\x00")
}

// externalConversionFiles returns the package files assigned by batch
// planning and rejects duplicate ownership instead of merging after freeze.
func externalConversionFiles(plans []*Plan) ([]*externalConversionFileFacts, error) {
	byOwner := make(map[*codegen.GeneratedPackage]struct{})
	var files []*externalConversionFileFacts
	for _, plan := range plans {
		for _, retained := range plan.facts.externalConversions {
			if _, exists := byOwner[retained.owner]; exists {
				return nil, fmt.Errorf(
					"external conversion package %q was assigned to more than one service plan",
					retained.owner.ImportPath(),
				)
			}
			byOwner[retained.owner] = struct{}{}
			files = append(files, retained)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].owner.ImportPath() < files[j].owner.ImportPath()
	})
	return files, nil
}

// planExternalConversion reflects one external shape, builds its immutable
// transform graph, and binds each recursive helper to the receiver package.
func planExternalConversion(
	service *serviceFacts,
	mapping *expr.TypeMap,
	owner *codegen.GeneratedPackage,
	identity externalConversionIdentity,
	externalAlias string,
	generation *codegen.Generation,
) (*externalConversionFacts, error) {
	externalType := identity.externalType
	externalDataType, reflectedTypes, err := buildExternalDesignType(externalType, mapping.User)
	if err != nil {
		return nil, err
	}
	externalPackages := make(map[expr.UserType]string, len(reflectedTypes))
	for userType, reflected := range reflectedTypes {
		importPath, alias, err := getExternalReflectTypeInfo(reflected)
		if err != nil {
			return nil, err
		}
		if err := generation.DeclareImport(codegen.NewImport(alias, importPath)); err != nil {
			return nil, err
		}
		externalPackages[userType.Origin()] = importPath
	}
	externalPath := identity.externalPath
	externalAttribute := &expr.AttributeExpr{Type: externalDataType}
	if identity.direction == externalConvertTo {
		externalAttribute.AddMeta("struct:type:name", externalDataType.Name())
	}
	receiverAttribute := expr.DupAtt(&expr.AttributeExpr{Type: mapping.User})
	source := receiverAttribute
	target := externalAttribute
	if identity.direction == externalCreateFrom {
		source, target = externalAttribute, source
	}
	transform, err := codegen.NewTransformPlan(source, target)
	if err != nil {
		return nil, err
	}
	operation := &externalConversionFacts{
		direction:         identity.direction,
		serviceName:       service.name,
		servicePath:       service.packagePath,
		receiverID:        mapping.User.ID(),
		receiverAttribute: receiverAttribute,
		externalType:      externalType,
		externalPath:      externalPath,
		externalAlias:     externalAlias,
		externalAttribute: externalAttribute,
		externalPackages:  externalPackages,
		externalScope:     codegen.NewNameScope(),
		plan:              transform,
		receiverType:      identity.receiver,
	}
	for _, helper := range transform.Helpers() {
		sourceName, sourceID := transformDataTypeName(helper.Source.Type)
		targetName, targetID := transformDataTypeName(helper.Target.Type)
		if identity.direction == externalConvertTo {
			targetName = externalAlias + codegen.Goify(targetName, true)
		} else {
			sourceName = externalAlias + codegen.Goify(sourceName, true)
		}
		order := externalConversionNameOrder{
			receiverID:  mapping.User.ID(),
			externalPkg: externalPath,
			external:    externalType.Name(),
			direction:   identity.direction,
			source:      sourceID,
			target:      targetID,
			occurrence:  helper.Occurrence,
			required:    helper.Required,
		}
		declaration := codegen.NewPreferredName(
			codegen.NameFunction,
			"transform"+codegen.Goify(sourceName, true)+"To"+codegen.Goify(targetName, true),
			codegen.UnexportedName,
			order,
		)
		if err := owner.DeclareName(declaration); err != nil {
			return nil, err
		}
		if err := transform.BindHelperDeclaration(helper.ID, declaration); err != nil {
			return nil, err
		}
	}
	return operation, nil
}

// finishExternalConversionFile fixes canonical operation order, assigns
// receiver-scoped method names, and plans the exact imports used by convert.go.
func finishExternalConversionFile(file *externalConversionFileFacts, generation *codegen.Generation) error {
	sort.Slice(file.operations, func(i, j int) bool {
		return externalConversionOperationLess(file.operations[i], file.operations[j])
	})
	takenByReceiver := make(map[*codegen.TypeDeclaration]map[string]struct{})
	for _, operation := range file.operations {
		receiver := operation.receiverType
		taken := takenByReceiver[receiver]
		if taken == nil {
			taken = make(map[string]struct{})
			takenByReceiver[receiver] = taken
		}
		prefix := "ConvertTo"
		if operation.direction == externalCreateFrom {
			prefix = "CreateFrom"
		}
		operation.methodName = uniquify(prefix+operation.externalType.Name(), taken)
	}

	definitions := make([]*expr.AttributeExpr, 0, len(file.operations)*2)
	references := make([]*expr.AttributeExpr, 0, len(file.operations)*2)
	for _, operation := range file.operations {
		definitions = append(definitions, operation.receiverAttribute, operation.externalAttribute)
		references = append(references, operation.receiverAttribute, operation.externalAttribute)
	}
	imports, err := retainFileImports(
		generation,
		file.owner.ImportPath(),
		nil,
		nil,
		definitions,
		references,
	)
	if err != nil {
		return err
	}
	for _, operation := range file.operations {
		for _, importPath := range operation.externalPackages {
			addRetainedImportPath(&imports, importPath)
		}
	}
	file.imports = imports
	return nil
}

// externalConversionOperationLess orders retained receiver operations by their
// complete semantic identity independently of root and service traversal.
func externalConversionOperationLess(left, right *externalConversionFacts) bool {
	if left.receiverID != right.receiverID {
		return left.receiverID < right.receiverID
	}
	if left.direction != right.direction {
		return left.direction < right.direction
	}
	if left.externalPath != right.externalPath {
		return left.externalPath < right.externalPath
	}
	return left.externalType.Name() < right.externalType.Name()
}

// linkExternalConversions binds frozen type resolvers and formats every
// retained operation without reflecting types or discovering helpers.
func linkExternalConversions(
	facts *rootFacts,
	generation *codegen.Generation,
	aliases *importAliases,
) error {
	for _, file := range facts.externalConversions {
		linkFileImports(&file.imports, generation)
		for _, operation := range file.operations {
			serviceResolver := newRetainedServiceResolver(
				generation,
				aliases,
				operation.serviceName,
				operation.servicePath,
				file.owner.ImportPath(),
			)
			if err := linkExternalConversion(operation, serviceResolver, aliases); err != nil {
				return err
			}
		}
	}
	return nil
}

// linkExternalConversion renders one retained graph with frozen service and
// reflected-package aliases selected for its output file.
func linkExternalConversion(
	operation *externalConversionFacts,
	serviceResolver *declarationResolver,
	aliases *importAliases,
) error {
	externalResolver := newExternalConversionResolver(
		operation.externalScope,
		operation.externalPackages,
		aliases,
	)
	externalContext := &codegen.AttributeContext{
		Scope: externalResolver,
	}
	serviceContext := &codegen.AttributeContext{
		UseDefault: true,
		Scope:      serviceResolver,
	}
	sourceContext, targetContext := serviceContext, externalContext
	sourceVar, targetVar := "t", "v"
	if operation.direction == externalCreateFrom {
		sourceContext, targetContext = externalContext, serviceContext
		sourceVar, targetVar = "v", "temp"
	}
	if err := operation.plan.BindContexts(sourceContext, targetContext); err != nil {
		return err
	}
	code, helpers, err := operation.plan.Render(sourceVar, targetVar, true)
	if err != nil {
		return err
	}
	operation.data = &convertData{
		Name:            operation.methodName,
		ReceiverTypeRef: "*" + operation.receiverType.Name(),
		TypeRef: externalResolver.Ref(
			operation.externalAttribute,
			externalResolver.Package(operation.externalAttribute),
		),
		Code: code,
	}
	if operation.direction == externalConvertTo {
		operation.data.TypeName = operation.externalType.Name()
	}
	operation.helpers = helpers
	return nil
}

// newExternalConversionResolver binds reflected user types to aliases selected
// by the generation-wide import catalog.
func newExternalConversionResolver(
	scope *codegen.NameScope,
	packages map[expr.UserType]string,
	aliases *importAliases,
) *externalConversionResolver {
	resolved := make(map[expr.UserType]string, len(packages))
	for userType, importPath := range packages {
		resolved[userType.Origin()] = aliases.name(importPath)
	}
	return &externalConversionResolver{
		scope:    codegen.NewAttributeScope(scope),
		packages: resolved,
	}
}

// Name renders an external reflected type with the alias for its own package.
func (r *externalConversionResolver) Name(att *expr.AttributeExpr, pkg string, ptr, useDefault bool) string {
	if userType, ok := att.Type.(expr.UserType); ok {
		pkg = r.packageName(userType)
	}
	return r.scope.Name(att, pkg, ptr, useDefault)
}

// Ref renders an external reflected type reference with its own package alias.
func (r *externalConversionResolver) Ref(att *expr.AttributeExpr, pkg string) string {
	if userType, ok := att.Type.(expr.UserType); ok {
		pkg = r.packageName(userType)
	}
	return r.scope.Ref(att, pkg)
}

// Field returns the reflected Go struct field selected for name.
func (r *externalConversionResolver) Field(att *expr.AttributeExpr, name string, firstUpper bool) string {
	return r.scope.Field(att, name, firstUpper)
}

// Package returns the frozen alias for an external named type.
func (r *externalConversionResolver) Package(att *expr.AttributeExpr) string {
	if userType, ok := att.Type.(expr.UserType); ok {
		return r.packageName(userType)
	}
	return ""
}

// Enter keeps the resolver because each nested named type selects its package
// independently from the attribute currently being transformed.
func (r *externalConversionResolver) Enter(*expr.AttributeExpr) codegen.Attributor {
	return r
}

// IsSumType reports the standard Goa transform representation.
func (r *externalConversionResolver) IsSumType() bool {
	return r.scope.IsSumType()
}

// ValidatorName is not part of external conversion rendering.
func (*externalConversionResolver) ValidatorName(*expr.AttributeExpr, string) string {
	panic("external conversion resolver does not own validators")
}

// Scope returns the lexical scope used for reflected field and local names.
func (r *externalConversionResolver) Scope() *codegen.NameScope {
	return r.scope.Scope()
}

// packageName returns the exact frozen alias for one reflected user type.
func (r *externalConversionResolver) packageName(userType expr.UserType) string {
	name, ok := r.packages[userType.Origin()]
	if !ok {
		panic(fmt.Sprintf("external reflected type %q has no planned package alias", userType.Name()))
	}
	return name
}

// ComparePackageName orders external conversion helpers by complete semantic
// operation identity rather than discovery order.
func (o externalConversionNameOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	right := other.(externalConversionNameOrder)
	required := 0
	if o.required != right.required {
		if o.required {
			required = 1
		} else {
			required = -1
		}
	}
	return cmp.Or(
		strings.Compare(o.receiverID, right.receiverID),
		strings.Compare(o.externalPkg, right.externalPkg),
		strings.Compare(o.external, right.external),
		cmp.Compare(o.direction, right.direction),
		strings.Compare(o.source, right.source),
		strings.Compare(o.target, right.target),
		cmp.Compare(o.occurrence, right.occurrence),
		required,
	)
}
