// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// V4CreateAuthTokenRequest v4 create auth token request
// swagger:model v4CreateAuthTokenRequest
type V4CreateAuthTokenRequest struct {

	// Your email address
	Email string `json:"email,omitempty"`

	// Your password as a base64 encoded string
	PasswordBase64 string `json:"password_base64,omitempty"`
}

// Validate validates this v4 create auth token request
func (m *V4CreateAuthTokenRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *V4CreateAuthTokenRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *V4CreateAuthTokenRequest) UnmarshalBinary(b []byte) error {
	var res V4CreateAuthTokenRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}