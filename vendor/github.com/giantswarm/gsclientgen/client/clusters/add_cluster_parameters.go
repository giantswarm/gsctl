// Code generated by go-swagger; DO NOT EDIT.

package clusters

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/giantswarm/gsclientgen/models"
)

// NewAddClusterParams creates a new AddClusterParams object
// with the default values initialized.
func NewAddClusterParams() *AddClusterParams {
	var ()
	return &AddClusterParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewAddClusterParamsWithTimeout creates a new AddClusterParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewAddClusterParamsWithTimeout(timeout time.Duration) *AddClusterParams {
	var ()
	return &AddClusterParams{

		timeout: timeout,
	}
}

// NewAddClusterParamsWithContext creates a new AddClusterParams object
// with the default values initialized, and the ability to set a context for a request
func NewAddClusterParamsWithContext(ctx context.Context) *AddClusterParams {
	var ()
	return &AddClusterParams{

		Context: ctx,
	}
}

// NewAddClusterParamsWithHTTPClient creates a new AddClusterParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewAddClusterParamsWithHTTPClient(client *http.Client) *AddClusterParams {
	var ()
	return &AddClusterParams{
		HTTPClient: client,
	}
}

/*AddClusterParams contains all the parameters to send to the API endpoint
for the add cluster operation typically these are written to a http.Request
*/
type AddClusterParams struct {

	/*Authorization
	  As described in the [authentication](#section/Authentication) section


	*/
	Authorization string
	/*XGiantSwarmActivity
	  Name of an activity to track, like "list-clusters". This allows to
	analyze several API requests sent in context and gives an idea on
	the purpose.


	*/
	XGiantSwarmActivity *string
	/*XGiantSwarmCmdLine
	  If activity has been issued by a CLI, this header can contain the
	command line


	*/
	XGiantSwarmCmdLine *string
	/*XRequestID
	  A randomly generated key that can be used to track a request throughout
	services of Giant Swarm.


	*/
	XRequestID *string
	/*Body
	  New cluster definition

	*/
	Body *models.V4AddClusterRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the add cluster params
func (o *AddClusterParams) WithTimeout(timeout time.Duration) *AddClusterParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the add cluster params
func (o *AddClusterParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the add cluster params
func (o *AddClusterParams) WithContext(ctx context.Context) *AddClusterParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the add cluster params
func (o *AddClusterParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the add cluster params
func (o *AddClusterParams) WithHTTPClient(client *http.Client) *AddClusterParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the add cluster params
func (o *AddClusterParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the add cluster params
func (o *AddClusterParams) WithAuthorization(authorization string) *AddClusterParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the add cluster params
func (o *AddClusterParams) SetAuthorization(authorization string) {
	o.Authorization = authorization
}

// WithXGiantSwarmActivity adds the xGiantSwarmActivity to the add cluster params
func (o *AddClusterParams) WithXGiantSwarmActivity(xGiantSwarmActivity *string) *AddClusterParams {
	o.SetXGiantSwarmActivity(xGiantSwarmActivity)
	return o
}

// SetXGiantSwarmActivity adds the xGiantSwarmActivity to the add cluster params
func (o *AddClusterParams) SetXGiantSwarmActivity(xGiantSwarmActivity *string) {
	o.XGiantSwarmActivity = xGiantSwarmActivity
}

// WithXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the add cluster params
func (o *AddClusterParams) WithXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) *AddClusterParams {
	o.SetXGiantSwarmCmdLine(xGiantSwarmCmdLine)
	return o
}

// SetXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the add cluster params
func (o *AddClusterParams) SetXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) {
	o.XGiantSwarmCmdLine = xGiantSwarmCmdLine
}

// WithXRequestID adds the xRequestID to the add cluster params
func (o *AddClusterParams) WithXRequestID(xRequestID *string) *AddClusterParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the add cluster params
func (o *AddClusterParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithBody adds the body to the add cluster params
func (o *AddClusterParams) WithBody(body *models.V4AddClusterRequest) *AddClusterParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the add cluster params
func (o *AddClusterParams) SetBody(body *models.V4AddClusterRequest) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *AddClusterParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// header param Authorization
	if err := r.SetHeaderParam("Authorization", o.Authorization); err != nil {
		return err
	}

	if o.XGiantSwarmActivity != nil {

		// header param X-Giant-Swarm-Activity
		if err := r.SetHeaderParam("X-Giant-Swarm-Activity", *o.XGiantSwarmActivity); err != nil {
			return err
		}

	}

	if o.XGiantSwarmCmdLine != nil {

		// header param X-Giant-Swarm-CmdLine
		if err := r.SetHeaderParam("X-Giant-Swarm-CmdLine", *o.XGiantSwarmCmdLine); err != nil {
			return err
		}

	}

	if o.XRequestID != nil {

		// header param X-Request-ID
		if err := r.SetHeaderParam("X-Request-ID", *o.XRequestID); err != nil {
			return err
		}

	}

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
