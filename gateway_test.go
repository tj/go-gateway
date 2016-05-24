package gateway

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func event(method, body string) json.RawMessage {
	return json.RawMessage(`{
	  "body": ` + body + `,
	  "params": {
	    "path": {
	      "method": "` + method + `"
	    },
	    "querystring": {},
	    "header": {
	      "Accept": "*/*",
	      "CloudFront-Forwarded-Proto": "https",
	      "CloudFront-Is-Desktop-Viewer": "true",
	      "CloudFront-Is-Mobile-Viewer": "false",
	      "CloudFront-Is-SmartTV-Viewer": "false",
	      "CloudFront-Is-Tablet-Viewer": "false",
	      "CloudFront-Viewer-Country": "CA",
	      "Content-Type": "application/json",
	      "Host": "whxkpa6fwf.execute-api.us-west-2.amazonaws.com",
	      "User-Agent": "curl/7.43.0",
	      "Via": "1.1 fc8d4c3a573bbd496e96047052c4d3f1.cloudfront.net (CloudFront)",
	      "X-Amz-Cf-Id": "RW7zWvoOaoxsxWM_OPEadaqJf_rTQg5Pkfu4SMAruaULcqYH0K9MUA==",
	      "X-Forwarded-For": "70.66.179.182, 54.182.214.52",
	      "X-Forwarded-Port": "443",
	      "X-Forwarded-Proto": "https"
	    }
	  },
	  "context": {
	    "account-id": "",
	    "api-id": "whxkpa6fwf",
	    "api-key": "",
	    "authorizer-principal-id": "",
	    "caller": "",
	    "cognito-authentication-provider": "",
	    "cognito-authentication-type": "",
	    "cognito-identity-id": "",
	    "cognito-identity-pool-id": "",
	    "http-method": "POST",
	    "stage": "prod",
	    "source-ip": "70.66.179.182",
	    "user": "",
	    "user-agent": "curl/7.43.0",
	    "user-arn": "",
	    "request-id": "55066e03-19f7-11e6-8e97-231379f58d27",
	    "resource-id": "cppmxl",
	    "resource-path": "/public/{method}"
	  }
	}`)
}

type Math struct{}

type AddInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

func (m *Math) AddSomeNumbers(in *AddInput) (interface{}, error) {
	return in.A + in.B, nil
}

func (m *Math) Add(in *AddInput) (interface{}, error) {
	return in.A + in.B, nil
}

func (m *Math) Sub(in *AddInput) (int, error) {
	return in.A - in.B, nil
}

func (m *Math) NoInput() (int, error) {
	return 5, nil
}

// func (m *Math) NoInputNoOutput() error {
// 	return nil
// }

func (m *Math) Error(in *AddInput) (int, error) {
	return 0, errors.New("boom")
}

func (m *Math) notExported(a, b int) error {
	return nil
}

func TestNewConfig(t *testing.T) {
	g := NewConfig(&Config{
		Service: &Math{},
		Verbose: true,
	})

	m := g.Methods()
	assert.Len(t, m, 5, "incorrect number of methods")
}

func TestGateway_Lookup(t *testing.T) {
	g := NewConfig(&Config{
		Service: &Math{},
	})

	{
		method := g.Lookup("add_some_numbers")
		assert.NotNil(t, method, "lookup by snake case failed")
		assert.Equal(t, "AddSomeNumbers", method.Name)
	}

	{
		method := g.Lookup("AddSomeNumbers")
		assert.NotNil(t, method, "lookup by snake case failed")
		assert.Equal(t, "AddSomeNumbers", method.Name)
	}

	{
		method := g.Lookup("whoop")
		assert.Nil(t, method, "should be missing")
	}
}

func TestGateway_Handle_noInput(t *testing.T) {
	g := New(&Math{})
	e := event("no_input", `{}`)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, &Response{200, 5}, v)
}

func TestGateway_Handle_lowercaseReturnInterface(t *testing.T) {
	g := New(&Math{})
	e := event("add", `{ "a": 5, "b": 10 }`)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, &Response{200, 15}, v)
}

func TestGateway_Handle_lowercaseReturn(t *testing.T) {
	g := New(&Math{})
	e := event("sub", `{ "a": 10, "b": 5 }`)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, &Response{200, 5}, v)
}

func TestGateway_Handle_notFound(t *testing.T) {
	g := New(&Math{})
	e := event("nothing", `{ "a": 10, "b": 5 }`)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, 404, v.(*Response).Status)
	assert.Equal(t, "Not Found", v.(*Response).Body)
}

func TestGateway_Handle_malformedRequest(t *testing.T) {
	g := New(&Math{})
	e := event("nothing", `{ "a": 10, `)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, v.(*Response).Status)
	assert.Equal(t, "Malformed Request", v.(*Response).Body)
}

func TestGateway_Handle_malformedRequestBody(t *testing.T) {
	g := New(&Math{})
	e := event("add", `5`)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, v.(*Response).Status)
	assert.Equal(t, "Malformed Request Body", v.(*Response).Body)
}

func TestGateway_Handle_errors(t *testing.T) {
	g := New(&Math{})
	e := event("error", `{ "a": 5, "b": 5 }`)
	v, err := g.Handle(e, nil)
	assert.NoError(t, err)
	assert.Equal(t, 500, v.(*Response).Status)
	assert.Equal(t, "Internal Server Error", v.(*Response).Body)
}
