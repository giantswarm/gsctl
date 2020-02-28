package util

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/microerror"
)

const listClustersActivityName = "list-clusters"

func GetClusterID(clusterNameOrID string, clientWrapper *client.Wrapper) (string, error) {
	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listClustersActivityName

	response, err := clientWrapper.GetClusters(auxParams)
	if err != nil {
		switch {
		case clienterror.IsUnauthorizedError(err):
			return "", microerror.Mask(errors.NotAuthorizedError)

		case clienterror.IsAccessForbiddenError(err):
			return "", microerror.Mask(errors.AccessForbiddenError)

		default:
			return "", microerror.Mask(err)
		}
	}

	var clusterIDs []string
	for _, cluster := range response.Payload {
		if matchesValidation(clusterNameOrID, cluster) {
			clusterIDs = append(clusterIDs, cluster.ID)
		}
	}

	switch {
	case clusterIDs == nil:
		return "", microerror.Mask(errors.ClusterNotFoundError)

	case len(clusterIDs) > 1:
		confirmed, id := handleNameCollision(clusterNameOrID, response.Payload)
		if !confirmed {
			return "", nil
		}

		return id, nil

	default:
		return clusterIDs[0], nil
	}
}

func matchesValidation(nameOrID string, cluster *models.V4ClusterListItem) bool {
	return cluster.DeleteDate == nil && (cluster.ID == nameOrID || cluster.Name == nameOrID)
}

func handleNameCollision(nameOrID string, clusters []*models.V4ClusterListItem) (bool, string) {
	var (
		clusterIDs []string

		table = []string{color.CyanString("ID | ORGANIZATION | NAME")}
	)

	for _, cluster := range clusters {
		if matchesValidation(nameOrID, cluster) {
			clusterIDs = append(clusterIDs, cluster.ID)
			table = append(table, fmt.Sprintf("%5s | %5s | %5s\n", cluster.ID, cluster.Owner, cluster.Name))
		}
	}

	// Print output
	fmt.Println("Multiple clusters found")
	fmt.Printf("\n")
	fmt.Println(columnize.SimpleFormat(table))
	fmt.Printf("\n")

	confirmed, id := confirm.AskStrictOneOf(
		fmt.Sprintf(
			"Found more than one cluster called '%s', please type the ID of the cluster that you would like to delete",
			nameOrID,
		),
		clusterIDs,
	)
	if !confirmed {
		return false, ""
	}
	return true, id
}
