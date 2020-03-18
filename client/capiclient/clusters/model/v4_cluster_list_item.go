package model

import "github.com/go-openapi/strfmt"

// Unused temporarily, until we migrate away from the GS API
// V4ClusterListItem v4 cluster list item
type V4ClusterListItem struct {
	// Date/time of cluster creation
	CreateDate string `json:"create_date,omitempty"`

	// Date/time when cluster has been deleted
	DeleteDate *strfmt.DateTime `json:"delete_date,omitempty"`

	// Unique cluster identifier
	ID string `json:"id,omitempty"`

	// Cluster name
	Name string `json:"name,omitempty"`

	// Name of the organization owning the cluster
	Owner string `json:"owner,omitempty"`

	// API path of the cluster resource
	Path string `json:"path,omitempty"`

	// The semantic version number of this cluster
	ReleaseVersion string `json:"release_version,omitempty"`
}
