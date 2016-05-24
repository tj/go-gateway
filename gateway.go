// Package gateway provides an RPC-style interface to a "service" (struct with methods)
// via API Gateway for HTTP access.
package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/apex/go-apex"
	"github.com/zhgo/nameconv"
)

// error interface type.
var errType = reflect.TypeOf((*error)(nil)).Elem()

// Responder is an interface allowing you to customize the HTTP response.
type Responder interface {
	Status() int
	Body() interface{}
}

// Context metadata.
type Context struct {
	AccountID                     string `json:"account_id"`
	APIID                         string `json:"api_id"`
	APIKey                        string `json:"api_key"`
	AuthorizerPrincipalID         string `json:"authorizer_principal_id"`
	Caller                        string `json:"caller"`
	CognitoAuthenticationProvider string `json:"cognito_authentication_provider"`
	CognitoAuthenticationType     string `json:"cognito_authentication_type"`
	CognitoIdentityID             string `json:"cognito_identity_id"`
	CognitoIdentityPoolID         string `json:"cognito_identity_pool_id"`
	HTTPMethod                    string `json:"http_method"`
	RequestID                     string `json:"request_id"`
	ResourceID                    string `json:"resource_id"`
	ResourcePath                  string `json:"resource_path"`
	SourceIP                      string `json:"source_ip"`
	Stage                         string `json:"stage"`
	User                          string `json:"user"`
	UserAgent                     string `json:"user_agent"`
	UserArn                       string `json:"user_arn"`
}

// Header fields.
type Header map[string]string

// Request from API Gateway requests.
type Request struct {
	Body   json.RawMessage `json:"body"` // Body of the request
	Params struct {
		Path struct {
			Method string `json:"method"` // Method is the RPC method name
		} `json:"path"`
		Header Header `json:"header"`
	} `json:"params"`
	Context *Context `json:"context"`
}

// Response for API Gateway requests.
type Response struct {
	Status int         `json:"status"`
	Body   interface{} `json:"body"`
}

// Gateway wraps your service to expose its methods.
type Gateway struct {
	*Config
	methods map[string]*reflect.Method
}

// Config for the gateway service.
type Config struct {
	Service interface{} // Service instance
	Verbose bool        // Verbose logging
}

// New returns a new gateway with `service`.
func New(service interface{}) *Gateway {
	return NewConfig(&Config{
		Service: service,
	})
}

// NewConfig returns a new gateway with `config`.
func NewConfig(config *Config) *Gateway {
	g := &Gateway{
		Config:  config,
		methods: make(map[string]*reflect.Method),
	}

	g.init()
	return g
}

// log when Verbose is enabled.
func (g *Gateway) log(s string, v ...interface{}) {
	if g.Verbose {
		log.Printf("gateway: "+s, v...)
	}
}

// init registers the service methods.
func (g *Gateway) init() {
	service := reflect.TypeOf(g.Service)
	for i := 0; i < service.NumMethod(); i++ {
		method := service.Method(i)

		// Method must be exported.
		if method.PkgPath != "" {
			g.log("%q unexported", method.Name)
			continue
		}

		g.methods[method.Name] = &method
	}
}

// Methods returns the method names registered.
func (g *Gateway) Methods() (v []*reflect.Method) {
	for _, m := range g.methods {
		v = append(v, m)
	}
	return
}

// Lookup method by `name`.
func (g *Gateway) Lookup(name string) *reflect.Method {
	cname := nameconv.UnderscoreToCamelcase(name, true)
	return g.methods[cname]
}

// Handle Lambda event.
func (g *Gateway) Handle(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
	var req Request

	if err := json.Unmarshal(event, &req); err != nil {
		return &Response{http.StatusBadRequest, "Malformed Request"}, nil
	}

	// lookup method
	name := req.Params.Path.Method
	method := g.Lookup(name)
	if method == nil {
		return &Response{http.StatusNotFound, "Not Found"}, nil
	}

	mtype := method.Type

	var args = []reflect.Value{reflect.ValueOf(g.Service)}

	// input
	if mtype.NumIn() > 1 {
		in := reflect.New(mtype.In(1).Elem())
		args = append(args, in)
		if err := json.Unmarshal(req.Body, in.Interface()); err != nil {
			return &Response{http.StatusBadRequest, "Malformed Request Body"}, nil
		}
	}

	// invoke
	out := method.Func.Call(args)

	// no output
	if len(out) == 0 {
		return &Response{200, nil}, nil
	}

	// one output: (error)
	if len(out) == 1 {
		err := out[0].Interface().(error)
		if r, ok := err.(Responder); ok {
			return &Response{r.Status(), r.Body()}, nil
		}

		return &Response{http.StatusInternalServerError, "Internal Server Error"}, nil
	}

	// two outputs: (interface{}, error)
	if err, ok := out[1].Interface().(error); ok && err != nil {
		if r, ok := err.(Responder); ok {
			return &Response{r.Status(), r.Body()}, nil
		}

		return &Response{http.StatusInternalServerError, "Internal Server Error"}, nil
	}

	// two outputs: (interface{}, error)
	if r, ok := out[0].Interface().(Responder); ok {
		return &Response{r.Status(), r.Body()}, nil
	}

	return &Response{200, out[0].Interface()}, nil
}
