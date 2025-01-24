package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Interceptor defines a request interceptor. Interceptors provide a type-safe way
// to read and write from and to the request and response.
//
// Interceptor must appear in an API, Service or Method expression.
//
// Interceptor accepts two arguments: the name of the interceptor and the
// defining DSL.
//
// Example:
//
//	var Cache = Interceptor("Cache", func() {
//	   Description("Server-side interceptor which implements a transparent cache for the loaded records")
//
//	   ReadPayload(func() {
//	      Attribute("id")
//	   })
//
//	   WriteResult(func() {
//	      Attribute("cachedAt")
//	   })
//	})
func Interceptor(name string, fn ...func()) *expr.InterceptorExpr {
	if len(fn) > 1 {
		eval.ReportError("interceptor %q cannot have multiple definitions", name)
		return nil
	}
	i := &expr.InterceptorExpr{Name: name}
	if name == "" {
		eval.ReportError("interceptor name cannot be empty")
		return i
	}
	if len(fn) > 0 {
		if !eval.Execute(fn[0], i) {
			return i
		}
	}
	for _, i := range expr.Root.Interceptors {
		if i.Name == name {
			eval.ReportError("interceptor %q already defined", name)
			return i
		}
	}
	expr.Root.Interceptors = append(expr.Root.Interceptors, i)
	return i
}

// ReadPayload defines the payload attributes read by the interceptor.
//
// ReadPayload must appear in an interceptor DSL.
//
// ReadPayload takes a function as argument which can use the Attribute DSL to
// define the attributes read by the interceptor.
//
// Example:
//
//	ReadPayload(func() {
//	   Attribute("id")
//	})
//
// ReadPayload also accepts user defined types:
//
//	// Interceptor can read any payload field
//	ReadPayload(MethodPayload)
func ReadPayload(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.ReadPayload = attr
	})
}

// WritePayload defines the payload attributes written by the interceptor.
//
// WritePayload must appear in an interceptor DSL.
//
// WritePayload takes a function as argument which can use the Attribute DSL to
// define the attributes written by the interceptor.
//
// Example:
//
//	WritePayload(func() {
//	   Attribute("auth")
//	})
//
// WritePayload also accepts user defined types:
//
//	// Interceptor can write any payload field
//	WritePayload(MethodPayload)
func WritePayload(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.WritePayload = attr
	})
}

// ReadResult defines the result attributes read by the interceptor.
//
// ReadResult must appear in an interceptor DSL.
//
// ReadResult takes a function as argument which can use the Attribute DSL to
// define the attributes read by the interceptor.
//
// Example:
//
//	ReadResult(func() {
//	   Attribute("cachedAt")
//	})
//
// ReadResult also accepts user defined types:
//
//	// Interceptor can read any result field
//	ReadResult(MethodResult)
func ReadResult(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.ReadResult = attr
	})
}

// WriteResult defines the result attributes written by the interceptor.
//
// WriteResult must appear in an interceptor DSL.
//
// WriteResult takes a function as argument which can use the Attribute DSL to
// define the attributes written by the interceptor.
//
// Example:
//
//	WriteResult(func() {
//	   Attribute("cachedAt")
//	})
//
// WriteResult also accepts user defined types:
//
//	// Interceptor can write any result field
//	WriteResult(MethodResult)
func WriteResult(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.WriteResult = attr
	})
}

// ReadStreamingPayload defines the streaming payload attributes read by the interceptor.
//
// ReadStreamingPayload must appear in an interceptor DSL.
//
// ReadStreamingPayload takes a function as argument which can use the Attribute DSL to
// define the attributes read by the interceptor.
//
// Example:
//
//	ReadStreamingPayload(func() {
//	   Attribute("id")
//	})
//
// ReadStreamingPayload also accepts user defined types:
//
//	// Interceptor can read any streaming payload field
//	ReadStreamingPayload(MethodStreamingPayload)
func ReadStreamingPayload(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.ReadStreamingPayload = attr
	})
}

// WriteStreamingPayload defines the streaming payload attributes written by the interceptor.
//
// WriteStreamingPayload must appear in an interceptor DSL.
//
// WriteStreamingPayload takes a function as argument which can use the Attribute DSL to
// define the attributes written by the interceptor.
//
// Example:
//
//	WriteStreamingPayload(func() {
//	   Attribute("id")
//	})
//
// WriteStreamingPayload also accepts user defined types:
//
//	// Interceptor can write any streaming payload field
//	WriteStreamingPayload(MethodStreamingPayload)
func WriteStreamingPayload(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.WriteStreamingPayload = attr
	})
}

// ReadStreamingResult defines the streaming result attributes read by the interceptor.
//
// ReadStreamingResult must appear in an interceptor DSL.
//
// ReadStreamingResult takes a function as argument which can use the Attribute DSL to
// define the attributes read by the interceptor.
//
// Example:
//
//	ReadStreamingResult(func() {
//	   Attribute("cachedAt")
//	})
//
// ReadStreamingResult also accepts user defined types:
//
//	// Interceptor can read any streaming result field
//	ReadStreamingResult(MethodStreamingResult)
func ReadStreamingResult(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.ReadStreamingResult = attr
	})
}

