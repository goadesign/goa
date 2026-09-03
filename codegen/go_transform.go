// This file generates Go transformations between compatible design types.
// Recursive helpers carry the package path for each side through nested named
// declarations so emitted references use the package selected during planning.
package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

type (
	// transformCopyAttributor passes the original expression to name
	// lookups that recorded expressions before the transform copied them.
	transformCopyAttributor struct {
		attributor Attributor
		copier     *expr.AttributeGraphCopier
	}

	// transformAttributePair identifies one exact pair in a plan.
	transformAttributePair struct {
		source *expr.AttributeExpr
		target *expr.AttributeExpr
	}

	// transformUnwrapChoice is the result of one planned UnwrapPair call.
	transformUnwrapChoice struct {
		source    *expr.AttributeExpr
		target    *expr.AttributeExpr
		directive *WrapDirective
	}

	// transformStructuralChoices remembers what the two structural hooks
	// returned while the transform was planned.
	transformStructuralChoices struct {
		unwrap           func(*expr.AttributeExpr, *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective)
		fieldPair        func(*expr.AttributeExpr, *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr)
		planUnionHelpers func(*expr.AttributeExpr, *expr.AttributeExpr, func(*expr.AttributeExpr, *expr.AttributeExpr))
		unchanged        func() bool
		mutationErr      error
		sourceCopier     *expr.AttributeGraphCopier
		targetCopier     *expr.AttributeGraphCopier
		unwrapPairs      map[transformAttributePair]transformUnwrapChoice
		fieldPairs       map[transformAttributePair]transformAttributePair
		planned          bool
	}
)

var transformGoArrayT, transformGoMapT, transformGoUnionT *template.Template

// NOTE: can't initialize inline because https://github.com/golang/go/issues/1817
func init() {
	fm := template.FuncMap{
		"transformAttribute":  TransformAttribute,
		"transformHelperName": TransformHelperName,
	}
	transformGoArrayT = template.Must(template.New("transformGoArray").Funcs(fm).Parse(codegenTemplates.Read(transformGoArrayTmplName)))
	transformGoMapT = template.Must(template.New("transformGoMap").Funcs(fm).Parse(codegenTemplates.Read(transformGoMapTmplName)))
	transformGoUnionT = template.Must(template.New("transformGoUnion").Funcs(fm).Parse(codegenTemplates.Read(transformGoUnionTmplName)))
}

// original returns the caller expression that was copied into the plan.
func (a *transformCopyAttributor) original(attribute *expr.AttributeExpr) *expr.AttributeExpr {
	return a.copier.Original(attribute)
}

// Name gives the wrapped resolver the expression it used during name planning.
func (a *transformCopyAttributor) Name(attribute *expr.AttributeExpr, pkg string, pointer, useDefault bool) string {
	return a.attributor.Name(a.original(attribute), pkg, pointer, useDefault)
}

// Ref gives the wrapped resolver the expression it used during name planning.
func (a *transformCopyAttributor) Ref(attribute *expr.AttributeExpr, pkg string) string {
	return a.attributor.Ref(a.original(attribute), pkg)
}

// Field gives the wrapped resolver the field it used during name planning.
func (a *transformCopyAttributor) Field(attribute *expr.AttributeExpr, name string, firstUpper bool) string {
	return a.attributor.Field(a.original(attribute), name, firstUpper)
}

// Package gives the wrapped resolver the expression it used during planning.
func (a *transformCopyAttributor) Package(attribute *expr.AttributeExpr) string {
	return a.attributor.Package(a.original(attribute))
}

// Enter retains the translation while entering the caller's planned child.
func (a *transformCopyAttributor) Enter(attribute *expr.AttributeExpr) Attributor {
	return &transformCopyAttributor{
		attributor: a.attributor.Enter(a.original(attribute)),
		copier:     a.copier,
	}
}

// IsSumType reports the layout selected by the caller's attributor.
func (a *transformCopyAttributor) IsSumType() bool {
	return a.attributor.IsSumType()
}

// ValidatorCall gives the wrapped resolver the expression it used during
// validation name planning.
func (a *transformCopyAttributor) ValidatorCall(attribute *expr.AttributeExpr, view, target, path string) string {
	return a.attributor.ValidatorCall(a.original(attribute), view, target, path)
}

// Scope returns the name scope owned by the caller's resolver.
func (a *transformCopyAttributor) Scope() *NameScope {
	return a.attributor.Scope()
}

// OneofWrapper forwards the gRPC oneof lookup when the wrapped resolver owns
// that lookup. A different resolver cannot render a protobuf oneof.
func (a *transformCopyAttributor) OneofWrapper(attribute *expr.AttributeExpr) string {
	resolver, ok := a.attributor.(interface {
		OneofWrapper(*expr.AttributeExpr) string
	})
	if !ok {
		panic("transform name resolver cannot resolve a protobuf oneof wrapper") // bug
	}
	return resolver.OneofWrapper(a.original(attribute))
}

// UnionConstructor gives a copied transform attribute to the resolver that
// retained the generated service constructor.
func (a *transformCopyAttributor) UnionConstructor(attribute *expr.AttributeExpr, branch string) (string, error) {
	resolver, ok := a.attributor.(interface {
		UnionConstructor(*expr.AttributeExpr, string) (string, error)
	})
	if !ok {
		return "", fmt.Errorf("transform name resolver cannot resolve a service OneOf constructor")
	}
	return resolver.UnionConstructor(a.original(attribute), branch)
}

// GoTypeLayout gives the wrapped resolver the expression it used while
// planning the generated type.
func (a *transformCopyAttributor) GoTypeLayout(attribute *expr.AttributeExpr, policy GoLayoutPolicy) (LinkedGoType, error) {
	resolver, ok := a.attributor.(GoTypeLayoutResolver)
	if !ok {
		return planGoTypeWithAttributor(a.original(attribute), policy, a.attributor)
	}
	return resolver.GoTypeLayout(a.original(attribute), policy)
}

// captureTransformStructuralHooks copies the hook set and replaces planning
// callbacks with memoized versions. Planning may add a choice; rendering may
// only read one that planning already made. unchanged checks that a hook left
// the plan-owned expression graphs intact.
func captureTransformStructuralHooks(hooks *TransformHooks, sourceCopier, targetCopier *expr.AttributeGraphCopier, unchanged func() bool) (*TransformHooks, *transformStructuralChoices) {
	if hooks == nil {
		return nil, nil
	}
	copied := *hooks
	choices := &transformStructuralChoices{
		unwrap:           copied.UnwrapPair,
		fieldPair:        copied.FieldPairAttrs,
		planUnionHelpers: copied.PlanUnionHelpers,
		unchanged:        unchanged,
		sourceCopier:     sourceCopier,
		targetCopier:     targetCopier,
		unwrapPairs:      make(map[transformAttributePair]transformUnwrapChoice),
		fieldPairs:       make(map[transformAttributePair]transformAttributePair),
	}
	if choices.unwrap != nil {
		copied.UnwrapPair = choices.unwrapPair
	}
	if choices.fieldPair != nil {
		copied.FieldPairAttrs = choices.fieldPairAttrs
	}
	if choices.planUnionHelpers != nil {
		copied.PlanUnionHelpers = choices.planUnionHelpersForPlan
	}
	return &copied, choices
}

// unwrapPair returns the first attributes and wrapper instruction chosen for
// pair. After planning, a missing choice means rendering took a new path.
func (c *transformStructuralChoices) unwrapPair(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
	pair := transformAttributePair{source: source, target: target}
	if choice, ok := c.unwrapPairs[pair]; ok {
		return choice.source, choice.target, choice.directive
	}
	if c.planned {
		panic("transform render requested an unplanned unwrap choice") // bug
	}
	source, target, directive := c.unwrap(source, target)
	c.recordMutation("UnwrapPair")
	source = c.sourceCopier.Copy(source)
	target = c.targetCopier.Copy(target)
	if directive != nil {
		copy := *directive
		copy.Target = c.targetCopier.Copy(directive.Target)
		directive = &copy
	}
	c.unwrapPairs[pair] = transformUnwrapChoice{
		source:    source,
		target:    target,
		directive: directive,
	}
	return source, target, directive
}

// fieldPairAttrs returns the first attributes chosen for pair.
func (c *transformStructuralChoices) fieldPairAttrs(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
	pair := transformAttributePair{source: source, target: target}
	if choice, ok := c.fieldPairs[pair]; ok {
		return choice.source, choice.target
	}
	if c.planned {
		panic("transform render requested an unplanned field-pair choice") // bug
	}
	source, target = c.fieldPair(source, target)
	c.recordMutation("FieldPairAttrs")
	source = c.sourceCopier.Copy(source)
	target = c.targetCopier.Copy(target)
	c.fieldPairs[pair] = transformAttributePair{source: source, target: target}
	return source, target
}

