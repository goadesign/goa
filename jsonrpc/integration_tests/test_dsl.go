package main

import . "goa.design/goa/v3/dsl"

var _ = API("test", func() {
	Title("WebSocket Test API")
	Version("1.0")
})

var _ = Service("streaming", func() {
	JSONRPC(func() {
		Path("/")
	})
	
	Method("server_stream", func() {
		StreamingResult(String)
		
		JSONRPC(func() {
		})
	})
})