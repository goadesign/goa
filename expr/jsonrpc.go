package expr

type (
	// JSONRPCExpr contains the API level JSON-RPC specific expressions.
	JSONRPCExpr struct {
		HTTPExpr
	}
)

// EvalName returns the name printed in case of evaluation error.
func (*JSONRPCExpr) EvalName() string {
	return "API JSON-RPC"
}

// Prepare copies shared HTTP request settings to the JSON-RPC API. JSON-RPC
// error mappings stay separate because their codes belong to JSON-RPC, not
// HTTP.
func (j *JSONRPCExpr) Prepare() {
	j.Path = Root.API.HTTP.Path
	j.Params = Root.API.HTTP.Params
	j.Headers = Root.API.HTTP.Headers
	j.Cookies = Root.API.HTTP.Cookies
	j.SSE = Root.API.HTTP.SSE
}
