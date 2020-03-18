package clusters

import (
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/microerror"
	"github.com/go-openapi/strfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterMaxAge = 1 // days
)

type Client struct {
	formats strfmt.Registry
}

func New(formats strfmt.Registry) *Client {
	client := &Client{
		formats: formats,
	}

	return client
}

// Using response model from gs api until it's deprecated, for compatibility
func (c *Client) GetClusters(clientset *versioned.Clientset) ([]*models.V4ClusterListItem, error) {
	// TODO: Assert if there is no clientset

	clustersV4, err := clientset.CoreV1alpha1().AWSClusterConfigs(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clustersV5, err := clientset.InfrastructureV1alpha2().AWSClusters(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var (
		deleteDate time.Time

		tClusters              = len(clustersV4.Items) + len(clustersV5.Items)
		payload                = make([]*models.V4ClusterListItem, 0, tClusters)
		minAllowedCreationDate = time.Now().UTC().AddDate(0, 0, -clusterMaxAge)
	)

	// V4 Clusters
	for _, cluster := range clustersV4.Items {
		formattedCluster := &models.V4ClusterListItem{
			ID:             cluster.Spec.Guest.ID,
			CreateDate:     cluster.GetCreationTimestamp().UTC().Format(time.RFC3339),
			Name:           cluster.Spec.Guest.Name,
			Owner:          cluster.Spec.Guest.Owner,
			ReleaseVersion: cluster.Spec.Guest.ReleaseVersion,
		}
		formattedCluster.Path = fmt.Sprintf("/v4/clusters/%s/", formattedCluster.ID)
		if cluster.DeletionTimestamp != nil {
			deleteDate = cluster.GetDeletionTimestamp().Time.UTC()
			if deleteDate.After(minAllowedCreationDate) {
				fDeleteDate := strfmt.DateTime(deleteDate)
				formattedCluster.DeleteDate = &fDeleteDate

				payload = append(payload, formattedCluster)
			}
			continue
		}

		payload = append(payload, formattedCluster)
	}

	// V5 Clusters
	for _, cluster := range clustersV5.Items {
		formattedCluster := &models.V4ClusterListItem{
			ID:             cluster.GetName(),
			CreateDate:     cluster.GetCreationTimestamp().UTC().Format(time.RFC3339),
			Name:           cluster.Spec.Cluster.Description,
			Owner:          cluster.Labels["giantswarm.io/organization"],
			ReleaseVersion: cluster.Labels["release.giantswarm.io/version"],
		}
		formattedCluster.Path = fmt.Sprintf("/v5/clusters/%s/", formattedCluster.ID)
		// TODO: Fix deletion date for v5 clusters

		payload = append(payload, formattedCluster)
	}

	return payload, nil
}
