package genjs

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
)

//NewGenerator returns an initialized instance of a JavaScript Client Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API       *design.APIDefinition // The API definition
	OutDir    string                // Destination directory
	Timeout   time.Duration         // Timeout used by JavaScript client when making requests
	Scheme    string                // Scheme used by JavaScript client
	Host      string                // Host addressed by JavaScript client
	NoExample bool                  // Do not generate an HTML example file
	genfiles  []string              // Generated files
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var (
		outDir, ver  string
		timeout      time.Duration
		scheme, host string
		noexample    bool
	)

	set := flag.NewFlagSet("client", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.String("design", "", "")
	set.DurationVar(&timeout, "timeout", time.Duration(20)*time.Second, "")
	set.StringVar(&scheme, "scheme", "", "")
	set.StringVar(&host, "host", "", "")
	set.StringVar(&ver, "version", "", "")
	set.BoolVar(&noexample, "noexample", false, "")
	set.Parse(os.Args[1:])

	// First check compatibility
	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	// Now proceed
	g := &Generator{OutDir: outDir, Timeout: timeout, Scheme: scheme, Host: host, NoExample: noexample, API: design.Design}

	return g.Generate()
}

// Generate produces the skeleton main.
func (g *Generator) Generate() (_ []string, err error) {
	if g.API == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	if g.Timeout == 0 {
		g.Timeout = 20 * time.Second
	}
	if g.Scheme == "" && len(g.API.Schemes) > 0 {
		g.Scheme = g.API.Schemes[0]
	}
	if g.Scheme == "" {
		g.Scheme = "http"
	}
	if g.Host == "" {
		g.Host = g.API.Host
	}
	if g.Host == "" {
		return nil, fmt.Errorf("missing host value, set it with --host")
	}

	g.OutDir = filepath.Join(g.OutDir, "js")
	if err := os.RemoveAll(g.OutDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(g.OutDir, 0755); err != nil {
		return nil, err
	}
	g.genfiles = append(g.genfiles, g.OutDir)

	// Generate client.js
	exampleAction, err := g.generateJS(filepath.Join(g.OutDir, "client.js"))
	if err != nil {
		return
	}

	// Generate axios.html
	if err = g.generateAxiosJS(); err != nil {
		return
	}

	if exampleAction != nil && !g.NoExample {
		// Generate index.html
		if err = g.generateIndexHTML(filepath.Join(g.OutDir, "index.html"), exampleAction); err != nil {
			return
		}

		// Generate example
		if err = g.generateExample(); err != nil {
			return
		}
	}

	return g.genfiles, nil
}

func (g *Generator) generateJS(jsFile string) (_ *design.ActionDefinition, err error) {
	file, err := codegen.SourceFileFor(jsFile)
	if err != nil {
		return
	}
	defer file.Close()
	g.genfiles = append(g.genfiles, jsFile)

	data := map[string]interface{}{
		"API":     g.API,
		"Host":    g.Host,
		"Scheme":  g.Scheme,
		"Timeout": int64(g.Timeout / time.Millisecond),
	}
	if err = file.ExecuteTemplate("module", moduleT, nil, data); err != nil {
		return
	}

	actions := make(map[string][]*design.ActionDefinition)
	g.API.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(action *design.ActionDefinition) error {
			if as, ok := actions[action.Name]; ok {
				actions[action.Name] = append(as, action)
			} else {
				actions[action.Name] = []*design.ActionDefinition{action}
			}
			return nil
		})
	})

	var exampleAction *design.ActionDefinition
	keys := []string{}
	for n := range actions {
		keys = append(keys, n)
	}
	sort.Strings(keys)
	for _, n := range keys {
		for _, a := range actions[n] {
			if exampleAction == nil && a.Routes[0].Verb == "GET" {
				exampleAction = a
			}
			data := map[string]interface{}{"Action": a}
			funcs := template.FuncMap{"params": params}
			if err = file.ExecuteTemplate("jsFuncs", jsFuncsT, funcs, data); err != nil {
				return
			}
		}
	}

	_, err = file.Write([]byte(moduleTend))
	return exampleAction, err
}

