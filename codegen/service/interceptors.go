package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// ServiceInterceptorData contains all data needed for generating interceptor code
	ServiceInterceptorData struct {
		Service                       string
		PkgName                       string
		Methods                       []*MethodInterceptorData
		ServerInterceptors            []*InterceptorData
		ClientInterceptors            []*InterceptorData
		AllInterceptors               []*InterceptorData
		HasPrivateImplementationTypes bool
	}

	// MethodInterceptorData contains interceptor data for a single method
	MethodInterceptorData struct {
		Service            string
		Method             string
		MethodVarName      string
		PayloadRef         string
		ResultRef          string
		ServerInterceptors []*InterceptorData
		ClientInterceptors []*InterceptorData
	}

	// InterceptorData describes a single interceptor.
	InterceptorData struct {
		Name                    string
		DesignName              string
		UnexportedName          string
		Description             string
		PayloadRef              string
		ResultRef               string
		ReadPayload             []*AttributeData
		WritePayload            []*AttributeData
		ReadResult              []*AttributeData
		WriteResult             []*AttributeData
		ServerStreamInputStruct string
		ClientStreamInputStruct string
	}

	// AttributeData describes a single attribute.
	AttributeData struct {
		Name         string
		TypeRef      string
		FieldPointer bool
	}
)

// InterceptorsFile returns the interceptors file for the given service.
func InterceptorsFile(genpkg string, service *expr.ServiceExpr) *codegen.File {
	svc := Services.Get(service.Name)
	data := interceptorsData(service)
	if len(data.ServerInterceptors) == 0 && len(data.ClientInterceptors) == 0 {
		return nil
	}

	path := filepath.Join(codegen.Gendir, svc.PathName, "interceptors.go")
	sections := []*codegen.SectionTemplate{
		codegen.Header(service.Name+" interceptors", svc.PkgName, []*codegen.ImportSpec{
			{Path: "context"},
			codegen.GoaImport(""),
		}),
		{
			Name:   "interceptors",
			Source: readTemplate("interceptors"),
			Data:   data,
		},
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func interceptorsData(service *expr.ServiceExpr) *ServiceInterceptorData {
	svc := Services.Get(service.Name)
	scope := svc.Scope

	// Build method data first
	methods := make([]*MethodInterceptorData, 0, len(service.Methods))
	seenInts := make(map[string]*InterceptorData)
	var serviceServerInts, serviceClientInts, allInts []*InterceptorData
	var hasTypes bool

	for _, m := range service.Methods {
		methodServerInts, methodClientInts := buildMethodInterceptors(m, scope)
		if len(methodServerInts) == 0 && len(methodClientInts) == 0 {
			continue
		}
		hasTypes = hasTypes || hasPrivateImplementationTypes(methodServerInts) || hasPrivateImplementationTypes(methodClientInts)

		// Add method data
		methods = append(methods, &MethodInterceptorData{
			Service:            svc.Name,
			Method:             m.Name,
			MethodVarName:      codegen.Goify(m.Name, true),
			PayloadRef:         scope.GoFullTypeRef(m.Payload, ""),
			ResultRef:          scope.GoFullTypeRef(m.Result, ""),
			ServerInterceptors: methodServerInts,
			ClientInterceptors: methodClientInts,
		})

		// Collect unique interceptors
		for _, i := range methodServerInts {
			if _, ok := seenInts[i.Name]; !ok {
				seenInts[i.Name] = i
				serviceServerInts = append(serviceServerInts, i)
				allInts = append(allInts, i)
			}
		}
		for _, i := range methodClientInts {
			if _, ok := seenInts[i.Name]; !ok {
				seenInts[i.Name] = i
				serviceClientInts = append(serviceClientInts, i)
				allInts = append(allInts, i)
			}
		}
	}

	return &ServiceInterceptorData{
		Service:                       service.Name,
		PkgName:                       svc.PkgName,
		Methods:                       methods,
		ServerInterceptors:            serviceServerInts,
		ClientInterceptors:            serviceClientInts,
		AllInterceptors:               allInts,
		HasPrivateImplementationTypes: hasTypes,
	}
}

func buildMethodInterceptors(m *expr.MethodExpr, scope *codegen.NameScope) ([]*InterceptorData, []*InterceptorData) {
	svc := Services.Get(m.Service.Name)
	methodData := svc.Method(m.Name)
	var serverEndpointStruct, clientEndpointStruct string
	if methodData.ServerStream != nil {
		serverEndpointStruct = methodData.ServerStream.EndpointStruct
	}
	if methodData.ClientStream != nil {
		clientEndpointStruct = methodData.ClientStream.EndpointStruct
	}
	var hasPrivateImplementationTypes bool
	buildInterceptor := func(intr *expr.InterceptorExpr) *InterceptorData {
		hasPrivateImplementationTypes = hasPrivateImplementationTypes ||
			intr.ReadPayload != nil || intr.WritePayload != nil || intr.ReadResult != nil || intr.WriteResult != nil

		return &InterceptorData{
			Name:                    codegen.Goify(intr.Name, true),
			DesignName:              intr.Name,
			UnexportedName:          codegen.Goify(intr.Name, false),
			Description:             intr.Description,
			PayloadRef:              methodData.PayloadRef,
			ResultRef:               methodData.ResultRef,
			ServerStreamInputStruct: serverEndpointStruct,
			ClientStreamInputStruct: clientEndpointStruct,
			ReadPayload:             collectAttributes(intr.ReadPayload, m.Payload, scope),
			WritePayload:            collectAttributes(intr.WritePayload, m.Payload, scope),
			ReadResult:              collectAttributes(intr.ReadResult, m.Result, scope),
			WriteResult:             collectAttributes(intr.WriteResult, m.Result, scope),
		}
	}

	serverInts := make([]*InterceptorData, len(m.ServerInterceptors))
	for i, intr := range m.ServerInterceptors {
		serverInts[i] = buildInterceptor(intr)
	}

	clientInts := make([]*InterceptorData, len(m.ClientInterceptors))
	for i, intr := range m.ClientInterceptors {
		clientInts[i] = buildInterceptor(intr)
	}

	return serverInts, clientInts
}

// hasPrivateImplementationTypes returns true if any of the interceptors have
// private implementation types.
func hasPrivateImplementationTypes(interceptors []*InterceptorData) bool {
	for _, intr := range interceptors {
		if intr.ReadPayload != nil || intr.WritePayload != nil || intr.ReadResult != nil || intr.WriteResult != nil {
			return true
		}
	}
	return false
}

// collectAttributes builds AttributeData from an AttributeExpr
func collectAttributes(attrNames, parent *expr.AttributeExpr, scope *codegen.NameScope) []*AttributeData {
	if attrNames == nil {
		return nil
	}

	obj := expr.AsObject(attrNames.Type)
	if obj == nil {
		return nil
	}

	data := make([]*AttributeData, len(*obj))
	for i, nat := range *obj {
		parentAttr := parent.Find(nat.Name)
		if parentAttr == nil {
			continue
		}

		data[i] = &AttributeData{
			Name:         codegen.Goify(nat.Name, true),
			TypeRef:      scope.GoTypeRef(parentAttr),
			FieldPointer: parent.IsPrimitivePointer(nat.Name, true),
		}
	}
	return data
}
