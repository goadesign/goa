// This file composes planned Go conversions with functions already owned by
// the generated package. A reused function owns its body and child calls;
// the conversion plan emits only its call and other reachable helper bodies.
package codegen

import "fmt"

// UseExistingHelperDefinition binds every equivalent call to an existing
// function in the package that renders this conversion. The caller owns that
// function's body and must give it the planned source and target Go signature.
// Call before BindContexts. Helpers and HelperDefinitions then exclude this
// body and children reached only through it; Render writes its calls without
// generating another implementation.
func (p *TransformPlan) UseExistingHelperDefinition(id TransformHelperDefinitionID, declaration *NameDeclaration) error {
	if p.sourceCtx != nil || p.targetCtx != nil {
		return fmt.Errorf("existing transform functions must be selected before contexts are bound")
	}
	if err := p.BindHelperDefinition(id, declaration); err != nil {
		return err
	}
	if p.existingHelpers == nil {
		p.existingHelpers = make(map[TransformHelperID]struct{})
	}
	for _, index := range p.definitions[id.index].helpers {
		p.existingHelpers[p.helpers[index].ID] = struct{}{}
	}
	return nil
}

// reachableHelpers follows calls from the conversion entry point. Existing
// functions terminate the walk because their owner supplies their child calls.
// Shared and recursive calls are visited once and keep their original IDs.
func (p *TransformPlan) reachableHelpers() map[TransformHelperID]struct{} {
	reachable := make(map[TransformHelperID]struct{}, len(p.helpers))
	operations := []*transformOperation{p.operations[0]}
	for len(operations) > 0 {
		operation := operations[len(operations)-1]
		operations = operations[:len(operations)-1]
		for _, call := range operation.calls {
			if _, visited := reachable[call.helper]; visited {
				continue
			}
			reachable[call.helper] = struct{}{}
			if _, existing := p.existingHelpers[call.helper]; !existing {
				operations = append(operations, p.operations[call.helper.index+1])
			}
		}
	}
	return reachable
}