// planUnionHelpersForPlan records the union helper choices and rejects a hook
// that changed a source or target expression while deciding those choices.
func (c *transformStructuralChoices) planUnionHelpersForPlan(source, target *expr.AttributeExpr, record func(*expr.AttributeExpr, *expr.AttributeExpr)) {
	c.planUnionHelpers(source, target, record)
	c.recordMutation("PlanUnionHelpers")
}

// recordMutation preserves the first planning-hook violation so
// NewTransformPlan can return it instead of a downstream compatibility error.
func (c *transformStructuralChoices) recordMutation(name string) {
	if c.mutationErr == nil && !c.unchanged() {
		c.mutationErr = fmt.Errorf("transform planning hook %s changed the retained plan", name)
	}
}

// GoTransform produces Go code that initializes the data structure defined
// by target from an instance of the data structure described by source.
// The data structures can be objects, arrays or maps. The algorithm
// matches object fields by name and ignores object fields in target that
// don't have a match in source. The matching and generated code leverage
// mapped attributes so that attribute names may use the "name:elem"
// syntax to define the name of the design attribute and the name of the
// corresponding generated Go struct field. The object field may also differ
// in that they may be pointers in one case and not the other. The function
// returns an error if target is not compatible with source (different type,
// fields of different type etc).
//
// As a special case GoTransform can map a union to or from an object envelope.
// The envelope has a string "Type" field that names the selected branch and a
// "Value" field that contains that branch.
//
// source and target are the attributes used in the transformation
//
// sourceVar and targetVar are the variable names used in the transformation
//
// sourceCtx and targetCtx are the attribute contexts for the source and target
// attributes
//
// prefix is the transformation helper function prefix
//
// newVar if true initializes a target variable with the generated Go code
// using `:=` operator. If false, it assigns Go code to the target variable
// using `=`.
func GoTransform(source, target *expr.AttributeExpr, sourceVar, targetVar string, sourceCtx, targetCtx *AttributeContext, prefix string, newVar bool) (string, []*TransformFunctionData, error) {
	return GoTransformWithAttrs(source, target, sourceVar, targetVar, &TransformAttrs{
		SourceCtx: sourceCtx,
		TargetCtx: targetCtx,
		Prefix:    prefix,
	}, newVar)
}

// GoTransformWithAttrs is GoTransform with a caller built TransformAttrs. It
// plans the conversion before writing it so structural hooks run once. Returned
// helpers keep their released Name and a nil Declaration for existing plugins.
// Generators may customize the conversion through TransformAttrs.Hooks; see
// TransformHooks.
func GoTransformWithAttrs(source, target *expr.AttributeExpr, sourceVar, targetVar string, ta *TransformAttrs, newVar bool) (string, []*TransformFunctionData, error) {
	plan, err := NewTransformPlan(source, target, ta.Prefix, ta.Hooks)
	if err != nil {
		return "", nil, err
	}
	if err := plan.BindContexts(ta.SourceCtx, ta.TargetCtx); err != nil {
		return "", nil, err
	}

	// Existing callers choose exact helper names while writing a file. Use one
	// declaration for every occurrence of the same released name so Render can
	// verify that the shared function body is identical.
	legacyPackage := newGeneratedPackage("legacy transform helpers", "goa.design/goa/v3/codegen/transform", "")
	declarations := make(map[string]*NameDeclaration, len(plan.helpers))
	nameAttrs := &TransformAttrs{
		SourceCtx: plan.sourceCtx,
		TargetCtx: plan.targetCtx,
		Prefix:    plan.prefix,
	}
	for _, helper := range plan.helpers {
		name := legacyTransformHelperName(helper.Source, helper.Target, nameAttrs)
		declaration := declarations[name]
		if declaration == nil {
			declaration = NewExactName(NameFunction, name)
			if err := legacyPackage.DeclareName(declaration); err != nil {
				return "", nil, err
			}
			declarations[name] = declaration
		}
		if err := plan.BindHelperDeclaration(helper.ID, declaration); err != nil {
			return "", nil, err
		}
	}
	if err := legacyPackage.freeze(); err != nil {
		return "", nil, err
	}
	code, helpers, err := plan.Render(sourceVar, targetVar, newVar)
	if err != nil {
		return "", nil, err
	}
	for _, helper := range helpers {
		helper.ID = TransformHelperID{}
		helper.Declaration = nil
	}
	return code, helpers, nil
}

// NewTransformProgram copies non-nil hooks into a conversion program that can
// make several plans. The program exposes no mutable state, so a package-level
// helper registry may compare plans by their shared program identity. Callers
// that need Goa's built-in conversion use NewTransformPlan with nil hooks. It
// returns an error when hooks is nil.
func NewTransformProgram(hooks *TransformHooks) (*TransformProgram, error) {
	if hooks == nil {
		return nil, fmt.Errorf("transform program hooks must not be nil")
	}
	copied := *hooks
	return &TransformProgram{hooks: &copied}, nil
}

// Plan copies source and target and records the conversion selected by the
// program. Plans made by one program use the same immutable hook choices.
func (p *TransformProgram) Plan(source, target *expr.AttributeExpr, prefix string) (*TransformPlan, error) {
	if p == nil {
		return nil, fmt.Errorf("transform program must not be nil")
	}
	return newTransformPlan(source, target, prefix, p)
}

// NewTransformPlan copies the source and target expression graphs and records
// every recursive conversion needed to turn one into the other. Distinct user
// types remain distinct even when they were copied from one declaration, and a
// true recursive edge points back to the same copied type. The plan walks only
// these copies during Render. It retains the original expression identities so
// name resolvers can find names they recorded before this call.
//
// NewTransformPlan calls UnwrapPair, FieldPairAttrs, and PlanUnionHelpers while
// planning and keeps their returned choices. The rendering hooks are called
// during Render. The caller may reuse or change source, target, and the
// TransformHooks value after this function returns; the plan owns its
// expression copies and its copied hook fields. A planning hook must inspect
// those copies without changing them; this function returns an error if it
// detects a mutation.
func NewTransformPlan(source, target *expr.AttributeExpr, prefix string, hooks *TransformHooks) (*TransformPlan, error) {
	if hooks == nil {
		return newTransformPlan(source, target, prefix, nil)
	}
	program, err := NewTransformProgram(hooks)
	if err != nil {
		return nil, err
	}
	return program.Plan(source, target, prefix)
}

// newTransformPlan records one conversion using program's retained hooks. A
// nil program selects Goa's built-in conversion and can share with every other
// plan that also has no custom hooks.
func newTransformPlan(source, target *expr.AttributeExpr, prefix string, program *TransformProgram) (*TransformPlan, error) {
	sourceCopier := expr.NewAttributeGraphCopier()
	targetCopier := expr.NewAttributeGraphCopier()
	source = sourceCopier.Copy(source)
	target = targetCopier.Copy(target)
	baselineSource := expr.NewAttributeGraphCopier()
	baselineTarget := expr.NewAttributeGraphCopier()
	plan := &TransformPlan{
		source:         source,
		target:         target,
		rootSource:     source,
		rootTarget:     target,
		sourceBaseline: baselineSource.Copy(source),
		targetBaseline: baselineTarget.Copy(target),
		sourceCopier:   sourceCopier,
		targetCopier:   targetCopier,
		prefix:         prefix,
		program:        program,
		operations:     []*transformOperation{{}},
		renders:        make(map[transformRenderRequest]transformRenderResult),
	}
	var hooks *TransformHooks
	if program != nil {
		hooks = program.hooks
	}
	retainedHooks, structuralChoices := captureTransformStructuralHooks(hooks, sourceCopier, targetCopier, func() bool {
		return !plan.changed()
	})
	plan.hooks = retainedHooks
	err := planTransformOperation(source, target, true, true, TransformHelperDefinitionLocation{}, plan.operations[0], make(map[transformPair]TransformHelperID), plan)
	if structuralChoices != nil && structuralChoices.mutationErr != nil {
		return nil, structuralChoices.mutationErr
	}
	if err != nil {
		return nil, err
	}
	if structuralChoices != nil {
		structuralChoices.planned = true
	}
	return plan, nil
}

// Helpers returns the recursive conversion functions selected by the plan so a
// generator can declare their names before writing code. Changing the returned
// slice or its Source and Target attributes does not change the plan. Source
// and Target identify the caller attributes that the plan copied; Render keeps
// and uses separate private attributes.
func (p *TransformPlan) Helpers() []TransformHelper {
	helpers := slices.Clone(p.helpers)
	for index := range helpers {
		sourceCopier := expr.NewAttributeGraphCopier()
		targetCopier := expr.NewAttributeGraphCopier()
		helpers[index].Source = sourceCopier.Copy(helpers[index].Source)
		helpers[index].Target = targetCopier.Copy(helpers[index].Target)
	}
	return helpers
}

