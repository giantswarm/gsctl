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

// vmSizeNotFoundErr means that API returned a VM size that is not known.
var vmSizeNotFoundErr = &microerror.Error{
	Kind: "vmSizeNotFoundErr",
}

// IsVMSizeNotFoundErr asserts vmSizeNotFoundErr.
func IsVMSizeNotFoundErr(err error) bool {
	return microerror.Cause(err) == vmSizeNotFoundErr
}