func (g *Generator) generateIndexHTML(htmlFile string, exampleAction *design.ActionDefinition) error {
	file, err := codegen.SourceFileFor(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()
	g.genfiles = append(g.genfiles, htmlFile)

	argNames := params(exampleAction)
	var args string
	if len(argNames) > 0 {
		query := exampleAction.QueryParams.Type.ToObject()
		argValues := make([]string, len(argNames))
		for i, n := range argNames {
			ex := query[n].GenerateExample(g.API.RandomGenerator(), nil)
			argValues[i] = fmt.Sprintf("%v", ex)
		}
		args = strings.Join(argValues, ", ")
	}
	examplePath := exampleAction.Routes[0].FullPath()
	pathParams := exampleAction.Routes[0].Params()
	if len(pathParams) > 0 {
		pathVars := exampleAction.AllParams().Type.ToObject()
		pathValues := make([]interface{}, len(pathParams))
		for i, n := range pathParams {
			ex := pathVars[n].GenerateExample(g.API.RandomGenerator(), nil)
			pathValues[i] = ex
		}
		format := design.WildcardRegex.ReplaceAllLiteralString(examplePath, "/%v")
		examplePath = fmt.Sprintf(format, pathValues...)
	}
	if len(argNames) > 0 {
		args = ", " + args
	}
	exampleFunc := fmt.Sprintf(
		`%s%s ("%s"%s)`,
		exampleAction.Name,
		strings.Title(exampleAction.Parent.Name),
		examplePath,
		args,
	)
	data := map[string]interface{}{
		"API":         g.API,
		"ExampleFunc": exampleFunc,
	}

	return file.ExecuteTemplate("exampleHTML", exampleT, nil, data)
}

func (g *Generator) generateAxiosJS() error {
	filePath := filepath.Join(g.OutDir, "axios.min.js")
	if err := ioutil.WriteFile(filePath, []byte(axios), 0644); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, filePath)

	return nil
}

func (g *Generator) generateExample() error {
	controllerFile := filepath.Join(g.OutDir, "example.go")
	file, err := codegen.SourceFileFor(controllerFile)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err == nil {
			err = file.FormatCode()
		}
	}()
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("github.com/dimfeld/httptreemux"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	if err := file.WriteHeader(fmt.Sprintf("%s JavaScript Client Example", g.API.Name), "js", imports); err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, controllerFile)

	data := map[string]interface{}{"ServeDir": g.OutDir}
	return file.ExecuteTemplate("examples", exampleCtrlT, nil, data)
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}

func params(action *design.ActionDefinition) []string {
	if action.QueryParams == nil {
		return nil
	}
	params := make([]string, len(action.QueryParams.Type.ToObject()))
	i := 0
	for n := range action.QueryParams.Type.ToObject() {
		params[i] = n
		i++
	}
	sort.Strings(params)
	return params
}

const moduleT = `// This module exports functions that give access to the {{.API.Name}} API hosted at {{.API.Host}}.
// It uses the axios javascript library for making the actual HTTP requests.
define(['axios'] , function (axios) {
  function merge(obj1, obj2) {
    var obj3 = {};
    for (var attrname in obj1) { obj3[attrname] = obj1[attrname]; }
    for (var attrname in obj2) { obj3[attrname] = obj2[attrname]; }
    return obj3;
  }

  return function (scheme, host, timeout) {
    scheme = scheme || '{{.Scheme}}';
    host = host || '{{.Host}}';
    timeout = timeout || {{.Timeout}};

    // Client is the object returned by this module.
    var client = axios;

    // URL prefix for all API requests.
    var urlPrefix = scheme + '://' + host;
`

const moduleTend = `  return client;
  };
});
`

const jsFuncsT = `{{$params := params .Action}}
  {{$name := printf "%s%s" .Action.Name (title .Action.Parent.Name)}}// {{if .Action.Description}}{{.Action.Description}}{{else}}{{$name}} calls the {{.Action.Name}} action of the {{.Action.Parent.Name}} resource.{{end}}
  // path is the request path, the format is "{{(index .Action.Routes 0).FullPath}}"
  {{if .Action.Payload}}// data contains the action payload (request body)
  {{end}}{{if $params}}// {{join $params ", "}} {{if gt (len $params) 1}}are{{else}}is{{end}} used to build the request query string.
  {{end}}// config is an optional object to be merged into the config built by the function prior to making the request.
  // The content of the config object is described here: https://github.com/mzabriskie/axios#request-api
  // This function returns a promise which raises an error if the HTTP response is a 4xx or 5xx.
  client.{{$name}} = function (path{{if .Action.Payload}}, data{{end}}{{if $params}}, {{join $params ", "}}{{end}}, config) {
    var cfg = {
      timeout: timeout,
      url: urlPrefix + path,
      method: '{{toLower (index .Action.Routes 0).Verb}}',
{{if $params}}      params: {
{{range $index, $param := $params}}{{if $index}},
{{end}}        {{$param}}: {{$param}}{{end}}
      },
{{end}}{{if .Action.Payload}}    data: data,
{{end}}      responseType: 'json'
    };
    if (config) {
      cfg = merge(cfg, config);
    }
    return client(cfg);
  }
`

