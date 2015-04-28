package design

var (
	apiDSLSource *apiDSL // API DSL
)

// apiDSL contains the non-evaluated API DSL functions
type apiDSL struct {
	name string   // Name of API
	dsls []func() // DSL funcs
}

// API creates or amends the API DSL
func API(name string, dsl func()) {
	if apiDSL != nil {
		if apiDSL.name != name {
			fatalf("API defined with conflicting names '%s' and '%s'",
				apiDSL.name, name)
		}
	} else {
		apiDSL = &apiDSL{name: name}
	}
	apiDSL.dsls = append(apiDSL.dsls, dsl)
}

// Run DSL to produce API definition
func (d *apiDSL) execute() (*APIDefinition, error) {
	def := APIDefinition{Name: d.name}
	ctxStack = append(ctxStack, def)
	for _, dsl := range d.dsls {
		dsl()
		if dslError != nil {
			return nil, dslError
		}
	}
	return def, nil
}


