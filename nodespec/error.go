package nodespec

import "github.com/giantswarm/microerror"

// instanceTypeNotFoundErr means the user tries to use an endpoint that is not defined
var instanceTypeNotFoundErr = &microerror.Error{
	Kind: "instanceTypeNotFoundErr",
}

// IsInstanceTypeNotFoundErr asserts instanceTypeNotFoundErr.
func IsInstanceTypeNotFoundErr(err error) bool {
	return microerror.Cause(err) == instanceTypeNotFoundErr
}