// HelperDefinitions returns the distinct recursive conversion function bodies
// selected by the plan. Calls with the same retained source and target facts
// share one definition even when one call is required and another is optional.
// Changing the returned slice or its Source and Target attributes does not
// change the plan.
func (p *TransformPlan) HelperDefinitions() []TransformHelperDefinition {
	definitions := slices.Clone(p.definitions)
	for index := range definitions {
		sourceCopier := expr.NewAttributeGraphCopier()
		targetCopier := expr.NewAttributeGraphCopier()
		definitions[index].Source = sourceCopier.Copy(definitions[index].Source)
		definitions[index].Target = targetCopier.Copy(definitions[index].Target)
		definitions[index].helpers = nil
	}
	return definitions
}

// BindHelperDeclaration assigns the package-level function declared for one
// value returned by Helpers. Equivalent conversions may share a declaration;
// Render verifies that their generated definitions match.
func (p *TransformPlan) BindHelperDeclaration(id TransformHelperID, declaration *NameDeclaration) error {
	if id.plan != p || id.index < 0 || id.index >= len(p.helpers) {
		return fmt.Errorf("transform helper does not belong to this plan")
	}
	if declaration == nil {
		return fmt.Errorf("transform helper declaration must not be nil")
	}
	if declaration.Kind() != NameFunction {
		return fmt.Errorf("transform helper declaration must be a function, got %s", declaration.Kind())
	}
	helper := &p.helpers[id.index]
	if helper.Declaration != nil && helper.Declaration != declaration {
		return fmt.Errorf("transform helper already has a different declaration")
	}
	helper.Declaration = declaration
	return nil
}

// BindHelperDefinition assigns one package-level function declaration to every
// equivalent helper call represented by a value returned by HelperDefinitions.
func (p *TransformPlan) BindHelperDefinition(id TransformHelperDefinitionID, declaration *NameDeclaration) error {
	if id.plan != p || id.index < 0 || id.index >= len(p.definitions) {
		return fmt.Errorf("transform helper definition does not belong to this plan")
	}
	if declaration == nil {
		return fmt.Errorf("transform helper definition declaration must not be nil")
	}
	if declaration.Kind() != NameFunction {
		return fmt.Errorf("transform helper definition declaration must be a function, got %s", declaration.Kind())
	}
	definition := &p.definitions[id.index]
	if definition.Declaration != nil && definition.Declaration != declaration {
		return fmt.Errorf("transform helper definition already has a different declaration")
	}
	for _, index := range definition.helpers {
		if bound := p.helpers[index].Declaration; bound != nil && bound != declaration {
			return fmt.Errorf("transform helper occurrence %d already has a different declaration", index+1)
		}
	}
	definition.Declaration = declaration
	for _, index := range definition.helpers {
		p.helpers[index].Declaration = declaration
	}
	return nil
}

// BindContexts copies the source and target type resolvers used by every call
// and helper definition. Call it after helper declarations and package names
// are final. It may be called once.
func (p *TransformPlan) BindContexts(source, target *AttributeContext) error {
	if source == nil || target == nil {
		return fmt.Errorf("transform contexts must not be nil")
	}
	if p.sourceCtx != nil || p.targetCtx != nil {
		return fmt.Errorf("transform contexts are already bound")
	}
	p.sourceCtx = source.Dup()
	p.sourceCtx.Scope = &transformCopyAttributor{
		attributor: source.Scope,
		copier:     p.sourceCopier,
	}
	p.targetCtx = target.Dup()
	p.targetCtx.Scope = &transformCopyAttributor{
		attributor: target.Scope,
		copier:     p.targetCopier,
	}
	return nil
}

// Render writes the top-level conversion and its recursive function bodies.
// Every helper must have a declaration and BindContexts must have been called.
// Repeating the same source variable, target variable, and new-variable choice
// returns the exact first result without calling hooks again. Different
// variables are separate generation requests and can produce different code.
func (p *TransformPlan) Render(sourceVar, targetVar string, newVar bool) (code string, helpers []*TransformFunctionData, err error) {
	if p.sourceCtx == nil || p.targetCtx == nil {
		return "", nil, fmt.Errorf("transform contexts are not bound")
	}
	request := transformRenderRequest{sourceVar: sourceVar, targetVar: targetVar, newVar: newVar}
	if cached, ok := p.renders[request]; ok {
		return cached.code, copyTransformFunctionData(cached.helpers), cached.err
	}
	if p.changed() {
		return "", nil, fmt.Errorf("transform render hook changed the retained plan")
	}
	defer func() {
		if p.changed() {
			code = ""
			helpers = nil
			err = fmt.Errorf("transform render hook changed the retained plan")
		}
		p.renders[request] = transformRenderResult{
			code:    code,
			helpers: copyTransformFunctionData(helpers),
			err:     err,
		}
		helpers = copyTransformFunctionData(helpers)
	}()
	var hooks *TransformHooks
	if p.hooks != nil {
		copied := *p.hooks
		hooks = &copied
	}
	renderAttrs := TransformAttrs{
		SourceCtx: p.sourceCtx.Dup(),
		TargetCtx: p.targetCtx.Dup(),
		Prefix:    p.prefix,
		Hooks:     hooks,
	}
	renderAttrs.locals, err = newTransformLocalScope(sourceVar, targetVar)
	if err != nil {
		return "", nil, err
	}
	renderAttrs.helpers = make(map[TransformHelperID]TransformHelper, len(p.helpers))
	for _, planned := range p.helpers {
		if planned.Declaration == nil {
			return "", nil, fmt.Errorf("transform helper occurrence %d has no declaration", planned.Occurrence)
		}
		renderAttrs.helpers[planned.ID] = planned
	}
	renderAttrs.calls = &transformCallCursor{calls: p.operations[0].calls}
	code, err = TransformAttribute(p.source, p.target, sourceVar, targetVar, newVar, &renderAttrs)
	if err != nil {
		return "", nil, err
	}
	if err := renderAttrs.calls.complete("top-level transform"); err != nil {
		return "", nil, err
	}
	helpers = make([]*TransformFunctionData, 0, len(p.helpers))
	definitions := make(map[*NameDeclaration]*TransformFunctionData, len(p.helpers))
	for index, planned := range p.helpers {
		entered := enterTransformAttrs(planned.Source, planned.Target, &renderAttrs)
		entered.locals, err = newTransformLocalScope("v", "res")
		if err != nil {
			return "", nil, err
		}
		entered.calls = &transformCallCursor{calls: p.operations[index+1].calls}
		helper, err := generateTransformHelper(planned, entered)
		if err != nil {
			return "", nil, err
		}
		if err := entered.calls.complete(fmt.Sprintf("transform helper occurrence %d", planned.Occurrence)); err != nil {
			return "", nil, err
		}
		if previous := definitions[planned.Declaration]; previous != nil {
			if !transformFunctionDefinitionsEqual(previous, helper) {
				return "", nil, fmt.Errorf("transform helper declaration %q has different definitions", planned.Declaration.Name())
			}
			continue
		}
		definitions[planned.Declaration] = helper
		helpers = append(helpers, helper)
	}
	return strings.TrimRight(code, "\n"), helpers, nil
}

// copyTransformFunctionData copies generated helper descriptions before they
// cross the plan boundary. Declaration is intentionally shared: it is the
// immutable package-level name selected before rendering.
func copyTransformFunctionData(helpers []*TransformFunctionData) []*TransformFunctionData {
	if helpers == nil {
		return nil
	}
	copied := make([]*TransformFunctionData, len(helpers))
	for index, helper := range helpers {
		if helper == nil {
			continue
		}
		value := *helper
		copied[index] = &value
	}
	return copied
}

// newTransformLocalScope reserves every variable referenced by the source and
// target expressions. Generated temporary variables can then use ordinary
// names without hiding a parameter or another value used by the conversion.
func newTransformLocalScope(expressions ...string) (*NameScope, error) {
	names := make(map[string]struct{})
	for _, expression := range expressions {
		parsed, err := parser.ParseExpr(expression)
		if err != nil {
			return nil, fmt.Errorf("parse transform expression %q: %w", expression, err)
		}
		collectTransformExpressionNames(parsed, names)
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	slices.Sort(ordered)
	scope := NewNameScope()
	for _, name := range ordered {
		scope.Unique(name)
	}
	return scope, nil
}

// collectTransformExpressionNames records variables used by expression. A
// selector field such as Field in value.Field is not a variable in the current
// Go block and therefore does not reserve a local name.
func collectTransformExpressionNames(expression ast.Expr, names map[string]struct{}) {
	ast.Inspect(expression, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.SelectorExpr:
			collectTransformExpressionNames(node.X, names)
			return false
		case *ast.Ident:
			if node.Name != "_" {
				names[node.Name] = struct{}{}
			}
		}
		return true
	})
}

// changed reports whether code generation would read an expression different
// from the one retained when planning completed. Render hooks receive these
// expressions for inspection only; a mutation invalidates the render attempt.
func (p *TransformPlan) changed() bool {
	return !transformAttributesEqual(p.source, p.sourceBaseline, make(map[transformAttributePair]struct{})) ||
		!transformAttributesEqual(p.target, p.targetBaseline, make(map[transformAttributePair]struct{}))
}

