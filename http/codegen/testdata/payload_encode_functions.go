package testdata

var PayloadQueryBoolEncodeCode = `// EncodeMethodQueryBoolRequest returns an encoder for requests sent to the
// ServiceQueryBool MethodQueryBool server.
func EncodeMethodQueryBoolRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicequerybool.MethodQueryBoolPayload)
		if !ok {
			return goahttp.ErrInvalidType("ServiceQueryBool", "MethodQueryBool", "*servicequerybool.MethodQueryBoolPayload", v)
		}
		values := req.URL.Query()
		if p.Q != nil {
			values.Add("q", *p.Q)
		}
		req.URL.RawQuery = values.Encode()
		return nil
	}
}
`

var PayloadQueryBoolValidateEncodeCode = `

`

var PayloadQueryIntEncodeCode = `

`

var PayloadQueryIntValidateEncodeCode = `

`

var PayloadQueryInt32EncodeCode = `

`

var PayloadQueryInt32ValidateEncodeCode = `

`

var PayloadQueryInt64EncodeCode = `

`

var PayloadQueryInt64ValidateEncodeCode = `

`

var PayloadQueryUIntEncodeCode = `

`

var PayloadQueryUIntValidateEncodeCode = `

`

var PayloadQueryUInt32EncodeCode = `

`

var PayloadQueryUInt32ValidateEncodeCode = `

`

var PayloadQueryUInt64EncodeCode = `

`

var PayloadQueryUInt64ValidateEncodeCode = `

`

var PayloadQueryFloat32EncodeCode = `

`

var PayloadQueryFloat32ValidateEncodeCode = `

`

var PayloadQueryFloat64EncodeCode = `

`

var PayloadQueryFloat64ValidateEncodeCode = `

`

var PayloadQueryStringEncodeCode = `

`

var PayloadQueryStringValidateEncodeCode = `

`

var PayloadQueryBytesEncodeCode = `

`

var PayloadQueryBytesValidateEncodeCode = `

`

var PayloadQueryAnyEncodeCode = `

`

var PayloadQueryAnyValidateEncodeCode = `

`

var PayloadQueryArrayBoolEncodeCode = `

`

var PayloadQueryArrayBoolValidateEncodeCode = `

`

var PayloadQueryArrayIntEncodeCode = `

`

var PayloadQueryArrayIntValidateEncodeCode = `

`

var PayloadQueryArrayInt32EncodeCode = `

`

var PayloadQueryArrayInt32ValidateEncodeCode = `

`

var PayloadQueryArrayInt64EncodeCode = `

`

var PayloadQueryArrayInt64ValidateEncodeCode = `

`

var PayloadQueryArrayUIntEncodeCode = `

`

var PayloadQueryArrayUIntValidateEncodeCode = `

`

var PayloadQueryArrayUInt32EncodeCode = `

`

var PayloadQueryArrayUInt32ValidateEncodeCode = `

`

var PayloadQueryArrayUInt64EncodeCode = `

`

var PayloadQueryArrayUInt64ValidateEncodeCode = `

`

var PayloadQueryArrayFloat32EncodeCode = `

`

var PayloadQueryArrayFloat32ValidateEncodeCode = `

`

var PayloadQueryArrayFloat64EncodeCode = `

`

var PayloadQueryArrayFloat64ValidateEncodeCode = `

`

var PayloadQueryArrayStringEncodeCode = `

`

var PayloadQueryArrayStringValidateEncodeCode = `

`

var PayloadQueryArrayBytesEncodeCode = `

`

var PayloadQueryArrayBytesValidateEncodeCode = `

`

var PayloadQueryArrayAnyEncodeCode = `

`

var PayloadQueryArrayAnyValidateEncodeCode = `

`

var PayloadQueryMapStringStringEncodeCode = `

`

var PayloadQueryMapStringStringValidateEncodeCode = `

`

var PayloadQueryMapStringBoolEncodeCode = `

`

var PayloadQueryMapStringBoolValidateEncodeCode = `

`

var PayloadQueryMapBoolStringEncodeCode = `

`

var PayloadQueryMapBoolStringValidateEncodeCode = `

`

var PayloadQueryMapBoolBoolEncodeCode = `

`

var PayloadQueryMapBoolBoolValidateEncodeCode = `

`

var PayloadQueryMapStringArrayStringEncodeCode = `

`

