// Code generated by go-swagger; DO NOT EDIT.

package apps

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

// NewModifyClusterAppParams creates a new ModifyClusterAppParams object
// with the default values initialized.
func NewModifyClusterAppParams() *ModifyClusterAppParams {
	var ()
	return &ModifyClusterAppParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewModifyClusterAppParamsWithTimeout creates a new ModifyClusterAppParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewModifyClusterAppParamsWithTimeout(timeout time.Duration) *ModifyClusterAppParams {
	var ()
	return &ModifyClusterAppParams{

		timeout: timeout,
	}
}

// NewModifyClusterAppParamsWithContext creates a new ModifyClusterAppParams object
// with the default values initialized, and the ability to set a context for a request
func NewModifyClusterAppParamsWithContext(ctx context.Context) *ModifyClusterAppParams {
	var ()
	return &ModifyClusterAppParams{

		Context: ctx,
	}
}

// NewModifyClusterAppParamsWithHTTPClient creates a new ModifyClusterAppParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewModifyClusterAppParamsWithHTTPClient(client *http.Client) *ModifyClusterAppParams {
	var ()
	return &ModifyClusterAppParams{
		HTTPClient: client,
	}
}

/*ModifyClusterAppParams contains all the parameters to send to the API endpoint
for the modify cluster app operation typically these are written to a http.Request
*/
type ModifyClusterAppParams struct {

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
	/*AppName
	  App Name

	*/
	AppName string
	/*Body*/
	Body *models.V4ModifyAppRequest
	/*ClusterID
	  Cluster ID

	*/
	ClusterID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the modify cluster app params
func (o *ModifyClusterAppParams) WithTimeout(timeout time.Duration) *ModifyClusterAppParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the modify cluster app params
func (o *ModifyClusterAppParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the modify cluster app params
func (o *ModifyClusterAppParams) WithContext(ctx context.Context) *ModifyClusterAppParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the modify cluster app params
func (o *ModifyClusterAppParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the modify cluster app params
func (o *ModifyClusterAppParams) WithHTTPClient(client *http.Client) *ModifyClusterAppParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the modify cluster app params
func (o *ModifyClusterAppParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the modify cluster app params
func (o *ModifyClusterAppParams) WithAuthorization(authorization string) *ModifyClusterAppParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the modify cluster app params
func (o *ModifyClusterAppParams) SetAuthorization(authorization string) {
	o.Authorization = authorization
}

// WithXGiantSwarmActivity adds the xGiantSwarmActivity to the modify cluster app params
func (o *ModifyClusterAppParams) WithXGiantSwarmActivity(xGiantSwarmActivity *string) *ModifyClusterAppParams {
	o.SetXGiantSwarmActivity(xGiantSwarmActivity)
	return o
}

// SetXGiantSwarmActivity adds the xGiantSwarmActivity to the modify cluster app params
func (o *ModifyClusterAppParams) SetXGiantSwarmActivity(xGiantSwarmActivity *string) {
	o.XGiantSwarmActivity = xGiantSwarmActivity
}

// WithXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the modify cluster app params
func (o *ModifyClusterAppParams) WithXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) *ModifyClusterAppParams {
	o.SetXGiantSwarmCmdLine(xGiantSwarmCmdLine)
	return o
}

// SetXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the modify cluster app params
func (o *ModifyClusterAppParams) SetXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) {
	o.XGiantSwarmCmdLine = xGiantSwarmCmdLine
}

// WithXRequestID adds the xRequestID to the modify cluster app params
func (o *ModifyClusterAppParams) WithXRequestID(xRequestID *string) *ModifyClusterAppParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the modify cluster app params
func (o *ModifyClusterAppParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithAppName adds the appName to the modify cluster app params
func (o *ModifyClusterAppParams) WithAppName(appName string) *ModifyClusterAppParams {
	o.SetAppName(appName)
	return o
}

// SetAppName adds the appName to the modify cluster app params
func (o *ModifyClusterAppParams) SetAppName(appName string) {
	o.AppName = appName
}

// WithBody adds the body to the modify cluster app params
func (o *ModifyClusterAppParams) WithBody(body *models.V4ModifyAppRequest) *ModifyClusterAppParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the modify cluster app params
func (o *ModifyClusterAppParams) SetBody(body *models.V4ModifyAppRequest) {
	o.Body = body
}

// WithClusterID adds the clusterID to the modify cluster app params
func (o *ModifyClusterAppParams) WithClusterID(clusterID string) *ModifyClusterAppParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the modify cluster app params
func (o *ModifyClusterAppParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WriteToRequest writes these params to a swagger request
func (o *ModifyClusterAppParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param app_name
	if err := r.SetPathParam("app_name", o.AppName); err != nil {
		return err
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
