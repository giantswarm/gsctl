// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// V4AppSpec v4 app spec
// swagger:model v4AppSpec
type V4AppSpec struct {

	// The catalog that this app came from
	Catalog string `json:"catalog,omitempty"`

	// Name of the chart that was used to install this app
	Name string `json:"name,omitempty"`

	// Namespace that this app is installed to
	Namespace string `json:"namespace,omitempty"`

	// Version of the chart that was used to install this app
	Version string `json:"version,omitempty"`
}

// Validate validates this v4 app spec
func (m *V4AppSpec) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *V4AppSpec) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *V4AppSpec) UnmarshalBinary(b []byte) error {
	var res V4AppSpec
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
