
# Gateway

Package gateway provides an RPC-style interface to a "service" (struct with methods) via API Gateway for HTTP access.

## About

Why would you go with RPC style for API Gateway? While it's a great tool for avoiding backend server maintenance, API Gateway provides a very convoluted and unintuitive interface for creating APIs. Defining an API is not hard, they really took something simple and made it more difficult.

Many of API Gateway's features are unnecessary unless you're re-mapping a legacy API, so it can be much simpler to (ab)use API Gateway's scaling capabilities while effectively ignoring its other features.

With this package you just define a struct full of methods, and public methods will be exposed via HTTP. This is similar to Dropbox's V2 API and [go-hpc](https://github.com/tj/go-hpc).

## Setup

Create an API Gateway route of `POST /{method}`, pointing to your Lambda function, then use the mapping template below to relay the request. Note that the parameter name of "{method}" is important.

Then create your Lambda function. This package implements the [apex](https://github.com/apex/go-apex).Handler, so an implementation may look something like this:

```go
package main

import (
  "github.com/tj/go-gateway"
  "github.com/apex/go-apex"
)

type Math struct{}

type AddInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

func (m *Math) Add(in *AddInput) (int, error) {
	return in.A + in.B, nil
}

func (m *Math) Sub(in *AddInput) (int, error) {
	return in.A - in.B, nil
}

func main() {
  apex.Handle(gateway.New(&Math{}))
}
```

Deploy the API and you'll be able to invoke `/Add` or `/Sub` with the request body `{ "a": 5, "b": 10 }`. Note that snake-case is also supported, so `/add` or `/sub` work here as well. If you'd like to separate by resource, simply deploy functions to `/pets/{method}`, `/books/{method}` and so on.

## Mapping Template

Use the following mapping template to relay the request information to your Lambda function.

```json
#set($allParams = $input.params())
{
"body" : $input.json('$'),
"params" : {
#foreach($type in $allParams.keySet())
    #set($params = $allParams.get($type))
"$type" : {
    #foreach($paramName in $params.keySet())
    "$paramName" : "$util.escapeJavaScript($params.get($paramName))"
        #if($foreach.hasNext),#end
    #end
}
    #if($foreach.hasNext),#end
#end
},
"context" : {
    "account-id" : "$context.identity.accountId",
    "api-id" : "$context.apiId",
    "api-key" : "$context.identity.apiKey",
    "authorizer-principal-id" : "$context.authorizer.principalId",
    "caller" : "$context.identity.caller",
    "cognito-authentication-provider" : "$context.identity.cognitoAuthenticationProvider",
    "cognito-authentication-type" : "$context.identity.cognitoAuthenticationType",
    "cognito-identity-id" : "$context.identity.cognitoIdentityId",
    "cognito-identity-pool-id" : "$context.identity.cognitoIdentityPoolId",
    "http-method" : "$context.httpMethod",
    "stage" : "$context.stage",
    "source-ip" : "$context.identity.sourceIp",
    "user" : "$context.identity.user",
    "user-agent" : "$context.identity.userAgent",
    "user-arn" : "$context.identity.userArn",
    "request-id" : "$context.requestId",
    "resource-id" : "$context.resourceId",
    "resource-path" : "$context.resourcePath"
    }
}
```

## Reference request

The request received by go-gateway looks something like the following:

```json
{
  "body": {
    "a": 5,
    "b": 5
  },
  "params": {
    "path": {
      "method": "Add"
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
      "Host": "whatever.execute-api.us-west-2.amazonaws.com",
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
    "api-id": "whatever",
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
}
```

## Badges

[![Build Status](https://semaphoreci.com/api/v1/tj/go-gateway/branches/master/badge.svg)](https://semaphoreci.com/tj/go-gateway)
[![GoDoc](https://godoc.org/github.com/tj/go-gateway?status.svg)](https://godoc.org/github.com/tj/go-gateway)
![](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/status-stable-green.svg)
[![](http://apex.sh/images/badge.svg)](https://apex.sh/ping/)

---

> [tjholowaychuk.com](http://tjholowaychuk.com) &nbsp;&middot;&nbsp;
> GitHub [@tj](https://github.com/tj) &nbsp;&middot;&nbsp;
> Twitter [@tjholowaychuk](https://twitter.com/tjholowaychuk)
