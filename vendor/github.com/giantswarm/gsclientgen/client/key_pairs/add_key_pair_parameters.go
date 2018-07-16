// Code generated by go-swagger; DO NOT EDIT.

package key_pairs

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

	models "github.com/giantswarm/gsclientgen/models"
)

// NewAddKeyPairParams creates a new AddKeyPairParams object
// with the default values initialized.
func NewAddKeyPairParams() *AddKeyPairParams {
	var ()
	return &AddKeyPairParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewAddKeyPairParamsWithTimeout creates a new AddKeyPairParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewAddKeyPairParamsWithTimeout(timeout time.Duration) *AddKeyPairParams {
	var ()
	return &AddKeyPairParams{

		timeout: timeout,
	}
}

// NewAddKeyPairParamsWithContext creates a new AddKeyPairParams object
// with the default values initialized, and the ability to set a context for a request
func NewAddKeyPairParamsWithContext(ctx context.Context) *AddKeyPairParams {
	var ()
	return &AddKeyPairParams{

		Context: ctx,
	}
}

// NewAddKeyPairParamsWithHTTPClient creates a new AddKeyPairParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewAddKeyPairParamsWithHTTPClient(client *http.Client) *AddKeyPairParams {
	var ()
	return &AddKeyPairParams{
		HTTPClient: client,
	}
}

/*AddKeyPairParams contains all the parameters to send to the API endpoint
for the add key pair operation typically these are written to a http.Request
*/
type AddKeyPairParams struct {

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
	  While the `ttl_hours` attribute is optional and will be set to a default value when omitted, the `description` is mandatory.


	*/
	Body *models.V4AddKeyPairRequest
	/*ClusterID
	  Cluster ID

	*/
	ClusterID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the add key pair params
func (o *AddKeyPairParams) WithTimeout(timeout time.Duration) *AddKeyPairParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the add key pair params
func (o *AddKeyPairParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the add key pair params
func (o *AddKeyPairParams) WithContext(ctx context.Context) *AddKeyPairParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the add key pair params
func (o *AddKeyPairParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the add key pair params
func (o *AddKeyPairParams) WithHTTPClient(client *http.Client) *AddKeyPairParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the add key pair params
func (o *AddKeyPairParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the add key pair params
func (o *AddKeyPairParams) WithAuthorization(authorization string) *AddKeyPairParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the add key pair params
func (o *AddKeyPairParams) SetAuthorization(authorization string) {
	o.Authorization = authorization
}

// WithXGiantSwarmActivity adds the xGiantSwarmActivity to the add key pair params
func (o *AddKeyPairParams) WithXGiantSwarmActivity(xGiantSwarmActivity *string) *AddKeyPairParams {
	o.SetXGiantSwarmActivity(xGiantSwarmActivity)
	return o
}

// SetXGiantSwarmActivity adds the xGiantSwarmActivity to the add key pair params
func (o *AddKeyPairParams) SetXGiantSwarmActivity(xGiantSwarmActivity *string) {
	o.XGiantSwarmActivity = xGiantSwarmActivity
}

// WithXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the add key pair params
func (o *AddKeyPairParams) WithXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) *AddKeyPairParams {
	o.SetXGiantSwarmCmdLine(xGiantSwarmCmdLine)
	return o
}

// SetXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the add key pair params
func (o *AddKeyPairParams) SetXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) {
	o.XGiantSwarmCmdLine = xGiantSwarmCmdLine
}

// WithXRequestID adds the xRequestID to the add key pair params
func (o *AddKeyPairParams) WithXRequestID(xRequestID *string) *AddKeyPairParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the add key pair params
func (o *AddKeyPairParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithBody adds the body to the add key pair params
func (o *AddKeyPairParams) WithBody(body *models.V4AddKeyPairRequest) *AddKeyPairParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the add key pair params
func (o *AddKeyPairParams) SetBody(body *models.V4AddKeyPairRequest) {
	o.Body = body
}

// WithClusterID adds the clusterID to the add key pair params
func (o *AddKeyPairParams) WithClusterID(clusterID string) *AddKeyPairParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the add key pair params
func (o *AddKeyPairParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WriteToRequest writes these params to a swagger request
func (o *AddKeyPairParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
