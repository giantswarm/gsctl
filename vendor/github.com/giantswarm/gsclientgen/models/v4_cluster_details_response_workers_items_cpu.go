// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// V4ClusterDetailsResponseWorkersItemsCPU v4 cluster details response workers items Cpu
// swagger:model v4ClusterDetailsResponseWorkersItemsCpu
type V4ClusterDetailsResponseWorkersItemsCPU struct {

	// Number of CPU cores
	Cores int64 `json:"cores,omitempty"`
}

// Validate validates this v4 cluster details response workers items Cpu
func (m *V4ClusterDetailsResponseWorkersItemsCPU) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *V4ClusterDetailsResponseWorkersItemsCPU) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *V4ClusterDetailsResponseWorkersItemsCPU) UnmarshalBinary(b []byte) error {
	var res V4ClusterDetailsResponseWorkersItemsCPU
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}