var PayloadQueryMapStringArrayStringValidateEncodeCode = `

`

var PayloadQueryMapStringArrayBoolEncodeCode = `

`

var PayloadQueryMapStringArrayBoolValidateEncodeCode = `

`

var PayloadQueryMapBoolArrayStringEncodeCode = `

`

var PayloadQueryMapBoolArrayStringValidateEncodeCode = `

`

var PayloadQueryMapBoolArrayBoolEncodeCode = `

`

var PayloadQueryMapBoolArrayBoolValidateEncodeCode = `

`

var PayloadQueryPrimitiveStringValidateEncodeCode = `

`

var PayloadQueryPrimitiveBoolValidateEncodeCode = `

`

var PayloadQueryPrimitiveArrayStringValidateEncodeCode = `

`

var PayloadQueryPrimitiveArrayBoolValidateEncodeCode = `

`

var PayloadQueryPrimitiveMapStringArrayStringValidateEncodeCode = `

`

var PayloadQueryPrimitiveMapStringBoolValidateEncodeCode = `

`

var PayloadQueryPrimitiveMapBoolArrayBoolValidateEncodeCode = `

`

var PayloadQueryStringMappedEncodeCode = `

`

var PayloadQueryStringDefaultEncodeCode = `

`

var PayloadQueryPrimitiveStringDefaultEncodeCode = `

`

var PayloadPathStringEncodeCode = `

`

var PayloadPathStringValidateEncodeCode = `

`

var PayloadPathArrayStringEncodeCode = `

`

var PayloadPathArrayStringValidateEncodeCode = `

`

var PayloadPathPrimitiveStringValidateEncodeCode = `

`

var PayloadPathPrimitiveBoolValidateEncodeCode = `

`

var PayloadPathPrimitiveArrayStringValidateEncodeCode = `

`

var PayloadPathPrimitiveArrayBoolValidateEncodeCode = `

`

var PayloadHeaderStringEncodeCode = `

`

var PayloadHeaderStringValidateEncodeCode = `

`

var PayloadHeaderArrayStringEncodeCode = `

`

var PayloadHeaderArrayStringValidateEncodeCode = `

`

var PayloadHeaderPrimitiveStringValidateEncodeCode = `

`

var PayloadHeaderPrimitiveBoolValidateEncodeCode = `

`

var PayloadHeaderPrimitiveArrayStringValidateEncodeCode = `

`

var PayloadHeaderPrimitiveArrayBoolValidateEncodeCode = `

`

var PayloadHeaderStringDefaultEncodeCode = `

`

var PayloadHeaderPrimitiveStringDefaultEncodeCode = `

`

var PayloadBodyStringEncodeCode = `

`

var PayloadBodyStringValidateEncodeCode = `
                // EncodeMethodBodyStringValidateRequest returns an encoder for requests sent                                                                             
                // to the ServiceBodyStringValidate MethodBodyStringValidate server.                                                                                      
                func EncodeMethodBodyStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                          
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodystringvalidate.MethodBodyStringValidatePayload)                                                                   
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyStringValidate", "MethodBodyStringValidate", "*servicebodystringvalidate.MethodBodyStrin
gValidatePayload", v)                                                                                                                                                     
                                }                                                                                                                                         
                                body := NewMethodBodyStringValidateRequestBody(p)                                                                                         
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyStringValidate", "MethodBodyStringValidate", err)                                     
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyObjectEncodeCode = `

`

var PayloadObjectBodyValidateEncodeCode = `