const exampleT = `<!doctype html>
<html>
  <head>
    <title>goa JavaScript client loader</title>
  </head>
  <body>
    <h1>{{.API.Name}} Client Test</h1>
    <div id="response"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/require.js/2.1.16/require.min.js"></script>
    <script>
      requirejs.config({
        paths: {
          axios: '/js/axios.min',
          client: '/js/client'
        }
      });
      requirejs(['client'], function (client) {
        client().{{.ExampleFunc}}
          .then(function (resp) {
            document.getElementById('response').innerHTML = resp.statusText;
          })
          .catch(function (resp) {
            document.getElementById('response').innerHTML = resp.statusText;
          });
      });
    </script>
  </body>
</html>
`

const exampleCtrlT = `// MountController mounts the JavaScript example controller under "/js".
// This is just an example, not the best way to do this. A better way would be to specify a file
// server using the Files DSL in the design.
// Use --noexample to prevent this file from being generated.
func MountController(service *goa.Service) {
	// Serve static files under js
	service.ServeFiles("/js/*filepath", {{printf "%q" .ServeDir}})
	service.LogInfo("mount", "ctrl", "JS", "action", "ServeFiles", "route", "GET /js/*")
}
`

const axios = `/* axios v0.7.0 | (c) 2015 by Matt Zabriskie */
!function(e,t){"object"==typeof exports&&"object"==typeof module?module.exports=t():"function"==typeof define&&define.amd?define([],t):"object"==typeof exports?exports.axios=t():e.axios=t()}(this,function(){return function(e){function t(r){if(n[r])return n[r].exports;var o=n[r]={exports:{},id:r,loaded:!1};return e[r].call(o.exports,o,o.exports,t),o.loaded=!0,o.exports}var n={};return t.m=e,t.c=n,t.p="",t(0)}([function(e,t,n){e.exports=n(1)},function(e,t,n){"use strict";var r=n(2),o=n(3),i=n(4),s=n(12),u=e.exports=function(e){"string"==typeof e&&(e=o.merge({url:arguments[0]},arguments[1])),e=o.merge({method:"get",headers:{},timeout:r.timeout,transformRequest:r.transformRequest,transformResponse:r.transformResponse},e),e.withCredentials=e.withCredentials||r.withCredentials;var t=[i,void 0],n=Promise.resolve(e);for(u.interceptors.request.forEach(function(e){t.unshift(e.fulfilled,e.rejected)}),u.interceptors.response.forEach(function(e){t.push(e.fulfilled,e.rejected)});t.length;)n=n.then(t.shift(),t.shift());return n};u.defaults=r,u.all=function(e){return Promise.all(e)},u.spread=n(13),u.interceptors={request:new s,response:new s},function(){function e(){o.forEach(arguments,function(e){u[e]=function(t,n){return u(o.merge(n||{},{method:e,url:t}))}})}function t(){o.forEach(arguments,function(e){u[e]=function(t,n,r){return u(o.merge(r||{},{method:e,url:t,data:n}))}})}e("delete","get","head"),t("post","put","patch")}()},function(e,t,n){"use strict";var r=n(3),o=/^\)\]\}',?\n/,i={"Content-Type":"application/x-www-form-urlencoded"};e.exports={transformRequest:[function(e,t){return r.isFormData(e)?e:r.isArrayBuffer(e)?e:r.isArrayBufferView(e)?e.buffer:!r.isObject(e)||r.isFile(e)||r.isBlob(e)?e:(r.isUndefined(t)||(r.forEach(t,function(e,n){"content-type"===n.toLowerCase()&&(t["Content-Type"]=e)}),r.isUndefined(t["Content-Type"])&&(t["Content-Type"]="application/json")),JSON.stringify(e))}],transformResponse:[function(e){if("string"==typeof e){e=e.replace(o,"");try{e=JSON.parse(e)}catch(t){}}return e}],headers:{common:{Accept:"application/json, text/plain, */*"},patch:r.merge(i),post:r.merge(i),put:r.merge(i)},timeout:0,xsrfCookieName:"XSRF-TOKEN",xsrfHeaderName:"X-XSRF-TOKEN"}},function(e,t){"use strict";function n(e){return"[object Array]"===v.call(e)}function r(e){return"[object ArrayBuffer]"===v.call(e)}function o(e){return"[object FormData]"===v.call(e)}function i(e){return"undefined"!=typeof ArrayBuffer&&ArrayBuffer.isView?ArrayBuffer.isView(e):e&&e.buffer&&e.buffer instanceof ArrayBuffer}function s(e){return"string"==typeof e}function u(e){return"number"==typeof e}function a(e){return"undefined"==typeof e}function f(e){return null!==e&&"object"==typeof e}function c(e){return"[object Date]"===v.call(e)}function p(e){return"[object File]"===v.call(e)}function l(e){return"[object Blob]"===v.call(e)}function d(e){return e.replace(/^\s*/,"").replace(/\s*$/,"")}function h(e){return"[object Arguments]"===v.call(e)}function m(){return"undefined"!=typeof window&&"undefined"!=typeof document&&"function"==typeof document.createElement}function y(e,t){if(null!==e&&"undefined"!=typeof e){var r=n(e)||h(e);if("object"==typeof e||r||(e=[e]),r)for(var o=0,i=e.length;i>o;o++)t.call(null,e[o],o,e);else for(var s in e)e.hasOwnProperty(s)&&t.call(null,e[s],s,e)}}function g(){var e={};return y(arguments,function(t){y(t,function(t,n){e[n]=t})}),e}var v=Object.prototype.toString;e.exports={isArray:n,isArrayBuffer:r,isFormData:o,isArrayBufferView:i,isString:s,isNumber:u,isObject:f,isUndefined:a,isDate:c,isFile:p,isBlob:l,isStandardBrowserEnv:m,forEach:y,merge:g,trim:d}},function(e,t,n){(function(t){"use strict";e.exports=function(e){return new Promise(function(r,o){try{"undefined"!=typeof XMLHttpRequest||"undefined"!=typeof ActiveXObject?n(6)(r,o,e):"undefined"!=typeof t&&n(6)(r,o,e)}catch(i){o(i)}})}}).call(t,n(5))},function(e,t){function n(){f=!1,s.length?a=s.concat(a):c=-1,a.length&&r()}function r(){if(!f){var e=setTimeout(n);f=!0;for(var t=a.length;t;){for(s=a,a=[];++c<t;)s&&s[c].run();c=-1,t=a.length}s=null,f=!1,clearTimeout(e)}}function o(e,t){this.fun=e,this.array=t}function i(){}var s,u=e.exports={},a=[],f=!1,c=-1;u.nextTick=function(e){var t=new Array(arguments.length-1);if(arguments.length>1)for(var n=1;n<arguments.length;n++)t[n-1]=arguments[n];a.push(new o(e,t)),1!==a.length||f||setTimeout(r,0)},o.prototype.run=function(){this.fun.apply(null,this.array)},u.title="browser",u.browser=!0,u.env={},u.argv=[],u.version="",u.versions={},u.on=i,u.addListener=i,u.once=i,u.off=i,u.removeListener=i,u.removeAllListeners=i,u.emit=i,u.binding=function(e){throw new Error("process.binding is not supported")},u.cwd=function(){return"/"},u.chdir=function(e){throw new Error("process.chdir is not supported")},u.umask=function(){return 0}},function(e,t,n){"use strict";var r=n(2),o=n(3),i=n(7),s=n(8),u=n(9);e.exports=function(e,t,a){var f=u(a.data,a.headers,a.transformRequest),c=o.merge(r.headers.common,r.headers[a.method]||{},a.headers||{});o.isFormData(f)&&delete c["Content-Type"];var p=new(XMLHttpRequest||ActiveXObject)("Microsoft.XMLHTTP");if(p.open(a.method.toUpperCase(),i(a.url,a.params),!0),p.timeout=a.timeout,p.onreadystatechange=function(){if(p&&4===p.readyState){var n=s(p.getAllResponseHeaders()),r=-1!==["text",""].indexOf(a.responseType||"")?p.responseText:p.response,o={data:u(r,n,a.transformResponse),status:p.status,statusText:p.statusText,headers:n,config:a};(p.status>=200&&p.status<300?e:t)(o),p=null}},o.isStandardBrowserEnv()){var l=n(10),d=n(11),h=d(a.url)?l.read(a.xsrfCookieName||r.xsrfCookieName):void 0;h&&(c[a.xsrfHeaderName||r.xsrfHeaderName]=h)}if(o.forEach(c,function(e,t){f||"content-type"!==t.toLowerCase()?p.setRequestHeader(t,e):delete c[t]}),a.withCredentials&&(p.withCredentials=!0),a.responseType)try{p.responseType=a.responseType}catch(m){if("json"!==p.responseType)throw m}o.isArrayBuffer(f)&&(f=new DataView(f)),p.send(f)}},function(e,t,n){"use strict";function r(e){return encodeURIComponent(e).replace(/%40/gi,"@").replace(/%3A/gi,":").replace(/%24/g,"$").replace(/%2C/gi,",").replace(/%20/g,"+").replace(/%5B/gi,"[").replace(/%5D/gi,"]")}var o=n(3);e.exports=function(e,t){if(!t)return e;var n=[];return o.forEach(t,function(e,t){null!==e&&"undefined"!=typeof e&&(o.isArray(e)&&(t+="[]"),o.isArray(e)||(e=[e]),o.forEach(e,function(e){o.isDate(e)?e=e.toISOString():o.isObject(e)&&(e=JSON.stringify(e)),n.push(r(t)+"="+r(e))}))}),n.length>0&&(e+=(-1===e.indexOf("?")?"?":"&")+n.join("&")),e}},function(e,t,n){"use strict";var r=n(3);e.exports=function(e){var t,n,o,i={};return e?(r.forEach(e.split("\n"),function(e){o=e.indexOf(":"),t=r.trim(e.substr(0,o)).toLowerCase(),n=r.trim(e.substr(o+1)),t&&(i[t]=i[t]?i[t]+", "+n:n)}),i):i}},function(e,t,n){"use strict";var r=n(3);e.exports=function(e,t,n){return r.forEach(n,function(n){e=n(e,t)}),e}},function(e,t,n){"use strict";var r=n(3);e.exports={write:function(e,t,n,o,i,s){var u=[];u.push(e+"="+encodeURIComponent(t)),r.isNumber(n)&&u.push("expires="+new Date(n).toGMTString()),r.isString(o)&&u.push("path="+o),r.isString(i)&&u.push("domain="+i),s===!0&&u.push("secure"),document.cookie=u.join("; ")},read:function(e){var t=document.cookie.match(new RegExp("(^|;\\s*)("+e+")=([^;]*)"));return t?decodeURIComponent(t[3]):null},remove:function(e){this.write(e,"",Date.now()-864e5)}}},function(e,t,n){"use strict";function r(e){var t=e;return s&&(u.setAttribute("href",t),t=u.href),u.setAttribute("href",t),{href:u.href,protocol:u.protocol?u.protocol.replace(/:$/,""):"",host:u.host,search:u.search?u.search.replace(/^\?/,""):"",hash:u.hash?u.hash.replace(/^#/,""):"",hostname:u.hostname,port:u.port,pathname:"/"===u.pathname.charAt(0)?u.pathname:"/"+u.pathname}}var o,i=n(3),s=/(msie|trident)/i.test(navigator.userAgent),u=document.createElement("a");o=r(window.location.href),e.exports=function(e){var t=i.isString(e)?r(e):e;return t.protocol===o.protocol&&t.host===o.host}},function(e,t,n){"use strict";function r(){this.handlers=[]}var o=n(3);r.prototype.use=function(e,t){return this.handlers.push({fulfilled:e,rejected:t}),this.handlers.length-1},r.prototype.eject=function(e){this.handlers[e]&&(this.handlers[e]=null)},r.prototype.forEach=function(e){o.forEach(this.handlers,function(t){null!==t&&e(t)})},e.exports=r},function(e,t){"use strict";e.exports=function(e){return function(t){return e.apply(null,t)}}}])});`