// WriteStreamingResult defines the streaming result attributes written by the interceptor.
//
// WriteStreamingResult must appear in an interceptor DSL.
//
// WriteStreamingResult takes a function as argument which can use the Attribute DSL to
// define the attributes written by the interceptor.
//
// Example:
//
//	WriteStreamingResult(func() {
//	   Attribute("cachedAt")
//	})
//
// WriteStreamingResult also accepts user defined types:
//
//	// Interceptor can write any streaming result field
//	WriteStreamingResult(MethodStreamingResult)
func WriteStreamingResult(arg any) {
	setInterceptorAttribute(arg, func(i *expr.InterceptorExpr, attr *expr.AttributeExpr) {
		i.WriteStreamingResult = attr
	})
}

// ServerInterceptor lists the server-side interceptors that apply to all the
// API endpoints, all the service endpoints or a specific endpoint.
//
// ServerInterceptor must appear in an API, Service or Method expression.
//
// ServerInterceptor accepts one or more interceptor or interceptor names as
// arguments. ServerInterceptor can appear multiple times in the same DSL.
//
// Example:
//
//	Method("get_record", func() {
//	   // Interceptor defined with the Interceptor DSL
//	   ServerInterceptor(SetDeadline)
//
//	   // Name of interceptor defined with the Interceptor DSL
//	   ServerInterceptor("Cache")
//
//	   // Interceptor defined inline
//	   ServerInterceptor(Interceptor("CheckUserID", func() {
//	      ReadPayload(func() {
//	         Attribute("auth")
//	      })
//	   }))
//
//	   // ... rest of the method DSL
//	})
func ServerInterceptor(interceptors ...any) {
	addInterceptors(interceptors, false)
}

// ClientInterceptor lists the client-side interceptors that apply to all the
// API endpoints, all the service endpoints or a specific endpoint.
//
// ClientInterceptor must appear in an API, Service or Method expression.
//
// ClientInterceptor accepts one or more interceptor or interceptor names as
// arguments. ClientInterceptor can appear multiple times in the same DSL.
//
// Example:
//
//	Method("get_record", func() {
//	   // Interceptor defined with the Interceptor DSL
//	   ClientInterceptor(Retry)
//
//	   // Name of interceptor defined with the Interceptor DSL
//	   ClientInterceptor("Cache")
//
//	   // Interceptor defined inline
//	   ClientInterceptor(Interceptor("Sign", func() {
//	      ReadPayload(func() {
//	         Attribute("user_id")
//	      })
//	      WritePayload(func() {
//	         Attribute("auth")
//	      })
//	   }))
//
//	   // ... rest of the method DSL
//	})
func ClientInterceptor(interceptors ...any) {
	addInterceptors(interceptors, true)
}

// setInterceptorAttribute is a helper function that handles the common logic for
// setting interceptor attributes (ReadPayload, WritePayload, ReadResult, WriteResult, ReadStreamingPayload, WriteStreamingPayload, ReadStreamingResult, WriteStreamingResult).
func setInterceptorAttribute(arg any, setter func(i *expr.InterceptorExpr, attr *expr.AttributeExpr)) {
	i, ok := eval.Current().(*expr.InterceptorExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}

	var attr *expr.AttributeExpr
	switch fn := arg.(type) {
	case func():
		attr = &expr.AttributeExpr{Type: &expr.Object{}}
		if !eval.Execute(fn, attr) {
			return
		}
	case *expr.AttributeExpr:
		attr = fn
	case expr.DataType:
		attr = &expr.AttributeExpr{Type: fn}
	default:
		eval.InvalidArgError("type, attribute or func()", arg)
		return
	}
	setter(i, attr)
}

// addInterceptors is a helper function that validates and adds interceptors to
// the current expression.
func addInterceptors(interceptors []any, client bool) {
	kind := "ServerInterceptor"
	if client {
		kind = "ClientInterceptor"
	}
	if len(interceptors) == 0 {
		eval.ReportError("%s: at least one interceptor must be specified", kind)
		return
	}

	var ints []*expr.InterceptorExpr
	for _, i := range interceptors {
		switch i := i.(type) {
		case *expr.InterceptorExpr:
			ints = append(ints, i)
		case string:
			if i == "" {
				eval.ReportError("%s: interceptor name cannot be empty", kind)
				return
			}
			var found bool
			for _, in := range expr.Root.Interceptors {
				if in.Name == i {
					ints = append(ints, in)
					found = true
					break
				}
			}
			if !found {
				eval.ReportError("%s: interceptor %q not found", kind, i)
			}
		default:
			eval.ReportError("%s: invalid interceptor %v", kind, i)
		}
	}

	current := eval.Current()
	switch actual := current.(type) {
	case *expr.APIExpr:
		if client {
			actual.ClientInterceptors = append(actual.ClientInterceptors, ints...)
		} else {
			actual.ServerInterceptors = append(actual.ServerInterceptors, ints...)
		}
	case *expr.ServiceExpr:
		if client {
			actual.ClientInterceptors = append(actual.ClientInterceptors, ints...)
		} else {
			actual.ServerInterceptors = append(actual.ServerInterceptors, ints...)
		}
	case *expr.MethodExpr:
		if client {
			actual.ClientInterceptors = append(actual.ClientInterceptors, ints...)
		} else {
			actual.ServerInterceptors = append(actual.ServerInterceptors, ints...)
		}
	default:
		eval.IncompatibleDSL()
	}
}