`

var PayloadBodyUserEncodeCode = `
                // EncodeMethodBodyUserRequest returns an encoder for requests sent to the                                                                                
                // ServiceBodyUser MethodBodyUser server.                                                                                                                 
                func EncodeMethodBodyUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                                    
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyuser.PayloadType)                                                                                                 
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyUser", "MethodBodyUser", "*servicebodyuser.PayloadType", v)                             
                                }                                                                                                                                         
                                body := NewMethodBodyUserRequestBody(p)                                                                                                   
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyUser", "MethodBodyUser", err)                                                         
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadUserBodyValidateEncodeCode = `
                // EncodeMethodBodyUserValidateRequest returns an encoder for requests sent to                                                                            
                // the ServiceBodyUserValidate MethodBodyUserValidate server.                                                                                             
                func EncodeMethodBodyUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                            
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyuservalidate.PayloadType)                                                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyUserValidate", "MethodBodyUserValidate", "*servicebodyuservalidate.PayloadType", v)     
                                }                                                                                                                                         
                                body := NewMethodBodyUserValidateRequestBody(p)                                                                                           
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyUserValidate", "MethodBodyUserValidate", err)                                         
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyArrayStringEncodeCode = `
                // EncodeMethodBodyArrayStringRequest returns an encoder for requests sent to                                                                             
                // the ServiceBodyArrayString MethodBodyArrayString server.                                                                                               
                func EncodeMethodBodyArrayStringRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                             
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyarraystring.MethodBodyArrayStringPayload)                                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyArrayString", "MethodBodyArrayString", "*servicebodyarraystring.MethodBodyArrayStringPay
load", v)                                                                                                                                                                 
                                }                                                                                                                                         
                                body := NewMethodBodyArrayStringRequestBody(p)                                                                                            
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyArrayString", "MethodBodyArrayString", err)                                           
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyArrayStringValidateEncodeCode = `
               // EncodeMethodBodyArrayStringValidateRequest returns an encoder for requests                                                                             
                // sent to the ServiceBodyArrayStringValidate MethodBodyArrayStringValidate                                                                               
                // server.                                                                                                                                                
                func EncodeMethodBodyArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                     
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyarraystringvalidate.MethodBodyArrayStringValidatePayload)                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyArrayStringValidate", "MethodBodyArrayStringValidate", "*servicebodyarraystringvalidate.
MethodBodyArrayStringValidatePayload", v)                                                                                                                                 
                                }                                                                                                                                         
                                body := NewMethodBodyArrayStringValidateRequestBody(p)                                                                                    
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyArrayStringValidate", "MethodBodyArrayStringValidate", err)                           
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyArrayUserEncodeCode = `
                // EncodeMethodBodyArrayUserRequest returns an encoder for requests sent to the                                                                           
                // ServiceBodyArrayUser MethodBodyArrayUser server.                                                                                                       
                func EncodeMethodBodyArrayUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                               
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyarrayuser.MethodBodyArrayUserPayload)                                                                             
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyArrayUser", "MethodBodyArrayUser", "*servicebodyarrayuser.MethodBodyArrayUserPayload", v
)                                                                                                                                                                         
                                }                                                                                                                                         
                                body := NewMethodBodyArrayUserRequestBody(p)                                                                                              
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyArrayUser", "MethodBodyArrayUser", err)                                               
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyArrayUserValidateEncodeCode = `
                // EncodeMethodBodyArrayUserValidateRequest returns an encoder for requests                                                                               
                // sent to the ServiceBodyArrayUserValidate MethodBodyArrayUserValidate server.                                                                           
                func EncodeMethodBodyArrayUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                       
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyarrayuservalidate.MethodBodyArrayUserValidatePayload)                                                             
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyArrayUserValidate", "MethodBodyArrayUserValidate", "*servicebodyarrayuservalidate.Method
BodyArrayUserValidatePayload", v)                                                                                                                                         
                                }                                                                                                                                         
                                body := NewMethodBodyArrayUserValidateRequestBody(p)                                                                                      
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyArrayUserValidate", "MethodBodyArrayUserValidate", err)                               
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyMapStringEncodeCode = `
                // EncodeMethodBodyMapStringRequest returns an encoder for requests sent to the                                                                           
                // ServiceBodyMapString MethodBodyMapString server.                                                                                                       
                func EncodeMethodBodyMapStringRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                               
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodymapstring.MethodBodyMapStringPayload)                                                                             
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyMapString", "MethodBodyMapString", "*servicebodymapstring.MethodBodyMapStringPayload", v
)                                                                                                                                                                         
                                }                                                                                                                                         
                                body := NewMethodBodyMapStringRequestBody(p)                                                                                              
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyMapString", "MethodBodyMapString", err)                                               
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyMapStringValidateEncodeCode = `
                // EncodeMethodBodyMapStringValidateRequest returns an encoder for requests                                                                               
                // sent to the ServiceBodyMapStringValidate MethodBodyMapStringValidate server.                                                                           
                func EncodeMethodBodyMapStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                       
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodymapstringvalidate.MethodBodyMapStringValidatePayload)                                                             
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyMapStringValidate", "MethodBodyMapStringValidate", "*servicebodymapstringvalidate.Method
BodyMapStringValidatePayload", v)                                                                                                                                         
                                }                                                                                                                                         
                                body := NewMethodBodyMapStringValidateRequestBody(p)                                                                                      
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyMapStringValidate", "MethodBodyMapStringValidate", err)                               
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyMapUserEncodeCode = `
                // EncodeMethodBodyMapUserRequest returns an encoder for requests sent to the                                                                             
                // ServiceBodyMapUser MethodBodyMapUser server.                                                                                                           
                func EncodeMethodBodyMapUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                                 
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodymapuser.MethodBodyMapUserPayload)                                                                                 
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyMapUser", "MethodBodyMapUser", "*servicebodymapuser.MethodBodyMapUserPayload", v)       
                                }                                                                                                                                         
                                body := NewMethodBodyMapUserRequestBody(p)                                                                                                
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyMapUser", "MethodBodyMapUser", err)                                                   
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyMapUserValidateEncodeCode = `
                // EncodeMethodBodyMapUserValidateRequest returns an encoder for requests sent                                                                            
                // to the ServiceBodyMapUserValidate MethodBodyMapUserValidate server.                                                                                    
                func EncodeMethodBodyMapUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                         
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodymapuservalidate.MethodBodyMapUserValidatePayload)                                                                 
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyMapUserValidate", "MethodBodyMapUserValidate", "*servicebodymapuservalidate.MethodBodyMa
pUserValidatePayload", v)                                                                                                                                                 
                                }                                                                                                                                         
                                body := NewMethodBodyMapUserValidateRequestBody(p)                                                                                        
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyMapUserValidate", "MethodBodyMapUserValidate", err)                                   
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPrimitiveStringValidateEncodeCode = `
                // EncodeMethodBodyPrimitiveStringValidateRequest returns an encoder for                                                                                  
                // requests sent to the ServiceBodyPrimitiveStringValidate                                                                                                
                // MethodBodyPrimitiveStringValidate server.                                                                                                              
                func EncodeMethodBodyPrimitiveStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                 
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(string)                                                                                                                       
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveStringValidate", "MethodBodyPrimitiveStringValidate", "string", v)             
                                }                                                                                                                                         
                                body := p                                                                                                                                 
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveStringValidate", "MethodBodyPrimitiveStringValidate", err)                   
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPrimitiveBoolValidateEncodeCode = `
                // EncodeMethodBodyPrimitiveBoolValidateRequest returns an encoder for requests                                                                           
                // sent to the ServiceBodyPrimitiveBoolValidate MethodBodyPrimitiveBoolValidate                                                                           
                // server.                                                                                                                                                
                func EncodeMethodBodyPrimitiveBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                   
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(bool)                                                                                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveBoolValidate", "MethodBodyPrimitiveBoolValidate", "bool", v)                   
                                }                                                                                                                                         
                                body := p                                                                                                                                 
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveBoolValidate", "MethodBodyPrimitiveBoolValidate", err)                       
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                } 
`

