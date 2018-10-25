// Code generated by go-swagger; DO NOT EDIT.

package users

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

// NewModifyUserParams creates a new ModifyUserParams object
// with the default values initialized.
func NewModifyUserParams() *ModifyUserParams {
	var ()
	return &ModifyUserParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewModifyUserParamsWithTimeout creates a new ModifyUserParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewModifyUserParamsWithTimeout(timeout time.Duration) *ModifyUserParams {
	var ()
	return &ModifyUserParams{

		timeout: timeout,
	}
}

// NewModifyUserParamsWithContext creates a new ModifyUserParams object
// with the default values initialized, and the ability to set a context for a request
func NewModifyUserParamsWithContext(ctx context.Context) *ModifyUserParams {
	var ()
	return &ModifyUserParams{

		Context: ctx,
	}
}

// NewModifyUserParamsWithHTTPClient creates a new ModifyUserParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewModifyUserParamsWithHTTPClient(client *http.Client) *ModifyUserParams {
	var ()
	return &ModifyUserParams{
		HTTPClient: client,
	}
}

/*ModifyUserParams contains all the parameters to send to the API endpoint
for the modify user operation typically these are written to a http.Request
*/
type ModifyUserParams struct {

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
	  User account details

	*/
	Body *models.V4ModifyUserRequest
	/*Email
	  The user's email address

	*/
	Email string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the modify user params
func (o *ModifyUserParams) WithTimeout(timeout time.Duration) *ModifyUserParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the modify user params
func (o *ModifyUserParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the modify user params
func (o *ModifyUserParams) WithContext(ctx context.Context) *ModifyUserParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the modify user params
func (o *ModifyUserParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the modify user params
func (o *ModifyUserParams) WithHTTPClient(client *http.Client) *ModifyUserParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the modify user params
func (o *ModifyUserParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAuthorization adds the authorization to the modify user params
func (o *ModifyUserParams) WithAuthorization(authorization string) *ModifyUserParams {
	o.SetAuthorization(authorization)
	return o
}

// SetAuthorization adds the authorization to the modify user params
func (o *ModifyUserParams) SetAuthorization(authorization string) {
	o.Authorization = authorization
}

// WithXGiantSwarmActivity adds the xGiantSwarmActivity to the modify user params
func (o *ModifyUserParams) WithXGiantSwarmActivity(xGiantSwarmActivity *string) *ModifyUserParams {
	o.SetXGiantSwarmActivity(xGiantSwarmActivity)
	return o
}

// SetXGiantSwarmActivity adds the xGiantSwarmActivity to the modify user params
func (o *ModifyUserParams) SetXGiantSwarmActivity(xGiantSwarmActivity *string) {
	o.XGiantSwarmActivity = xGiantSwarmActivity
}

// WithXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the modify user params
func (o *ModifyUserParams) WithXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) *ModifyUserParams {
	o.SetXGiantSwarmCmdLine(xGiantSwarmCmdLine)
	return o
}

// SetXGiantSwarmCmdLine adds the xGiantSwarmCmdLine to the modify user params
func (o *ModifyUserParams) SetXGiantSwarmCmdLine(xGiantSwarmCmdLine *string) {
	o.XGiantSwarmCmdLine = xGiantSwarmCmdLine
}

// WithXRequestID adds the xRequestID to the modify user params
func (o *ModifyUserParams) WithXRequestID(xRequestID *string) *ModifyUserParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the modify user params
func (o *ModifyUserParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithBody adds the body to the modify user params
func (o *ModifyUserParams) WithBody(body *models.V4ModifyUserRequest) *ModifyUserParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the modify user params
func (o *ModifyUserParams) SetBody(body *models.V4ModifyUserRequest) {
	o.Body = body
}

// WithEmail adds the email to the modify user params
func (o *ModifyUserParams) WithEmail(email string) *ModifyUserParams {
	o.SetEmail(email)
	return o
}

// SetEmail adds the email to the modify user params
func (o *ModifyUserParams) SetEmail(email string) {
	o.Email = email
}

// WriteToRequest writes these params to a swagger request
func (o *ModifyUserParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param email
	if err := r.SetPathParam("email", o.Email); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
