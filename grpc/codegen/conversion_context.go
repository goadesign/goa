// This file binds gRPC conversions to their planned service Go types. Nested
// helpers keep the package inherited from their enclosing service value, while
// the protobuf resolver supplies protoc's field names and native Go types.
package codegen

import "goa.design/goa/v3/codegen"

// bindContexts attaches the saved service layout to the source when converting
// to protobuf, or to the target when converting from protobuf. Repeated uses
// share that binding and the planned helper names.
func (c *grpcConversion) bindContexts(source, target *codegen.AttributeContext, proto bool) error {
	if c.bound {
		return nil
	}
	serviceContext := target
	if proto {
		serviceContext = source
	}
	serviceContext, err := serviceContext.WithGoTypeLayout(c.serviceLayout.Link(c.pkg.ImportPath(), c.pkg.ImportName))
	if err != nil {
		return err
	}
	if proto {
		source = serviceContext
	} else {
		target = serviceContext
	}
	if err := c.transform.BindContexts(source, target); err != nil {
		return err
	}
	c.bound = true
	return nil
}