var PayloadBodyPrimitiveArrayStringValidateEncodeCode = `
                // EncodeMethodBodyPrimitiveArrayStringValidateRequest returns an encoder for                                                                             
                // requests sent to the ServiceBodyPrimitiveArrayStringValidate                                                                                           
                // MethodBodyPrimitiveArrayStringValidate server.                                                                                                         
                func EncodeMethodBodyPrimitiveArrayStringValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {            
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.([]string)                                                                                                                     
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayStringValidate", "MethodBodyPrimitiveArrayStringValidate", "[]string", v) 
                                }                                                                                                                                         
                                body := p                                                                                                                                 
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayStringValidate", "MethodBodyPrimitiveArrayStringValidate", err)         
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPrimitiveArrayBoolValidateEncodeCode = `
                // EncodeMethodBodyPrimitiveArrayBoolValidateRequest returns an encoder for                                                                               
                // requests sent to the ServiceBodyPrimitiveArrayBoolValidate                                                                                             
                // MethodBodyPrimitiveArrayBoolValidate server.                                                                                                           
                func EncodeMethodBodyPrimitiveArrayBoolValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {              
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.([]bool)                                                                                                                       
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayBoolValidate", "MethodBodyPrimitiveArrayBoolValidate", "[]bool", v)       
                                }                                                                                                                                         
                                body := p                                                                                                                                 
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayBoolValidate", "MethodBodyPrimitiveArrayBoolValidate", err)             
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPrimitiveArrayUserValidateEncodeCode = `
                // EncodeMethodBodyPrimitiveArrayUserValidateRequest returns an encoder for                                                                               
                // requests sent to the ServiceBodyPrimitiveArrayUserValidate                                                                                             
                // MethodBodyPrimitiveArrayUserValidate server.                                                                                                           
                func EncodeMethodBodyPrimitiveArrayUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {              
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.([]*servicebodyprimitivearrayuservalidate.PayloadType)                                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", "[]*servicebodyprim
itivearrayuservalidate.PayloadType", v)                                                                                                                                   
                                }                                                                                                                                         
                                body := NewPayloadTypeRequestBody(p)                                                                                                      
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", err)             
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPrimitiveFieldArrayUserEncodeCode = `
                // EncodeMethodBodyPrimitiveArrayUserRequest returns an encoder for requests                                                                              
                // sent to the ServiceBodyPrimitiveArrayUser MethodBodyPrimitiveArrayUser                                                                                 
                // server.                                                                                                                                                
                func EncodeMethodBodyPrimitiveArrayUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                      
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyprimitivearrayuser.PayloadType)                                                                                   
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayUser", "MethodBodyPrimitiveArrayUser", "*servicebodyprimitivearrayuser.Pay
loadType", v)                                                                                                                                                             
                                }                                                                                                                                         
                                body := p                                                                                                                                 
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayUser", "MethodBodyPrimitiveArrayUser", err)                             
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPrimitiveFieldArrayUserValidateEncodeCode = `
                // EncodeMethodBodyPrimitiveArrayUserValidateRequest returns an encoder for                                                                               
                // requests sent to the ServiceBodyPrimitiveArrayUserValidate                                                                                             
                // MethodBodyPrimitiveArrayUserValidate server.                                                                                                           
                func EncodeMethodBodyPrimitiveArrayUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {              
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyprimitivearrayuservalidate.PayloadType)                                                                           
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", "*servicebodyprimit
ivearrayuservalidate.PayloadType", v)                                                                                                                                     
                                }                                                                                                                                         
                                body := p                                                                                                                                 
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPrimitiveArrayUserValidate", "MethodBodyPrimitiveArrayUserValidate", err)             
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryObjectEncodeCode = `
               // EncodeMethodBodyQueryObjectRequest returns an encoder for requests sent to                                                                             
                // the ServiceBodyQueryObject MethodBodyQueryObject server.                                                                                               
                func EncodeMethodBodyQueryObjectRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                             
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyqueryobject.MethodBodyQueryObjectPayload)                                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyQueryObject", "MethodBodyQueryObject", "*servicebodyqueryobject.MethodBodyQueryObjectPay
load", v)                                                                                                                                                                 
                                }                                                                                                                                         
                                values := req.URL.Query()                                                                                                                 
                                if p.B != nil {                                                                                                                           
                                        values.Add("b", *p.B)                                                                                                             
                                }                                                                                                                                         
                                req.URL.RawQuery = values.Encode()                                                                                                        
                                body := NewMethodBodyQueryObjectRequestBody(p)                                                                                            
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyQueryObject", "MethodBodyQueryObject", err)                                           
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryObjectValidateEncodeCode = `
                // EncodeMethodBodyQueryObjectValidateRequest returns an encoder for requests                                                                             
                // sent to the ServiceBodyQueryObjectValidate MethodBodyQueryObjectValidate                                                                               
                // server.                                                                                                                                                
                func EncodeMethodBodyQueryObjectValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                     
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload)                                                         
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyQueryObjectValidate", "MethodBodyQueryObjectValidate", "*servicebodyqueryobjectvalidate.
MethodBodyQueryObjectValidatePayload", v)                                                                                                                                 
                                }                                                                                                                                         
                                values := req.URL.Query()                                                                                                                 
                                values.Add("b", p.B)                                                                                                                      
                                req.URL.RawQuery = values.Encode()                                                                                                        
                                body := NewMethodBodyQueryObjectValidateRequestBody(p)                                                                                    
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyQueryObjectValidate", "MethodBodyQueryObjectValidate", err)                           
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryUserEncodeCode = `
               // EncodeMethodBodyQueryUserRequest returns an encoder for requests sent to the                                                                           
                // ServiceBodyQueryUser MethodBodyQueryUser server.                                                                                                       
                func EncodeMethodBodyQueryUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                               
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyqueryuser.PayloadType)                                                                                            
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyQueryUser", "MethodBodyQueryUser", "*servicebodyqueryuser.PayloadType", v)              
                                }                                                                                                                                         
                                values := req.URL.Query()                                                                                                                 
                                if p.B != nil {                                                                                                                           
                                        values.Add("b", *p.B)                                                                                                             
                                }                                                                                                                                         
                                req.URL.RawQuery = values.Encode()                                                                                                        
                                body := NewMethodBodyQueryUserRequestBody(p)                                                                                              
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyQueryUser", "MethodBodyQueryUser", err)                                               
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryUserValidateEncodeCode = `
                // EncodeMethodBodyQueryUserValidateRequest returns an encoder for requests                                                                               
                // sent to the ServiceBodyQueryUserValidate MethodBodyQueryUserValidate server.                                                                           
                func EncodeMethodBodyQueryUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                       
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyqueryuservalidate.PayloadType)                                                                                    
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyQueryUserValidate", "MethodBodyQueryUserValidate", "*servicebodyqueryuservalidate.Payloa
dType", v)                                                                                                                                                                
                                }                                                                                                                                         
                                values := req.URL.Query()                                                                                                                 
                                values.Add("b", p.B)                                                                                                                      
                                req.URL.RawQuery = values.Encode()                                                                                                        
                                body := NewMethodBodyQueryUserValidateRequestBody(p)                                                                                      
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyQueryUserValidate", "MethodBodyQueryUserValidate", err)                               
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPathObjectEncodeCode = `
                // EncodeMethodBodyPathObjectRequest returns an encoder for requests sent to                                                                              
                // the ServiceBodyPathObject MethodBodyPathObject server.                                                                                                 
                func EncodeMethodBodyPathObjectRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                              
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodypathobject.MethodBodyPathObjectPayload)                                                                           
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPathObject", "MethodBodyPathObject", "*servicebodypathobject.MethodBodyPathObjectPayload
", v)                                                                                                                                                                     
                                }                                                                                                                                         
                                body := NewMethodBodyPathObjectRequestBody(p)                                                                                             
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPathObject", "MethodBodyPathObject", err)                                             
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPathObjectValidateEncodeCode = `
               // EncodeMethodBodyPathObjectValidateRequest returns an encoder for requests                                                                              
                // sent to the ServiceBodyPathObjectValidate MethodBodyPathObjectValidate                                                                                 
                // server.                                                                                                                                                
                func EncodeMethodBodyPathObjectValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                      
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload)                                                           
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPathObjectValidate", "MethodBodyPathObjectValidate", "*servicebodypathobjectvalidate.Met
hodBodyPathObjectValidatePayload", v)                                                                                                                                     
                                }                                                                                                                                         
                                body := NewMethodBodyPathObjectValidateRequestBody(p)                                                                                     
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPathObjectValidate", "MethodBodyPathObjectValidate", err)                             
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                } 
`

