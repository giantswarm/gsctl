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
)

// NewGetClusterAppConfigParams creates a new GetClusterAppConfigParams object
// with the default values initialized.
func NewGetClusterAppConfigParams() *GetClusterAppConfigParams {
	var ()
	return &GetClusterAppConfigParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetClusterAppConfigParamsWithTimeout creates a new GetClusterAppConfigParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetClusterAppConfigParamsWithTimeout(timeout time.Duration) *GetClusterAppConfigParams {
	var ()
	return &GetClusterAppConfigParams{

		timeout: timeout,
	}
}

// NewGetClusterAppConfigParamsWithContext creates a new GetClusterAppConfigParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetClusterAppConfigParamsWithContext(ctx context.Context) *GetClusterAppConfigParams {
	var ()
	return &GetClusterAppConfigParams{

		Context: ctx,
	}
}

// NewGetClusterAppConfigParamsWithHTTPClient creates a new GetClusterAppConfigParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetClusterAppConfigParamsWithHTTPClient(client *http.Client) *GetClusterAppConfigParams {
	var ()
	return &GetClusterAppConfigParams{
		HTTPClient: client,
	}
}

/*GetClusterAppConfigParams contains all the parameters to send to the API endpoint
for the get cluster app config operation typically these are written to a http.Request
*/
type GetClusterAppConfigParams struct {

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
	/*ClusterID
	  Cluster ID

	*/
	ClusterID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get cluster app config params
func (o *GetClusterAppConfigParams) WithTimeout(timeout time.Duration) *GetClusterAppConfigParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get cluster app config params
func (o *GetClusterAppConfigParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get cluster app config params
func (o *GetClusterAppConfigParams) WithContext(ctx context.Context) *GetClusterAppConfigParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get cluster app config params
func (o *GetClusterAppConfigParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get cluster app config params
func (o *GetClusterAppConfigParams) WithHTTPClient(client *http.Client) *GetClusterAppConfigParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get cluster app config params
func (o *GetClusterAppConfigParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the get cluster app config params
func (o *GetClusterAppConfigParams) WithAuthorization(authorization string) *GetClusterAppConfigParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the get cluster app config params
func (o *GetClusterAppConfigParams) SetAuthorization(authorization string) {
	o.Authorization = authorization
}

// WithXGiantSwarmActivity adds the xGiantSwarmActivity to the get cluster app config params
func (o *GetClusterAppConfigParams) WithXGiantSwarmActivity(xGiantSwarmActivity *string) *GetClusterAppConfigParams {
	o.SetXGiantSwarmActivity(xGiantSwarmActivity)
	return o
}

// SetXGiantSwarmActivity adds the xGiantSwarmActivity to the get cluster app config params
func (o *GetClusterAppConfigParams) SetXGiantSwarmActivity(xGiantSwarmActivity *string) {
	o.XGiantSwarmActivity = xGiantSwarmActivity
}

// WithXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the get cluster app config params
func (o *GetClusterAppConfigParams) WithXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) *GetClusterAppConfigParams {
	o.SetXGiantSwarmCmdLine(xGiantSwarmCmdLine)
	return o
}

// SetXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the get cluster app config params
func (o *GetClusterAppConfigParams) SetXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) {
	o.XGiantSwarmCmdLine = xGiantSwarmCmdLine
}

// WithXRequestID adds the xRequestID to the get cluster app config params
func (o *GetClusterAppConfigParams) WithXRequestID(xRequestID *string) *GetClusterAppConfigParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the get cluster app config params
func (o *GetClusterAppConfigParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithAppName adds the appName to the get cluster app config params
func (o *GetClusterAppConfigParams) WithAppName(appName string) *GetClusterAppConfigParams {
	o.SetAppName(appName)
	return o
}

// SetAppName adds the appName to the get cluster app config params
func (o *GetClusterAppConfigParams) SetAppName(appName string) {
	o.AppName = appName
}

// WithClusterID adds the clusterID to the get cluster app config params
func (o *GetClusterAppConfigParams) WithClusterID(clusterID string) *GetClusterAppConfigParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the get cluster app config params
func (o *GetClusterAppConfigParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WriteToRequest writes these params to a swagger request
func (o *GetClusterAppConfigParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
