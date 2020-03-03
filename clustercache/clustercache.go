package clustercache

import (
	"fmt"
	"path"
	"sort"
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

// GetID gets the cluster ID for a provided name/ID
// by checking in both the user cache and on the API
func GetID(clusterNameOrID string, clientWrapper *client.Wrapper) (string, error) {
	isInCache := IsInCache(clusterNameOrID)
	if isInCache {
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

	var (
		matchingIDs   []string
		allClusterIDs []string = make([]string, 0, len(response.Payload))
	)
	for _, cluster := range response.Payload {
		allClusterIDs = append(allClusterIDs, cluster.ID)
		if matchesValidation(clusterNameOrID, cluster) {
			matchingIDs = append(matchingIDs, cluster.ID)
		}
	}

	if allClusterIDs != nil {
		CacheIDs(allClusterIDs...)
	}

	if matchingIDs == nil {
		return "", microerror.Mask(errors.ClusterNotFoundError)
	} else if len(matchingIDs) > 1 {
		_, id := handleNameCollision(clusterNameOrID, response.Payload)

		return id, nil
	}

	return matchingIDs[0], nil
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

// CacheIDs adds cluster IDs to a persistent cache,
// which can be used for decreasing timeout in getting
// cluster IDs, for commands that take both cluster names and IDs
func CacheIDs(c ...string) {
	fs := config.FileSystem

	existingC := make(chan []string)
	go func() {
		e, _ := read(fs)
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
		err := write(fs, allClusters...)
		writeC <- err
	}()
	// Ignore error output
	_ = <-writeC
}

// IsInCache checks if a cluster ID is present in the
// persistent cluster cache
func IsInCache(ID string) bool {
	existing, _ := read(config.FileSystem)

	for _, name := range existing {
		if name == ID {
			return true
		}
	}

	return false
}

func read(fs afero.Fs) ([]string, error) {
	filePath := path.Join(config.ConfigDirPath, clusterCacheFileName)
	output, err := afero.ReadFile(fs, filePath)
	if err != nil {
		return []string{}, err
	}

	c := strings.Split(string(output), ",")

	return c, nil
}

func write(fs afero.Fs, c ...string) error {
	filePath := path.Join(config.ConfigDirPath, clusterCacheFileName)
	output := []byte(strings.Join(c, ","))

	err := afero.WriteFile(fs, filePath, output, config.ConfigFilePermission)

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

	sort.Strings(final)

	return final
}
