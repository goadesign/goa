// These tests render native JSON-RPC clients and servers with explicit HTTP
// query inputs. The query carries transport values, while JSON-RPC parameters
// retain only the authored body. No MCP plugin participates in this check.
package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratedJSONRPCQueryInputs(t *testing.T) {
	dir := renderParamsRuntimeModule(t)
	source := filepath.Join(dir, "jsonrpc", "param_shapes", "client", "query_runtime_test.go")
	require.NoError(t, os.WriteFile(source, []byte(queryRuntimeTest), 0o600))
	runParamsRuntimeTests(t, dir)
}

const queryRuntimeTest = `package client_test

import (
 "context"
 "io"
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "testing"

 "github.com/stretchr/testify/require"
 genservice "generated.local/gen/param_shapes"
 genclient "generated.local/gen/jsonrpc/param_shapes/client"
 genserver "generated.local/gen/jsonrpc/param_shapes/server"
 goahttp "goa.design/goa/v3/http"
 "goa.design/goa/v3/jsonrpc"
)

func TestQueryInputsRoundTripOutsideParams(t *testing.T) {
 for _,present:=range []bool{false,true}{
  t.Run(map[bool]string{false:"absent",true:"present"}[present],func(t *testing.T){
   input:=&genservice.QueryPayload{Domain:"unchanged",Count:7,Labels:[]string{"one","two"},Pairs:map[string]int{"first":3,"second":5}}
   if present {value:=genservice.Alias("opaque");input.QueryValue=&value}
   var calls int
   endpoints:=&genservice.Endpoints{Query:func(_ context.Context,raw any)(any,error){
    calls++
    require.Equal(t,input,raw.(*genservice.QueryPayload))
    return input.Domain,nil
   }}
   mux:=goahttp.NewMuxer()
   server:=genserver.New(endpoints,mux,goahttp.RequestDecoder,goahttp.ResponseEncoder,nil)
   genserver.Mount(mux,server)
   peer:=httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter,request *http.Request){
    body,err:=io.ReadAll(request.Body);require.NoError(t,err)
    require.NoError(t,request.Body.Close())
    require.NotContains(t,string(body),"credential")
    require.NotContains(t,string(body),"opaque")
    require.NotContains(t,string(body),"labels")
    require.Equal(t,"7",request.URL.Query().Get("count"))
    request.Body=io.NopCloser(strings.NewReader(string(body)))
    mux.ServeHTTP(writer,request)
   }))
   defer peer.Close()
   address,err:=url.Parse(peer.URL);require.NoError(t,err)
   client:=genclient.NewClient(address.Scheme,address.Host,peer.Client(),goahttp.RequestEncoder,goahttp.ResponseDecoder,false)
   result,err:=client.Query()(t.Context(),input)
   require.NoError(t,err)
   require.Equal(t,"unchanged",result)
   require.Equal(t,1,calls)
  })
 }
}

func TestQueryInputsRejectMissingAndMalformedValues(t *testing.T) {
 for _,query:=range []string{"","count=bad","count=7&key=INVALID","count=7&pairs[first]=bad"}{
  request:=httptest.NewRequest(http.MethodPost,"/rpc?"+query,nil)
  _,err:=genserver.DecodeQueryRequest(nil,goahttp.RequestDecoder)(request,&jsonrpc.RawRequest{Params:[]byte("{\"domain\":\"unchanged\"}")})
  require.Error(t,err,query)
 }
}
`