// transformAttributesEqual compares the expression facts that conversion
// planning and rendering consume. It follows paired recursive attributes once.
func transformAttributesEqual(left, right *expr.AttributeExpr, seen map[transformAttributePair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	pair := transformAttributePair{source: left, target: right}
	if _, compared := seen[pair]; compared {
		return true
	}
	seen[pair] = struct{}{}
	if left.Description != right.Description || !reflect.DeepEqual(left.Docs, right.Docs) ||
		!reflect.DeepEqual(left.Validation, right.Validation) || !reflect.DeepEqual(left.Meta, right.Meta) ||
		!reflect.DeepEqual(left.DefaultValue, right.DefaultValue) || len(left.Bases) != len(right.Bases) ||
		len(left.References) != len(right.References) || len(left.UserExamples) != len(right.UserExamples) {
		return false
	}
	for index := range left.Bases {
		if !transformDataTypesEqual(left.Bases[index], right.Bases[index], seen) {
			return false
		}
	}
	for index := range left.References {
		if !transformDataTypesEqual(left.References[index], right.References[index], seen) {
			return false
		}
	}
	for index := range left.UserExamples {
		if left.UserExamples[index] == nil || right.UserExamples[index] == nil {
			if left.UserExamples[index] != right.UserExamples[index] {
				return false
			}
			continue
		}
		if left.UserExamples[index].Summary != right.UserExamples[index].Summary ||
			left.UserExamples[index].Description != right.UserExamples[index].Description ||
			!reflect.DeepEqual(left.UserExamples[index].Value, right.UserExamples[index].Value) {
			return false
		}
	}
	return transformDataTypesEqual(left.Type, right.Type, seen)
}

// transformDataTypesEqual compares the expression type graph below two
// attributes without comparing DSL functions or expression pointer identities.
func transformDataTypesEqual(left, right expr.DataType, seen map[transformAttributePair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false
	}
	switch left := left.(type) {
	case expr.Primitive:
		return left == right.(expr.Primitive)
	case *expr.Array:
		right := right.(*expr.Array)
		return left.NonNullableElems == right.NonNullableElems && transformAttributesEqual(left.ElemType, right.ElemType, seen)
	case *expr.Object:
		right := right.(*expr.Object)
		if len(*left) != len(*right) {
			return false
		}
		for index, attribute := range *left {
			other := (*right)[index]
			if attribute.Name != other.Name || !transformAttributesEqual(attribute.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case *expr.Map:
		right := right.(*expr.Map)
		return transformAttributesEqual(left.KeyType, right.KeyType, seen) &&
			transformAttributesEqual(left.ElemType, right.ElemType, seen)
	case *expr.Union:
		right := right.(*expr.Union)
		if left.TypeName != right.TypeName || left.TypeKey != right.TypeKey || left.ValueKey != right.ValueKey || len(left.Values) != len(right.Values) {
			return false
		}
		for index, attribute := range left.Values {
			other := right.Values[index]
			if attribute.Name != other.Name || !transformAttributesEqual(attribute.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case expr.UserType:
		right := right.(expr.UserType)
		return left.Origin() == right.Origin() && left.Name() == right.Name() &&
			transformAttributesEqual(left.Attribute(), right.Attribute(), seen)
	default:
		panic(fmt.Sprintf("cannot compare transform type %T", left)) // bug
	}
}

// TransformAttribute returns the code to transform source attribute to target
// attribute. It returns an error if source and target are not compatible for
// transformation. It is exported so that TransformHooks implementations can
// recurse into the transform engine from the code they render.
func TransformAttribute(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if ta.locals == nil {
		var err error
		ta.locals, err = newTransformLocalScope(sourceVar, targetVar)
		if err != nil {
			return "", err
		}
	}
	var (
		prelude                      string
		sourcePointer, targetPointer bool
	)
	if h := ta.Hooks; h != nil && h.UnwrapPair != nil {
		wrappedSource, wrappedTarget := source, target
		var dir *WrapDirective
		source, target, dir = h.UnwrapPair(source, target)
		prelude = dir.apply(&sourceVar, &targetVar, &newVar, ta)
		if dir != nil {
			if dir.WrapTarget {
				targetPointer = wrappedPrimitivePointer(wrappedTarget, ta.TargetCtx)
			} else {
				sourcePointer = wrappedPrimitivePointer(wrappedSource, ta.SourceCtx)
			}
		}
	}
	ta = enterTransformAttrs(source, target, ta)
	if err := IsCompatible(source.Type, target.Type, sourceVar, targetVar); err != nil {
		return "", err
	}
	var (
		code string
		err  error
	)
	switch {
	case expr.IsArray(source.Type):
		if h := ta.Hooks; h != nil && h.TransformArray != nil {
			code, err = h.TransformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), sourceVar, targetVar, newVar, ta)
		} else {
			code, err = transformArray(expr.AsArray(source.Type), expr.AsArray(target.Type), sourceVar, targetVar, newVar, ta)
		}
	case expr.IsMap(source.Type):
		if h := ta.Hooks; h != nil && h.TransformMap != nil {
			code, err = h.TransformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), sourceVar, targetVar, newVar, ta)
		} else {
			code, err = transformMap(expr.AsMap(source.Type), expr.AsMap(target.Type), sourceVar, targetVar, newVar, ta)
		}
	case expr.IsUnion(source.Type):
		if h := ta.Hooks; h != nil && h.TransformUnion != nil {
			code, err = h.TransformUnion(source, target, sourceVar, targetVar, newVar, nil, nil, ta)
		} else {
			code, err = transformUnion(source, target, sourceVar, targetVar, newVar, ta)
		}
	case expr.IsObject(source.Type):
		code, err = transformObject(source, target, sourceVar, targetVar, newVar, ta)
	default:
		code = transformPrimitive(source, target, sourceVar, targetVar, newVar, sourcePointer, targetPointer, ta)
	}
	if err != nil {
		return "", err
	}
	return prelude + code, nil
}

// TransformHelperName returns the recursive function used to initialize target
// from source. A TransformPlan calls it only for named object pairs.
// GoTransformWithAttrs still chooses its helper name while writing code.
func TransformHelperName(source, target *expr.AttributeExpr, ta *TransformAttrs) string {
	if ta.calls != nil {
		call := ta.calls.consume()
		helper, ok := ta.helpers[call.helper]
		if !ok {
			panic("planned transform call references an unknown helper") // bug
		}
		return helper.Declaration.Name()
	}
	return legacyTransformHelperName(source, target, ta)
}

// legacyTransformHelperName preserves the naming strategy used by generators
// that have not yet bound their package-level helper declarations.
func legacyTransformHelperName(source, target *expr.AttributeExpr, ta *TransformAttrs) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		ta = enterTransformAttrs(source, target, ta)
		sname = Goify(ta.SourceCtx.Scope.Name(source, ta.SourceCtx.Pkg(source), ta.SourceCtx.Pointer, ta.SourceCtx.UseDefault), true)
		tname = Goify(ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault), true)
		prefix = ta.Prefix
		if prefix == "" {
			prefix = "transform"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}

// usesTransformHelper reports whether source and target can define one named
// helper function signature. Anonymous objects are rendered inline because
// they have no package-level parameter or result declaration.
func usesTransformHelper(source, target *expr.AttributeExpr) bool {
	_, sourceNamed := source.Type.(expr.UserType)
	_, targetNamed := target.Type.(expr.UserType)
	return sourceNamed && targetNamed && expr.IsObject(source.Type) && expr.IsObject(target.Type)
}

// transformFunctionDefinitionsEqual reports whether one function can serve
// every call assigned to the same declaration.
func transformFunctionDefinitionsEqual(left, right *TransformFunctionData) bool {
	return left.ParamTypeRef == right.ParamTypeRef &&
		left.ResultTypeRef == right.ResultTypeRef &&
		left.Code == right.Code
}

// transformPrimitive returns the code to transform source primitive type to
// target primitive type. The caller (TransformAttribute) already verified that
// source and target are compatible. The pointer flags describe primitive
// fields inside wrappers removed before this function runs.
func transformPrimitive(
	source, target *expr.AttributeExpr,
	sourceVar, targetVar string,
	newVar, sourcePointer, targetPointer bool,
	ta *TransformAttrs,
) string {
	assign := "="
	if newVar {
		assign = ":="
	}

	handled := false
	expression := sourceVar
	if h := ta.Hooks; h != nil && h.ConvertPrimitive != nil {
		if result, ok := h.ConvertPrimitive(source, target, sourceVar, sourcePointer, targetPointer, ta); ok {
			expression = result
			handled = true
		}
	}
	if !handled {
		srcRef := ta.SourceCtx.Scope.Ref(source, ta.SourceCtx.Pkg(source))
		tgtRef := ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target))
		if sourcePointer {
			expression = "*" + expression
		}
		if srcRef != tgtRef {
			expression = fmt.Sprintf("%s(%s)", tgtRef, expression)
		}
	}
	if targetPointer {
		targetRef := ta.TargetCtx.Scope.Ref(target, ta.TargetCtx.Pkg(target))
		return fmt.Sprintf("%s = new(%s)\n*%s = %s\n", targetVar, targetRef, targetVar, expression)
	}
	return fmt.Sprintf("%s %s %s\n", targetVar, assign, expression)
}