var PayloadBodyPathUserEncodeCode = `
                // EncodeMethodBodyPathUserRequest returns an encoder for requests sent to the                                                                            
                // ServiceBodyPathUser MethodBodyPathUser server.                                                                                                         
                func EncodeMethodBodyPathUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                                
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodypathuser.PayloadType)                                                                                             
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPathUser", "MethodBodyPathUser", "*servicebodypathuser.PayloadType", v)                 
                                }                                                                                                                                         
                                body := NewMethodBodyPathUserRequestBody(p)                                                                                               
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPathUser", "MethodBodyPathUser", err)                                                 
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyPathUserValidateEncodeCode = `
               // EncodeMethodUserBodyPathValidateRequest returns an encoder for requests sent                                                                           
                // to the ServiceBodyPathUserValidate MethodUserBodyPathValidate server.                                                                                  
                func EncodeMethodUserBodyPathValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                        
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodypathuservalidate.PayloadType)                                                                                     
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyPathUserValidate", "MethodUserBodyPathValidate", "*servicebodypathuservalidate.PayloadTy
pe", v)                                                                                                                                                                   
                                }                                                                                                                                         
                                body := NewMethodUserBodyPathValidateRequestBody(p)                                                                                       
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyPathUserValidate", "MethodUserBodyPathValidate", err)                                 
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryPathObjectEncodeCode = `
                // EncodeMethodBodyQueryPathObjectRequest returns an encoder for requests sent                                                                            
                // to the ServiceBodyQueryPathObject MethodBodyQueryPathObject server.                                                                                    
                func EncodeMethodBodyQueryPathObjectRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                         
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyquerypathobject.MethodBodyQueryPathObjectPayload)                                                                 
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyQueryPathObject", "MethodBodyQueryPathObject", "*servicebodyquerypathobject.MethodBodyQu
eryPathObjectPayload", v)                                                                                                                                                 
                                }                                                                                                                                         
                                values := req.URL.Query()                                                                                                                 
                                if p.B != nil {                                                                                                                           
                                        values.Add("b", *p.B)                                                                                                             
                                }                                                                                                                                         
                                req.URL.RawQuery = values.Encode()                                                                                                        
                                body := NewMethodBodyQueryPathObjectRequestBody(p)                                                                                        
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyQueryPathObject", "MethodBodyQueryPathObject", err)                                   
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryPathObjectValidateEncodeCode = `
                // EncodeMethodBodyQueryPathObjectValidateRequest returns an encoder for                                                                                  
                // requests sent to the ServiceBodyQueryPathObjectValidate                                                                                                
                // MethodBodyQueryPathObjectValidate server.                                                                                                              
                func EncodeMethodBodyQueryPathObjectValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {                 
                        return func(req *http.Request, v interface{}) error {                                                                                             
                                p, ok := v.(*servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload)                                                 
                                if !ok {                                                                                                                                  
                                        return goahttp.ErrInvalidType("ServiceBodyQueryPathObjectValidate", "MethodBodyQueryPathObjectValidate", "*servicebodyquerypathobj
ectvalidate.MethodBodyQueryPathObjectValidatePayload", v)                                                                                                                 
                                }                                                                                                                                         
                                values := req.URL.Query()                                                                                                                 
                                values.Add("b", p.B)                                                                                                                      
                                req.URL.RawQuery = values.Encode()                                                                                                        
                                body := NewMethodBodyQueryPathObjectValidateRequestBody(p)                                                                                
                                if err := encoder(req).Encode(&body); err != nil {                                                                                        
                                        return goahttp.ErrEncodingError("ServiceBodyQueryPathObjectValidate", "MethodBodyQueryPathObjectValidate", err)                   
                                }                                                                                                                                         
                                return nil                                                                                                                                
                        }                                                                                                                                                 
                }
`

var PayloadBodyQueryPathUserEncodeCode = `
// EncodeMethodBodyQueryPathUserRequest returns an encoder for requests sent to
// the ServiceBodyQueryPathUser MethodBodyQueryPathUser server.
func EncodeMethodBodyQueryPathUserRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyquerypathuser.PayloadType)
		if !ok {
			return goahttp.ErrInvalidType("ServiceBodyQueryPathUser", "MethodBodyQueryPathUser", "*servicebodyquerypathuser.PayloadType", v)
		}
		values := req.URL.Query()
		if p.B != nil {
			values.Add("b", *p.B)
		}
		req.URL.RawQuery = values.Encode()
		body := NewMethodBodyQueryPathUserRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("ServiceBodyQueryPathUser", "MethodBodyQueryPathUser", err)
		}
		return nil
	}
}
`

var PayloadBodyQueryPathUserValidateEncodeCode = `// EncodeMethodBodyQueryPathUserValidateRequest returns an encoder for requests
// sent to the ServiceBodyQueryPathUserValidate MethodBodyQueryPathUserValidate
// server.
func EncodeMethodBodyQueryPathUserValidateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*servicebodyquerypathuservalidate.PayloadType)
		if !ok {
				return goahttp.ErrInvalidType("ServiceBodyQueryPathUserValidate", "MethodBodyQueryPathUserValidate", "*servicebodyquerypathuservalidate.PayloadType", v)
		}
		values := req.URL.Query()
		values.Add("b", p.B)
		req.URL.RawQuery = values.Encode()
		body := NewMethodBodyQueryPathUserValidateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
				return goahttp.ErrEncodingError("ServiceBodyQueryPathUserValidate", "MethodBodyQueryPathUserValidate", err)
		}
		return nil
	}
}
`
