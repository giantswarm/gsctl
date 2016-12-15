[![GoDoc](https://godoc.org/github.com/giantswarm/gsclientgen?status.svg)](https://godoc.org/github.com/giantswarm/gsclientgen)
[![Travis-CI Build Badge](https://api.travis-ci.org/giantswarm/gsclientgen.svg?branch=master)](https://travis-ci.org/giantswarm/gsclientgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/gsclientgen)](https://goreportcard.com/report/github.com/giantswarm/gsclientgen)
[![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

# Giant Swarm Golang Client (generated)

Experimental Go client for the Giant Swarm API, auto-generated based on an OAI/Swagger specification using Swagger Codegen.

Note: This client currently covers only a part of the API. Expect lots of breaking changes within the code. Use at your own risk.

Documentation can be found in the sub folder `docs`.

## Usage

In your Go package, import like this:

```go
import "github.com/giantswarm/gsclientgen"
```

Some usage examples:

```go
client := gsclientgen.NewDefaultApi()
myToken := ""

// get an auth token (aka "Login")
requestBody := gsclientgen.LoginBody{Password: base64EncodedPass}
loginResponse, _, err := client.UserLoginModel("email@example.com", requestBody)
if err != nil {
	log.Fatal(err)
}
if loginResponse.StatusCode == 10000 {
	myToken = loginResponse.Data.Id
	fmt.Printf("Successfully logged in. Token is %s.\n", loginResponse.Data.Id)
}

// list organizations the user is member of
authHeader := "giantswarm " + myToken
orgsResponse, _, err := client.GetUserOrganizations(authHeader)
if err != nil {
	log.Fatal(err)
}
if orgsResponse.StatusCode == 10000 {
	var organizations = orgsResponse.Data
	for _, orgName := orgsResponse.Data {
		fmt.Println(orgName)
	}
}

// log out
logoutResponse, _, err := client.UserLogout(authHeader)
if err != nil {
	log.Fatal(err)
}
if logoutResponse.StatusCode == 10007 {
	myToken = ""
	fmt.Println("Successfully logged out")
}
```

## Development

The source API specification can be found in `api-spec/oai-spec.yaml`. Changes here will recflect in changes of the generated code.

To generate client library code after changes in above file, run:

```nohighlight
make generate
```