// wrappedPrimitivePointer reports whether the single field exposed by a
// wrapper uses a pointer in the generated source or target type.
func wrappedPrimitivePointer(wrapper *expr.AttributeExpr, context *AttributeContext) bool {
	object := expr.AsObject(wrapper.Type)
	if object == nil || len(*object) != 1 {
		panic("transform wrapper must contain exactly one field")
	}
	field := (*object)[0]
	return expr.IsPrimitive(field.Attribute.Type) && context.IsPrimitivePointer(field.Name, wrapper)
}

// transformObject generates Go code to transform source object to target
// object.
func transformObject(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	var (
		initCode     string
		postInitCode string
		err          error
	)
	{
		// walk through primitives first to initialize the struct
		walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
			if err != nil {
				return
			}
			if !expr.IsPrimitive(srcc.Type) {
				return
			}
			// Source and/or target could be primitive user type. Make sure the
			// aliased type is compatible for transformation.
			if err = IsCompatible(srcc.Type, tgtc.Type, sourceVar, targetVar); err != nil {
				return
			}
			var (
				exp string

				srcPtr     = ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr)
				tgtPtr     = ta.TargetCtx.IsPrimitivePointer(n, tgtMatt.AttributeExpr)
				srcField   = sourceVar + "." + ta.SourceCtx.Scope.Field(srcc, srcMatt.ElemName(n), true)
				tgtField   = ta.TargetCtx.Scope.Field(tgtc, tgtMatt.ElemName(n), true)
				_, isSrcUT = srcc.Type.(expr.UserType)
				_, isTgtUT = tgtc.Type.(expr.UserType)
			)
			{
				var (
					convExp string
					hasConv bool
				)
				if h := ta.Hooks; h != nil && h.ConvertPrimitive != nil {
					convExp, hasConv = h.ConvertPrimitive(srcc, tgtc, srcField, srcPtr, tgtPtr, ta)
				}
				switch {
				case hasConv && (isSrcUT || isTgtUT || convExp != srcField), !hasConv && (isSrcUT || isTgtUT):
					if hasConv {
						exp = convExp
					} else {
						deref := ""
						if srcPtr {
							deref = "*"
						}
						exp = fmt.Sprintf("%s(%s%s)", ta.TargetCtx.Scope.Ref(tgtc, ta.TargetCtx.Pkg(tgtc)), deref, srcField)
					}
					if srcPtr && !srcMatt.IsRequired(n) {
						postInitCode += fmt.Sprintf("if %s != nil {\n", srcField)
						if tgtPtr {
							tmp := ta.enterLocalBlock().uniqueLocal(Goify(tgtMatt.ElemName(n), false))
							postInitCode += fmt.Sprintf("%s := %s\n%s.%s = &%s\n", tmp, exp, targetVar, tgtField, tmp)
						} else {
							postInitCode += fmt.Sprintf("%s.%s = %s\n", targetVar, tgtField, exp)
						}
						postInitCode += "}\n"
						return
					} else if tgtPtr {
						tmp := ta.uniqueLocal(Goify(tgtMatt.ElemName(n), false))
						postInitCode += fmt.Sprintf("%s := %s\n%s.%s = &%s\n", tmp, exp, targetVar, tgtField, tmp)
						return
					}
				case srcPtr && !tgtPtr:
					exp = "*" + srcField
					if !srcMatt.IsRequired(n) {
						postInitCode += fmt.Sprintf("if %s != nil {\n\t%s.%s = %s\n}\n", srcField, targetVar, tgtField, exp)
						return
					}
				case !srcPtr && tgtPtr:
					exp = "&" + srcField
				default:
					exp = srcField
				}
			}
			initCode += fmt.Sprintf("\n%s: %s,", tgtField, exp)
		})
		if initCode != "" {
			initCode += "\n"
		}
	}
	if err != nil {
		return "", err
	}

	buffer := &bytes.Buffer{}
	deref := "&"
	if h := ta.Hooks; h != nil && h.ObjectDeref != nil {
		if d, ok := h.ObjectDeref(target); ok {
			deref = d
		}
	}
	assign := "="
	if newVar {
		assign = ":="
	}
	name := ta.TargetCtx.Scope.Name(target, ta.TargetCtx.Pkg(target), ta.TargetCtx.Pointer, ta.TargetCtx.UseDefault)
	fmt.Fprintf(buffer, "%s %s %s%s{%s}\n", targetVar, assign, deref, name, initCode)
	buffer.WriteString(postInitCode)

	// iterate through attributes to initialize rest of the struct fields and
	// handle default values
	walkMatches(source, target, func(srcMatt, tgtMatt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string) {
		if err != nil {
			return
		}
		srcField := ta.SourceCtx.Scope.Field(srcc, srcMatt.ElemName(n), true)
		tgtField := ta.TargetCtx.Scope.Field(tgtc, tgtMatt.ElemName(n), true)
		h := ta.Hooks
		if h != nil && h.FieldPairAttrs != nil {
			srcc, tgtc = h.FieldPairAttrs(srcc, tgtc)
		}
		var (
			srcVar = sourceVar + "." + srcField
			tgtVar = targetVar + "." + tgtField
		)
		var dir *WrapDirective
		if h != nil && h.UnwrapPair != nil {
			srcc, tgtc, dir = h.UnwrapPair(srcc, tgtc)
		}
		if err = IsCompatible(srcc.Type, tgtc.Type, sourceVar, targetVar); err != nil {
			return
		}
		fieldAttrs := enterTransformAttrs(srcc, tgtc, ta)

		var code string
		{
			// The wrap directive (if any) only redirects the code
			// transforming the field value: nil guards and default value
			// handling keep using the unwrapped field variables.
			dispatchSrcVar, dispatchTgtVar, dispatchNewVar := srcVar, tgtVar, false
			prelude := dir.apply(&dispatchSrcVar, &dispatchTgtVar, &dispatchNewVar, ta)
			var postlude string
			if expr.IsUnion(tgtc.Type) && ta.TargetCtx.IsFieldPointer(n, tgtMatt.AttributeExpr) {
				unionVar := Goify(tgtMatt.ElemName(n), false) + "Value"
				unionRef := fieldAttrs.TargetCtx.Scope.Name(tgtc, fieldAttrs.TargetCtx.Pkg(tgtc), false, fieldAttrs.TargetCtx.UseDefault)
				prelude += fmt.Sprintf("var %s %s\n", unionVar, unionRef)
				dispatchTgtVar = unionVar
				postlude = fmt.Sprintf("%s = &%s\n", tgtVar, unionVar)
			}
			switch {
			case expr.IsArray(srcc.Type):
				if h != nil && h.TransformArray != nil {
					code, err = h.TransformArray(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, fieldAttrs)
				} else {
					code, err = transformArray(expr.AsArray(srcc.Type), expr.AsArray(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, fieldAttrs)
				}
			case expr.IsMap(srcc.Type):
				if h != nil && h.TransformMap != nil {
					code, err = h.TransformMap(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, fieldAttrs)
				} else {
					code, err = transformMap(expr.AsMap(srcc.Type), expr.AsMap(tgtc.Type), dispatchSrcVar, dispatchTgtVar, dispatchNewVar, fieldAttrs)
				}
			case expr.IsUnion(srcc.Type):
				if h != nil && h.TransformUnion != nil {
					code, err = h.TransformUnion(srcc, tgtc, dispatchSrcVar, dispatchTgtVar, dispatchNewVar, source, target, fieldAttrs)
				} else {
					code, err = transformUnion(srcc, tgtc, dispatchSrcVar, dispatchTgtVar, dispatchNewVar, fieldAttrs)
				}
			case usesTransformHelper(srcc, tgtc):
				code = fmt.Sprintf("%s = %s(%s)\n", dispatchTgtVar, TransformHelperName(srcc, tgtc, ta), dispatchSrcVar)
			case expr.IsObject(srcc.Type):
				code, err = TransformAttribute(srcc, tgtc, dispatchSrcVar, dispatchTgtVar, dispatchNewVar, ta)
			}
			if code != "" {
				code = prelude + code + postlude
			}
		}
		if err != nil {
			return
		}

		// We need to check for a nil source if it holds a reference (pointer to
		// primitive or an object, array or map) and is not required. We also want
		// to always check nil if the attribute is not a primitive; it's a
		// 1) user type and we want to avoid calling transform helper functions
		// with nil value
		// 2) it's an object, map or array to avoid making empty arrays and maps
		// and to avoid derefencing nil.
		var guarded bool
		if h != nil && h.GuardCondition != nil {
			var cond string
			if cond, guarded = h.GuardCondition(srcc, srcVar, srcMatt.IsRequired(n), ta.SourceCtx.IsFieldPointer(n, srcMatt.AttributeExpr)); guarded && cond != "" && code != "" {
				code = fmt.Sprintf("%s\t%s}\n", cond, code)
			}
		}
		if !guarded {
			var checkNil bool
			{
				isRef := !expr.IsPrimitive(srcc.Type) && !srcMatt.IsRequired(n) || ta.SourceCtx.IsPrimitivePointer(n, srcMatt.AttributeExpr) && expr.IsPrimitive(srcc.Type)
				marshalNonPrimitive := !expr.IsPrimitive(srcc.Type) && ta.SourceCtx.UseDefault && ta.TargetCtx.UseDefault
				checkNil = isRef || marshalNonPrimitive
			}
			if code != "" && checkNil {
				cond := fmt.Sprintf("if %s != nil {\n", srcVar)
				// A pointer-backed union uses nil as its sole absence value. Preserve a
				// non-nil zero union so validation rejects its missing discriminator.
				if expr.IsUnion(srcc.Type) && !ta.SourceCtx.IsFieldPointer(n, srcMatt.AttributeExpr) {
					cond = fmt.Sprintf("if %s.Kind() != \"\" {\n", srcVar)
				}
				code = fmt.Sprintf("%s\t%s}", cond, code)
				if expr.IsArray(srcc.Type) && srcMatt.IsRequired(n) {
					elem := expr.AsArray(tgtc.Type).ElemType
					code += fmt.Sprintf("else {\n\t%s = []%s{}\n}\n", tgtVar, fieldAttrs.TargetCtx.Scope.Ref(elem, fieldAttrs.TargetCtx.Pkg(elem)))
				} else {
					code += "\n"
				}
			}
		}

		// Default value handling. We need to handle default values if the target
		// type uses default values (i.e. attributes with default values are
		// non-pointers) and has a default value set.
		if tdef := tgtMatt.GetDefault(n); tdef != nil && ta.TargetCtx.UseDefault && !ta.TargetCtx.Pointer && !srcMatt.IsRequired(n) {
			switch {
			case transformFieldUsesNilPresence(ta.SourceCtx, n, srcMatt.AttributeExpr):
				// A nil source means the field was omitted. This includes
				// slices and message values as well as explicit pointers.
				assignment, renderErr := renderTransformDefault(
					tgtc,
					tdef,
					fieldAttrs.TargetCtx,
					ta.TargetCtx.IsFieldPointer(n, tgtMatt.AttributeExpr),
					tgtVar,
					tgtMatt.ElemName(n)+"Default",
				)
				if renderErr != nil {
					err = renderErr
					return
				}
				code += fmt.Sprintf("if %s == nil {\n%s}\n", srcVar, Indent(assignment, "\t"))
			case expr.IsPrimitive(srcc.Type) && srcMatt.HasDefaultValue(n) && ta.SourceCtx.UseDefault:
				// source attribute is a primitive with default value
				// (the field is not a pointer in this case)
				assignment, renderErr := renderTransformDefault(
					tgtc,
					tdef,
					fieldAttrs.TargetCtx,
					ta.TargetCtx.IsFieldPointer(n, tgtMatt.AttributeExpr),
					tgtVar,
					tgtMatt.ElemName(n)+"Default",
				)
				if renderErr != nil {
					err = renderErr
					return
				}
				code += "{\n\t"
				var zeroName string
				nilable := IsNilable(tgtc.Type) || valueIsNilable(tdef)
				if h != nil && h.ZeroTypeName != nil {
					if name, ok := h.ZeroTypeName(tgtc); ok {
						zeroName = name
					}
				}
				if zeroName == "" {
					zeroName = ta.TargetCtx.Scope.Ref(tgtc, ta.TargetCtx.Pkg(tgtc))
				}
				if !nilable {
					code += fmt.Sprintf("var zero %s\n\t", zeroName)
					code += fmt.Sprintf("if %s == zero ", tgtVar)
				} else {
					code += fmt.Sprintf("if %s == nil ", tgtVar)
				}
				code += fmt.Sprintf("{\n%s}\n", Indent(assignment, "\t"))
				code += "}\n"
			}
		}
		buffer.WriteString(code)
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// renderTransformDefault writes the declarations and assignment needed when a
// source field is absent. The target context supplies the exact generated Go
// names and pointer layout.
func renderTransformDefault(attribute *expr.AttributeExpr, value any, context *AttributeContext, pointer bool, target, localPrefix string) (string, error) {
	resolver, ok := context.Scope.(GoTypeLayoutResolver)
	if !ok {
		return "", fmt.Errorf("render transform default: target package has no linked Go type resolver")
	}
	layout, err := resolver.GoTypeLayout(attribute, context.LayoutPolicy())
	if err != nil {
		return "", err
	}
	var resolveUnion UnionConstructorResolver
	if resolver, ok := context.Scope.(interface {
		UnionConstructor(*expr.AttributeExpr, string) (string, error)
	}); ok {
		resolveUnion = resolver.UnionConstructor
	}
	rendered, err := RenderGoValue(attribute, value, layout, pointer, resolveUnion, localPrefix)
	if err != nil {
		return "", err
	}
	var code strings.Builder
	for _, declaration := range rendered.Declarations {
		code.WriteString(declaration)
		code.WriteByte('\n')
	}
	fmt.Fprintf(&code, "%s = %s\n", target, rendered.Expression)
	return code.String(), nil
}

// transformFieldUsesNilPresence reports whether nil means that a generated
// source field was omitted. Bytes and Any can preserve this distinction even
// though Goa does not add another pointer around them.
func transformFieldUsesNilPresence(context *AttributeContext, name string, parent *expr.AttributeExpr) bool {
	field := expr.AsObject(parent.Type).Attribute(name)
	return !expr.IsPrimitive(field.Type) ||
		context.IsPrimitivePointer(name, parent) ||
		!context.UseDefault && IsNilable(field.Type)
}

// valueIsNilable reports whether a typed default can be compared only with
// nil. It preserves named slices, maps, pointers, functions, and channels
// without inspecting the Go name supplied by field metadata.
func valueIsNilable(value any) bool {
	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice
}

// transformArray generates Go code to transform source array to target array.
func transformArray(source, target *expr.Array, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.ElemType.Type, target.ElemType.Type, sourceVar+"[0]", targetVar+"[0]"); err != nil {
		return "", err
	}
	sourceElement := "val"
	if ta.SourceCtx.IsArrayElementPointer(source) {
		sourceElement = "*val"
	}
	loopVar, childAttrs := ta.EnterCollection()
	localExpressions := []string{sourceVar, targetVar, loopVar, "val"}
	if ta.TargetCtx.IsArrayElementPointer(target) {
		localExpressions = append(localExpressions, "transformed")
	}
	childAttrs, err := childAttrs.EnterLocalBlock(localExpressions...)
	if err != nil {
		return "", err
	}
	data := map[string]any{
		"ElemTypeRef":       ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg(target.ElemType)),
		"SourceElem":        source.ElemType,
		"SourceElement":     sourceElement,
		"TargetElem":        target.ElemType,
		"SourceVar":         sourceVar,
		"TargetVar":         targetVar,
		"NewVar":            newVar,
		"TransformAttrs":    childAttrs,
		"LoopVar":           loopVar,
		"SourceIsObject":    expr.IsObject(source.ElemType.Type),
		"TargetElemPointer": ta.TargetCtx.IsArrayElementPointer(target),
		"UseHelper":         usesTransformHelper(source.ElemType, target.ElemType),
	}
	var buf bytes.Buffer
	if err := transformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transformMap generates Go code to transform source map to target map.
func transformMap(source, target *expr.Map, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if err := IsCompatible(source.KeyType.Type, target.KeyType.Type, sourceVar+"[key]", targetVar+"[key]"); err != nil {
		return "", err
	}
	if err := IsCompatible(source.ElemType.Type, target.ElemType.Type, sourceVar+"[*]", targetVar+"[*]"); err != nil {
		return "", err
	}
	loopVar := ""
	if depth := MapDepth(target); depth > 0 {
		loopVar = string(rune(97 + depth))
	}
	mapAttrs, err := ta.EnterLocalBlock(sourceVar, targetVar, "key", "val", "tk", "tv"+loopVar)
	if err != nil {
		return "", err
	}
	data := map[string]any{
		"KeyTypeRef":     ta.TargetCtx.Scope.Ref(target.KeyType, ta.TargetCtx.Pkg(target.KeyType)),
		"ElemTypeRef":    ta.TargetCtx.Scope.Ref(target.ElemType, ta.TargetCtx.Pkg(target.ElemType)),
		"SourceKey":      source.KeyType,
		"TargetKey":      target.KeyType,
		"SourceElem":     source.ElemType,
		"TargetElem":     target.ElemType,
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TransformAttrs": mapAttrs,
		"LoopVar":        loopVar,
		"ElemIsObject":   expr.IsObject(source.ElemType.Type),
		"UseKeyHelper":   usesTransformHelper(source.KeyType, target.KeyType),
		"UseElemHelper":  usesTransformHelper(source.ElemType, target.ElemType),
	}
	var buf bytes.Buffer
	if err := transformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// transformUnion generates Go code to transform source union to target union.
//
// Note: transport to/from service transforms are always object to union or
// union to object. The only case a transform is union to union is when
// converting a projected type from/to a service type.
func transformUnion(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, ta *TransformAttrs) (string, error) {
	if !expr.IsUnion(target.Type) {
		return "", fmt.Errorf("cannot transform union %s to non-union %s", source.Type.Name(), target.Type.Name())
	}
	srcUnion, tgtUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	if len(srcUnion.Values) != len(tgtUnion.Values) {
		return "", fmt.Errorf("cannot transform union: number of union types differ (%s has %d, %s has %d)",
			source.Type.Name(), len(srcUnion.Values), target.Type.Name(), len(tgtUnion.Values))
	}
	for i, st := range srcUnion.Values {
		if err := IsCompatible(st.Attribute.Type, tgtUnion.Values[i].Attribute.Type, sourceVar, targetVar); err != nil {
			return "", fmt.Errorf("cannot transform union %s to %s: type at index %d: %w",
				source.Type.Name(), target.Type.Name(), i, err)
		}
	}

	// Unions are generated as concrete sum-type structs with Kind/AsX/SetX
	// helpers. Transform by branching on the runtime Kind discriminator.
	unionPkg := ta.TargetCtx.Pkg(target)
	typeRef := ta.TargetCtx.Scope.Ref(target, unionPkg)

	// The outer union keeps Goa's released local spelling. Nested unions use
	// numbered locals selected from traversal depth, never from caller code.
	tempVarName := "obj"
	if ta.unionDepth > 0 {
		tempVarName = "tmp"
		tempVarName += strconv.Itoa(ta.unionDepth + 1)
	}
	childAttrs := *ta
	childAttrs.unionDepth++

	cases := make([]map[string]any, 0, len(srcUnion.Values))
	for i, st := range srcUnion.Values {
		tt := tgtUnion.Values[i]
		branchAttrs := *ta
		branchAttrs.SourceCtx = ta.SourceCtx.Enter(st.Attribute)
		branchAttrs.TargetCtx = ta.TargetCtx.Enter(tt.Attribute)
		useHelper := usesTransformHelper(st.Attribute, tt.Attribute)
		helperName := ""
		if useHelper {
			helperName = TransformHelperName(st.Attribute, tt.Attribute, &branchAttrs)
		}
		cases = append(cases, map[string]any{
			"CaseName":        st.Name,
			"SourceFieldName": Goify(st.Name, true),
			"TargetFieldName": Goify(tt.Name, true),
			"SourceAttr":      st.Attribute,
			"TargetAttr":      tt.Attribute,
			"TargetCastType":  branchAttrs.TargetCtx.Scope.Ref(tt.Attribute, branchAttrs.TargetCtx.Pkg(tt.Attribute)),
			"SourceNilable":   IsNilable(st.Attribute.Type),
			"UseHelper":       useHelper,
			"HelperName":      helperName,
		})
	}

	data := map[string]any{
		"SourceVar":      sourceVar,
		"TargetVar":      targetVar,
		"NewVar":         newVar,
		"TypeRef":        typeRef,
		"ValueTypeRef":   ta.TargetCtx.Scope.Name(target, unionPkg, false, ta.TargetCtx.UseDefault),
		"TempVarName":    tempVarName,
		"Cases":          cases,
		"TransformAttrs": &childAttrs,
	}

	var buf bytes.Buffer
	if err := transformGoUnionT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// planTransformOperation records the recursive calls made by the main
// conversion or one generated function. It reuses a function when that same
// source and target pair is already being converted.
func planTransformOperation(source, target *expr.AttributeExpr, required, topLevel bool, location TransformHelperDefinitionLocation, operation *transformOperation, active map[transformPair]TransformHelperID, plan *TransformPlan) error {
	return planTransformOperationWithHelper(source, target, required, topLevel, false, location, operation, active, plan)
}

// planTransformOperationWithHelper records one conversion. forceHelper is true
// when a custom union renderer declares that it calls TransformHelperName for
// this pair, including named arrays and aliases that the default renderer
// writes inline.
func planTransformOperationWithHelper(source, target *expr.AttributeExpr, required, topLevel, forceHelper bool, location TransformHelperDefinitionLocation, operation *transformOperation, active map[transformPair]TransformHelperID, plan *TransformPlan) error {
	helperSource, helperTarget := source, target
	if forceHelper {
		_, sourceNamed := source.Type.(expr.UserType)
		_, targetNamed := target.Type.(expr.UserType)
		if !sourceNamed && !targetNamed {
			return fmt.Errorf("custom union transform helper requires a named source or target type")
		}
	}
	var rootWrap *WrapDirective
	if plan.hooks != nil && plan.hooks.UnwrapPair != nil {
		source, target, rootWrap = plan.hooks.UnwrapPair(source, target)
	}
	if location.encoded == "" {
		plan.rootSource = source
		plan.rootTarget = target
		plan.rootWrap = rootWrap
	}
	if !forceHelper {
		helperSource, helperTarget = source, target
	}
	if err := IsCompatible(source.Type, target.Type, "source", "target"); err != nil {
		return err
	}
	if topLevel {
		required = true
	} else if forceHelper || usesTransformHelper(source, target) {
		pair := transformPair{
			source: source.Type,
			target: target.Type,
		}
		if ancestor, recursive := active[pair]; recursive {
			operation.calls = append(operation.calls, transformCall{
				helper: ancestor,
			})
			return nil
		}

		id := TransformHelperID{plan: plan, index: len(plan.helpers)}
		plan.helpers = append(plan.helpers, TransformHelper{
			ID:         id,
			Source:     helperSource,
			Target:     helperTarget,
			Required:   required,
			Occurrence: id.index + 1,
			location:   location,
		})
		plan.recordHelperDefinition(id, helperSource, helperTarget, location)
		operation.calls = append(operation.calls, transformCall{
			helper: id,
		})
		body := &transformOperation{}
		plan.operations = append(plan.operations, body)
		active[pair] = id
		bodySource, bodyTarget := source, target
		if !forceHelper && plan.hooks != nil && plan.hooks.UnwrapPair != nil {
			bodySource, bodyTarget, _ = plan.hooks.UnwrapPair(bodySource, bodyTarget)
		}
		if err := IsCompatible(bodySource.Type, bodyTarget.Type, "source", "target"); err != nil {
			delete(active, pair)
			return err
		}
		err := planTransformChildren(bodySource, bodyTarget, required, location, body, active, plan)
		delete(active, pair)
		return err
	}
	return planTransformChildren(source, target, required, location, operation, active, plan)
}

// recordHelperDefinition groups calls with the same retained source and target
// facts. Hooks may inspect metadata, defaults, and validation when writing a
// function body, so matching only the data type would merge different code.
// Requiredness stays on the call because it only decides whether the caller
// checks for nil.
func (p *TransformPlan) recordHelperDefinition(id TransformHelperID, source, target *expr.AttributeExpr, location TransformHelperDefinitionLocation) {
	for index := range p.definitions {
		definition := &p.definitions[index]
		same := transformAttributesEqual(definition.Source, source, make(map[transformAttributePair]struct{})) &&
			transformAttributesEqual(definition.Target, target, make(map[transformAttributePair]struct{}))
		if p.hooks != nil && p.hooks.SameHelperDefinition != nil {
			same = sameTransformUserTypeOrigin(definition.Source, source) &&
				sameTransformUserTypeOrigin(definition.Target, target) &&
				p.hooks.SameHelperDefinition(definition.Source, definition.Target, source, target)
		}
		if same {
			definition.helpers = append(definition.helpers, id.index)
			if location.Compare(definition.Location) < 0 {
				definition.Location = location
			}
			return
		}
	}
	definitionID := TransformHelperDefinitionID{plan: p, index: len(p.definitions)}
	p.definitions = append(p.definitions, TransformHelperDefinition{
		ID:       definitionID,
		Source:   source,
		Target:   target,
		Location: location,
		helpers:  []int{id.index},
	})
}

// sameTransformUserTypeOrigin prevents a hook from merging functions for two
// separately authored named types. A named type's origin owns its generated Go
// declaration even when another type has the same name and fields.
func sameTransformUserTypeOrigin(first, next *expr.AttributeExpr) bool {
	firstType, firstNamed := first.Type.(expr.UserType)
	nextType, nextNamed := next.Type.(expr.UserType)
	if firstNamed != nextNamed {
		return false
	}
	return !firstNamed || firstType.Origin() == nextType.Origin()
}

// append adds one authored path component. NUL bytes inside names are escaped,
// and two NUL bytes end each component, so empty and nested names remain
// distinct while string comparison preserves their lexical order.
func (l TransformHelperDefinitionLocation) append(kind byte, name string) TransformHelperDefinitionLocation {
	var encoded strings.Builder
	encoded.Grow(len(l.encoded) + len(name) + 3)
	encoded.WriteString(l.encoded)
	encoded.WriteByte(kind)
	for index := range len(name) {
		if name[index] == 0 {
			encoded.WriteByte(0)
			encoded.WriteByte(0xff)
			continue
		}
		encoded.WriteByte(name[index])
	}
	encoded.WriteByte(0)
	encoded.WriteByte(0)
	return TransformHelperDefinitionLocation{encoded: encoded.String()}
}

// objectField returns the location below one authored object field.
func (l TransformHelperDefinitionLocation) objectField(name string) TransformHelperDefinitionLocation {
	return l.append(transformObjectFieldLocation, name)
}

// arrayElement returns the location below an array element.
func (l TransformHelperDefinitionLocation) arrayElement() TransformHelperDefinitionLocation {
	return l.append(transformArrayElementLocation, "")
}

// mapKey returns the location below a map key.
func (l TransformHelperDefinitionLocation) mapKey() TransformHelperDefinitionLocation {
	return l.append(transformMapKeyLocation, "")
}

// mapValue returns the location below a map value.
func (l TransformHelperDefinitionLocation) mapValue() TransformHelperDefinitionLocation {
	return l.append(transformMapValueLocation, "")
}

// unionBranch returns the location below an authored union branch.
func (l TransformHelperDefinitionLocation) unionBranch(name string) TransformHelperDefinitionLocation {
	return l.append(transformUnionBranchLocation, name)
}

// transformUnionHelperBranch returns the authored branch name selected by a
// custom union hook. Both attributes must be the retained authored branch
// pair so the location cannot silently identify a different function body.
func transformUnionHelperBranch(source, target *expr.Union, sourceBranch, targetBranch *expr.AttributeExpr) (string, bool) {
	for index, branch := range source.Values {
		if branch.Attribute == sourceBranch && target.Values[index].Attribute == targetBranch {
			return branch.Name, true
		}
	}
	return "", false
}

// planTransformChildren walks child transformations in the same order as the
// core templates consume helper names.
func planTransformChildren(source, target *expr.AttributeExpr, required bool, location TransformHelperDefinitionLocation, operation *transformOperation, active map[transformPair]TransformHelperID, plan *TransformPlan) error {
	collect := func(source, target *expr.AttributeExpr, childRequired bool, top bool, childLocation TransformHelperDefinitionLocation) error {
		return planTransformOperation(source, target, childRequired, top, childLocation, operation, active, plan)
	}
	elementTop := plan.hooks != nil && plan.hooks.InlineCompositeElems
	switch {
	case expr.IsArray(source.Type):
		return collect(expr.AsArray(source.Type).ElemType, expr.AsArray(target.Type).ElemType, required, elementTop, location.arrayElement())
	case expr.IsMap(source.Type):
		sourceMap, targetMap := expr.AsMap(source.Type), expr.AsMap(target.Type)
		if err := collect(sourceMap.KeyType, targetMap.KeyType, required, elementTop, location.mapKey()); err != nil {
			return err
		}
		return collect(sourceMap.ElemType, targetMap.ElemType, required, elementTop, location.mapValue())
	case expr.IsUnion(source.Type):
		targetUnion := expr.AsUnion(target.Type)
		if targetUnion == nil {
			return nil
		}
		sourceUnion := expr.AsUnion(source.Type)
		if plan.hooks != nil && plan.hooks.TransformUnion != nil {
			if plan.hooks.PlanUnionHelpers == nil {
				return nil
			}
			var planErr error
			plan.hooks.PlanUnionHelpers(source, target, func(sourceBranch, targetBranch *expr.AttributeExpr) {
				if planErr != nil {
					return
				}
				branch, ok := transformUnionHelperBranch(sourceUnion, targetUnion, sourceBranch, targetBranch)
				if !ok {
					planErr = fmt.Errorf("custom union transform helper must use retained authored branch attributes")
					return
				}
				planErr = planTransformOperationWithHelper(sourceBranch, targetBranch, required, false, true, location.unionBranch(branch), operation, active, plan)
			})
			return planErr
		}
		if len(sourceUnion.Values) != len(targetUnion.Values) {
			return fmt.Errorf("cannot transform union: number of union types differ (%s has %d, %s has %d)",
				source.Type.Name(), len(sourceUnion.Values), target.Type.Name(), len(targetUnion.Values))
		}
		for index, branch := range sourceUnion.Values {
			if err := IsCompatible(branch.Attribute.Type, targetUnion.Values[index].Attribute.Type, "source", "target"); err != nil {
				return fmt.Errorf("cannot transform union %s to %s: type at index %d: %w",
					source.Type.Name(), target.Type.Name(), index, err)
			}
		}
		for index, branch := range sourceUnion.Values {
			if err := collect(branch.Attribute, targetUnion.Values[index].Attribute, required, false, location.unionBranch(branch.Name)); err != nil {
				return err
			}
		}
	case expr.IsObject(source.Type):
		if expr.IsUnion(target.Type) {
			return nil
		}
		var walkErr error
		walkMatches(source, target, func(sourceMapped, _ *expr.MappedAttributeExpr, sourceChild, targetChild *expr.AttributeExpr, name string) {
			if walkErr == nil {
				if plan.hooks != nil && plan.hooks.FieldPairAttrs != nil {
					sourceChild, targetChild = plan.hooks.FieldPairAttrs(sourceChild, targetChild)
				}
				walkErr = collect(sourceChild, targetChild, sourceMapped.IsRequired(name), false, location.objectField(name))
			}
		})
		return walkErr
	}
	return nil
}

// consume returns the next recursive call recorded by NewTransformPlan.
func (c *transformCallCursor) consume() transformCall {
	if c.next >= len(c.calls) {
		panic("transform render consumed more helper calls than the plan retained") // bug
	}
	call := c.calls[c.next]
	c.next++
	return call
}

// complete reports whether Render skipped any recorded recursive calls.
func (c *transformCallCursor) complete(owner string) error {
	if c.next != len(c.calls) {
		return fmt.Errorf("%s rendered %d of %d retained helper calls", owner, c.next, len(c.calls))
	}
	return nil
}

// enterTransformAttrs returns a copy that looks up fields and types beneath the
// supplied source and target.
func enterTransformAttrs(source, target *expr.AttributeExpr, attributes *TransformAttrs) *TransformAttrs {
	entered := *attributes
	entered.SourceCtx = attributes.SourceCtx.Enter(source)
	entered.TargetCtx = attributes.TargetCtx.Enter(target)
	return &entered
}

// generateTransformHelper writes one recursive conversion function selected by
// TransformPlan.
func generateTransformHelper(helper TransformHelper, ta *TransformAttrs) (*TransformFunctionData, error) {
	code, err := TransformAttribute(helper.Source, helper.Target, "v", "res", true, ta)
	if err != nil {
		return nil, err
	}
	tfd := &TransformFunctionData{
		ID:            helper.ID,
		Declaration:   helper.Declaration,
		Name:          helper.Declaration.Name(),
		ParamTypeRef:  ta.SourceCtx.Scope.Ref(helper.Source, ta.SourceCtx.Pkg(helper.Source)),
		ResultTypeRef: ta.TargetCtx.Scope.Ref(helper.Target, ta.TargetCtx.Pkg(helper.Target)),
		Code:          code,
	}
	return tfd, nil
}

// walkMatches iterates through the attributes of source and looks for
// attributes with identical names in target. walkMatches calls the walker
// function for each pair of matched attributes. Both source and target must be
// objects or else walkMatches panics.
func walkMatches(source, target *expr.AttributeExpr, walker func(src, tgt *expr.MappedAttributeExpr, srcc, tgtc *expr.AttributeExpr, n string)) {
	srcMatt := expr.NewMappedAttributeExpr(source)
	tgtMatt := expr.NewMappedAttributeExpr(target)
	srcFields := originalMappedFields(source)
	tgtFields := originalMappedFields(target)
	srcObj := expr.AsObject(srcMatt.Type)
	tgtObj := expr.AsObject(tgtMatt.Type)
	for _, nat := range *srcObj {
		if att := tgtObj.Attribute(nat.Name); att != nil {
			walker(srcMatt, tgtMatt, srcFields[nat.Name], tgtFields[nat.Name], nat.Name)
		}
	}
}

// originalMappedFields returns each child under the name used for matching.
// The returned children are the values supplied by the caller, not copies.
func originalMappedFields(attribute *expr.AttributeExpr) map[string]*expr.AttributeExpr {
	object := expr.AsObject(attribute.Type)
	fields := make(map[string]*expr.AttributeExpr, len(*object))
	for _, named := range *object {
		name := strings.SplitN(named.Name, ":", 2)[0]
		fields[name] = named.Attribute
	}
	return fields
}
