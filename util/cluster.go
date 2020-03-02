package util

import (
	"fmt"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
)

const (
	listClustersActivityName = "list-clusters"
	clusterCacheFileName     = "clustercache"
)

// GetClusterID gets the cluster ID for a provided name/ID
// by checking in both the user cache and on the API
func GetClusterID(clusterNameOrID string, clientWrapper *client.Wrapper) (string, error) {
	isID := IsInClusterCache(clusterNameOrID)
	if isID {
		return clusterNameOrID, nil
	}

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
		CacheClusterIDs(clusterIDs...)

		confirmed, id := handleNameCollision(clusterNameOrID, response.Payload)
		if !confirmed {
			return "", nil
		}

		return id, nil

	default:
		CacheClusterIDs(clusterIDs[0])

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

	printNameCollisionTable(table)

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

func printNameCollisionTable(table []string) {
	// Print output
	fmt.Println("Multiple clusters found")
	fmt.Printf("\n")
	fmt.Println(columnize.SimpleFormat(table))
	fmt.Printf("\n")
}

func CacheClusterIDs(c ...string) {
	existingC := make(chan []string)
	go func() {
		e, _ := readClusterCache()
		existingC <- e
	}()
	existing := <-existingC

	var (
		totalLen    = len(c) + len(existing)
		allClusters = make([]string, 0, totalLen)
	)

	allClusters = append(allClusters, existing...)
	allClusters = append(allClusters, c...)
	allClusters = removeDuplicates(allClusters)

	writeC := make(chan error)
	go func() {
		err := writeClusterCache(allClusters...)
		writeC <- err
	}()
	// Ignore error output
	_ = <-writeC
}

func IsInClusterCache(ID string) bool {
	existing, _ := readClusterCache()

	for _, name := range existing {
		if name == ID {
			fmt.Printf("%s found in cache\n", ID)

			return true
		}
	}

	return false
}

func readClusterCache() ([]string, error) {
	filePath := path.Join(config.ConfigDirPath, clusterCacheFileName)
	output, err := afero.ReadFile(config.FileSystem, filePath)
	if err != nil {
		return []string{}, err
	}

	c := strings.Split(string(output), ",")

	return c, nil
}

func writeClusterCache(c ...string) error {
	filePath := path.Join(config.ConfigDirPath, clusterCacheFileName)
	output := []byte(strings.Join(c, ","))

	err := afero.WriteFile(config.FileSystem, filePath, output, config.ConfigFilePermission)

	return err
}

func removeDuplicates(c []string) []string {
	uniqueVals := make(map[string]bool)

	for _, cluster := range c {
		uniqueVals[cluster] = true
	}

	final := make([]string, 0, len(uniqueVals))
	for ID := range uniqueVals {
		final = append(final, ID)
	}

	return final
}
