# api-schema

This repo contains types, constructors, and other functions that help in creating and parsing Giant Swarm API responses. It is the source of truth for what our status codes mean, and is useful as a dependency in go programs that wish to talk to or otherwise implement the Giant Swarm API.

### docs
For detailed information and responder as well as receiver implementations, see
http://godoc.org/github.com/giantswarm/api-schema.

### facts
- HTTP status codes not always meet the required business logic
- status codes will be tried to match against behaviour
- error messages conflict with user information
- thus error handling is pain
- one wants to use a consistent schema communicating through network

### schema
```
POST /app/

> Content-Type: application/json
{ ... data ... }

< 200
{"status_code": 2000, "status_text": "ok", "data": { ... }}
{"status_code": 2001, "status_text": "created"}

< 500
{"status_code": 4000, "status_text": "bad request"}
{"status_code": 4004, "status_text": "not found"}
{"status_code": 5000, "status_text": "internal server error"}
```

### summary
- the client makes a request to a well defined URL
- the URL points to a resource or a colection
- there are always two respond codes, 200 and 500
- response code 200 means kind of success
- response code 500 means kind of failure
- request and response bodies are always valid JSON
- there is always a "status_code" field describing internal status using an integer
- there is always a "status_text" field describing internal status using an string
- there is optionally a "data" field describing additionally sent data of any type
- the client defines external user information regarding server responses

### usage
Using the [middleware-server
package](https://github.com/giantswarm/middleware-server) a middleware
implementation could look like the following.
```go
package v1

import (
	"net/http"

	apiSchemaPkg "github.com/giantswarm/api-schema"
	srvPkg "github.com/giantswarm/middleware-server"
)

// Reply with status "ressource created".
func (this *V1) CreateApp(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	return ctx.Response.Json(apiSchemaPkg.StatusRessourceCreated(), http.StatusOK)
}

// Reply with data.
func (this *V1) StatusApp(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	return ctx.Response.Json(apiSchemaPkg.StatusData("status"), http.StatusOK)
}
```
