// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// V4ClusterDetailsResponseKvmPortMappingsItems v4 cluster details response kvm port mappings items
// swagger:model v4ClusterDetailsResponseKvmPortMappingsItems
type V4ClusterDetailsResponseKvmPortMappingsItems struct {

	// The port on the host cluster that will forward traffic to the guest cluster
	//
	Port int64 `json:"port,omitempty"`

	// The protocol this port mapping is made for.
	//
	Protocol string `json:"protocol,omitempty"`
}

// Validate validates this v4 cluster details response kvm port mappings items
func (m *V4ClusterDetailsResponseKvmPortMappingsItems) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *V4ClusterDetailsResponseKvmPortMappingsItems) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *V4ClusterDetailsResponseKvmPortMappingsItems) UnmarshalBinary(b []byte) error {
	var res V4ClusterDetailsResponseKvmPortMappingsItems
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}