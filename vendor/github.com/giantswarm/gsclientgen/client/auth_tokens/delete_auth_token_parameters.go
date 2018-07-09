// Code generated by go-swagger; DO NOT EDIT.

package auth_tokens

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewDeleteAuthTokenParams creates a new DeleteAuthTokenParams object
// with the default values initialized.
func NewDeleteAuthTokenParams() *DeleteAuthTokenParams {
	var ()
	return &DeleteAuthTokenParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteAuthTokenParamsWithTimeout creates a new DeleteAuthTokenParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDeleteAuthTokenParamsWithTimeout(timeout time.Duration) *DeleteAuthTokenParams {
	var ()
	return &DeleteAuthTokenParams{

		timeout: timeout,
	}
}

// NewDeleteAuthTokenParamsWithContext creates a new DeleteAuthTokenParams object
// with the default values initialized, and the ability to set a context for a request
func NewDeleteAuthTokenParamsWithContext(ctx context.Context) *DeleteAuthTokenParams {
	var ()
	return &DeleteAuthTokenParams{

		Context: ctx,
	}
}

// NewDeleteAuthTokenParamsWithHTTPClient creates a new DeleteAuthTokenParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDeleteAuthTokenParamsWithHTTPClient(client *http.Client) *DeleteAuthTokenParams {
	var ()
	return &DeleteAuthTokenParams{
		HTTPClient: client,
	}
}

/*DeleteAuthTokenParams contains all the parameters to send to the API endpoint
for the delete auth token operation typically these are written to a http.Request
*/
type DeleteAuthTokenParams struct {

	/*Authorization
	  giantswarm AUTH_TOKEN_HERE

	*/
	Authorization string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the delete auth token params
func (o *DeleteAuthTokenParams) WithTimeout(timeout time.Duration) *DeleteAuthTokenParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete auth token params
func (o *DeleteAuthTokenParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete auth token params
func (o *DeleteAuthTokenParams) WithContext(ctx context.Context) *DeleteAuthTokenParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete auth token params
func (o *DeleteAuthTokenParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete auth token params
func (o *DeleteAuthTokenParams) WithHTTPClient(client *http.Client) *DeleteAuthTokenParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete auth token params
func (o *DeleteAuthTokenParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the delete auth token params
func (o *DeleteAuthTokenParams) WithAuthorization(authorization string) *DeleteAuthTokenParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the delete auth token params
func (o *DeleteAuthTokenParams) SetAuthorization(authorization string) {
	o.Authorization = authorization
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteAuthTokenParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// header param Authorization
	if err := r.SetHeaderParam("Authorization", o.Authorization); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
